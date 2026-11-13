package onboarding

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/database"
	"github.com/Ashomondi/ikonex-academy/backend/internal/middleware"
)

// TestParentRegistrationVerification covers the parent QR path end to end:
// provisioning, QR creation, verifier delegation, public submission, approval
// with document deletion, and rejection (AC-BE-A03/04/05/06).
func TestParentRegistrationVerification(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run the onboarding integration test")
	}

	db, err := database.Open(dsn, 5, 5)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	service := NewService(db.DB, database.NewAuditLogger(db.DB), nil, nil, 7)
	schoolID, principalID, membershipID, classID := provisionSchoolForB3(t, db, service)

	suffix := time.Now().UnixNano()
	guardianPhone := fmt.Sprintf("+2547%08d", suffix%100000000)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE phone_number = $1`, guardianPhone)
	})

	// Create a 12-hour parent QR.
	createQR := postJSON(t, service.handleCreateParentQR, map[string]string{"notes": "Opening Day"}, &middleware.Identity{
		UserID: principalID, SchoolID: schoolID, MembershipID: membershipID, Role: "principal",
	})
	if createQR.Code != http.StatusCreated {
		t.Fatalf("create parent QR status = %d, body = %s", createQR.Code, createQR.Body.String())
	}
	var qrResponse struct {
		Data struct {
			QRID    string `json:"qr_id"`
			QRToken string `json:"qr_token"`
		} `json:"data"`
	}
	decodeData(t, createQR, &qrResponse)
	if qrResponse.Data.QRID == "" || qrResponse.Data.QRToken == "" {
		t.Fatalf("missing QR fields: %s", createQR.Body.String())
	}

	// Delegate verification to a staff membership.
	grant := postJSON(t, service.handleVerifierGrant, map[string]string{
		"staff_membership_id": membershipID,
		"parent_qr_id":        qrResponse.Data.QRID,
	}, &middleware.Identity{UserID: principalID, SchoolID: schoolID, MembershipID: membershipID, Role: "principal"})
	if grant.Code != http.StatusCreated {
		t.Fatalf("verifier grant status = %d, body = %s", grant.Code, grant.Body.String())
	}

	// Public submission creates a pending record and no active student.
	submission := postJSON(t, service.handleRegistrationSubmission, registrationRequest{
		QRToken:        qrResponse.Data.QRToken,
		ApplicantName:  "Mary Wanjiku",
		ApplicantPhone: guardianPhone,
		Relationship:   "mother",
		Student: registrationStudent{
			AdmissionNumber: fmt.Sprintf("ADM-%d", suffix),
			FullName:        "Wanjiku Njeri",
			DOB:             "2011-05-14",
			Gender:          "F",
			ClassID:         classID,
		},
		DocumentType:   "national_id",
		DocumentBase64: base64.StdEncoding.EncodeToString([]byte("fake-id-image")),
	}, nil)
	if submission.Code != http.StatusCreated {
		t.Fatalf("registration status = %d, body = %s", submission.Code, submission.Body.String())
	}
	var submissionResponse struct {
		Data struct {
			SubmissionID string `json:"submission_id"`
		} `json:"data"`
	}
	decodeData(t, submission, &submissionResponse)
	submissionID := submissionResponse.Data.SubmissionID

	if orphanStudents := countRows(t, db, schoolID, `SELECT COUNT(*) FROM students WHERE admission_number = $1`, fmt.Sprintf("ADM-%d", suffix)); orphanStudents != 0 {
		t.Fatalf("pending submission created %d students, want 0", orphanStudents)
	}

	// Approve: document bytes deleted, verification record and enrolment created.
	approve := postJSON(t, service.handleVerifyRegistration, map[string]string{
		"action":                 "approve",
		"document_type_verified": "national_id",
		"notes":                  "Verified national ID in person",
	}, &middleware.Identity{UserID: principalID, SchoolID: schoolID, MembershipID: membershipID, Role: "principal"}, "id", submissionID)
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status = %d, body = %s", approve.Code, approve.Body.String())
	}

	var blobDeleted bool
	var deletedAt sql.NullTime
	if err := database.WithTenantTx(context.Background(), db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT document_blob IS NULL, document_deleted_at FROM registration_submissions WHERE id = $1`, submissionID).
			Scan(&blobDeleted, &deletedAt)
	}); err != nil {
		t.Fatalf("inspect approved submission: %v", err)
	}
	if !blobDeleted || !deletedAt.Valid {
		t.Fatalf("document was not deleted on approval (null=%v, deleted_at=%v)", blobDeleted, deletedAt.Valid)
	}
	if verifications := countRows(t, db, schoolID, `SELECT COUNT(*) FROM verification_records WHERE submission_id = $1`, submissionID); verifications != 1 {
		t.Fatalf("verification_records = %d, want 1", verifications)
	}
	if students := countRows(t, db, schoolID, `SELECT COUNT(*) FROM students WHERE admission_number = $1`, fmt.Sprintf("ADM-%d", suffix)); students != 1 {
		t.Fatalf("students = %d, want 1", students)
	}
	if guardians := countRows(t, db, schoolID, `SELECT COUNT(*) FROM student_guardians sg JOIN users u ON u.id = sg.user_id WHERE u.phone_number = $1`, guardianPhone); guardians != 1 {
		t.Fatalf("student_guardians = %d, want 1", guardians)
	}

	// The guardian now has a parent membership; granting a verifier badge to it
	// must be rejected with 422.
	var parentMembershipID string
	if err := database.WithTenantTx(context.Background(), db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRow(
			`SELECT m.id::text FROM memberships m JOIN users u ON u.id = m.user_id
			  WHERE m.school_id = $1 AND u.phone_number = $2 AND m.role = 'parent'`,
			schoolID, guardianPhone).Scan(&parentMembershipID)
	}); err != nil {
		t.Fatalf("load parent membership: %v", err)
	}
	parentGrant := postJSON(t, service.handleVerifierGrant, map[string]string{
		"staff_membership_id": parentMembershipID,
		"parent_qr_id":        qrResponse.Data.QRID,
	}, &middleware.Identity{UserID: principalID, SchoolID: schoolID, MembershipID: membershipID, Role: "principal"})
	if parentGrant.Code != http.StatusUnprocessableEntity {
		t.Fatalf("parent verifier grant status = %d, want 422", parentGrant.Code)
	}

	// Revoking the QR invalidates the delegated verifier badge.
	revoke := postJSON(t, service.handleRevokeParentQR, map[string]string{}, &middleware.Identity{
		UserID: principalID, SchoolID: schoolID, MembershipID: membershipID, Role: "principal",
	}, "id", qrResponse.Data.QRID)
	if revoke.Code != http.StatusOK {
		t.Fatalf("revoke parent QR status = %d, body = %s", revoke.Code, revoke.Body.String())
	}
	activeGrants := countRows(t, db, schoolID,
		`SELECT COUNT(*) FROM membership_badges mb JOIN badges b ON b.id = mb.badge_id
		  WHERE b.school_id = $1 AND b.slug = 'delegated_verifier' AND mb.scope_id = $2
		    AND (mb.valid_until IS NULL OR mb.valid_until > NOW())`,
		qrResponse.Data.QRID)
	if activeGrants != 0 {
		t.Fatalf("verifier badges still active after revocation: %d", activeGrants)
	}
}

func provisionSchoolForB3(t *testing.T, db *database.DB, service *Service) (schoolID, principalID, membershipID, classID string) {
	t.Helper()
	suffix := time.Now().UnixNano()

	createBody := validCreateSchoolRequest()
	createBody.Code = fmt.Sprintf("B3-%d", suffix)
	createBody.PrincipalPhone = fmt.Sprintf("+2547%08d", (suffix+1)%100000000)
	createBody.ContactEmail = fmt.Sprintf("office%d@school.ac.ke", suffix)

	created := postJSON(t, service.handleCreateSchool, createBody, nil)
	if created.Code != http.StatusCreated {
		t.Fatalf("create school status = %d, body = %s", created.Code, created.Body.String())
	}
	var createResponse struct {
		Data struct {
			SchoolID        string `json:"school_id"`
			PrincipalUserID string `json:"principal_user_id"`
		} `json:"data"`
	}
	decodeData(t, created, &createResponse)
	schoolID = createResponse.Data.SchoolID
	principalID = createResponse.Data.PrincipalUserID

	wizard := postJSON(t, service.handleSetupWizard, validSetupWizardRequest(), &middleware.Identity{
		UserID: principalID, SchoolID: schoolID, Role: "principal",
	})
	if wizard.Code != http.StatusOK {
		t.Fatalf("setup wizard status = %d, body = %s", wizard.Code, wizard.Body.String())
	}

	if err := database.WithTenantTx(context.Background(), db.DB, schoolID, func(tx *sql.Tx) error {
		if err := tx.QueryRow(`SELECT id::text FROM memberships WHERE school_id = $1 AND role = 'principal' LIMIT 1`, schoolID).Scan(&membershipID); err != nil {
			return err
		}
		return tx.QueryRow(`SELECT id::text FROM classes WHERE school_id = $1 AND stream_name = 'North' LIMIT 1`, schoolID).Scan(&classID)
	}); err != nil {
		t.Fatalf("load provisioned ids: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM audit_log WHERE school_id = $1`, schoolID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM schools WHERE id = $1`, schoolID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, principalID)
	})
	return schoolID, principalID, membershipID, classID
}

func postJSON(t *testing.T, handler http.HandlerFunc, body any, identity *middleware.Identity, pathValues ...string) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(encoded))
	if identity != nil {
		request = request.WithContext(middleware.WithIdentity(context.Background(), *identity))
	}
	for i := 0; i+1 < len(pathValues); i += 2 {
		request.SetPathValue(pathValues[i], pathValues[i+1])
	}
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return recorder
}

func decodeData(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v (body=%s)", err, recorder.Body.String())
	}
}

func countRows(t *testing.T, db *database.DB, schoolID, query string, args ...any) int {
	t.Helper()
	var count int
	if err := database.WithTenantTx(context.Background(), db.DB, schoolID, func(tx *sql.Tx) error {
		return tx.QueryRow(query, args...).Scan(&count)
	}); err != nil {
		t.Fatalf("count query: %v", err)
	}
	return count
}
