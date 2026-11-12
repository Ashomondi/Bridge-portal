package onboarding

import "testing"

func TestValidateNewStudent(t *testing.T) {
	valid := registrationStudent{
		AdmissionNumber: "ADM-1001",
		FullName:        "Wanjiku Njeri",
		DOB:             "2011-05-14",
		Gender:          "F",
		ClassID:         "123e4567-e89b-12d3-a456-426614174000",
	}
	if err := validateNewStudent(valid); err != nil {
		t.Fatalf("valid student rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*registrationStudent)
	}{
		{"missing admission", func(s *registrationStudent) { s.AdmissionNumber = "" }},
		{"missing name", func(s *registrationStudent) { s.FullName = "" }},
		{"bad gender", func(s *registrationStudent) { s.Gender = "X" }},
		{"bad dob", func(s *registrationStudent) { s.DOB = "14/05/2011" }},
		{"missing class", func(s *registrationStudent) { s.ClassID = "" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			student := valid
			tc.mutate(&student)
			if err := validateNewStudent(student); err == nil {
				t.Fatalf("invalid student accepted")
			}
		})
	}
}

func TestValidateAdmitRequest(t *testing.T) {
	valid := admitStudentRequest{
		AdmissionNumber: "ADM-2002",
		FullName:        "Brian Otieno",
		DOB:             "2010-09-01",
		Gender:          "M",
		ClassID:         "123e4567-e89b-12d3-a456-426614174000",
		GuardianName:    "Mary Otieno",
		GuardianPhone:   "+254712345678",
	}
	if err := validateAdmitRequest(valid); err != nil {
		t.Fatalf("valid admit request rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*admitStudentRequest)
	}{
		{"missing admission", func(r *admitStudentRequest) { r.AdmissionNumber = "" }},
		{"bad gender", func(r *admitStudentRequest) { r.Gender = "female" }},
		{"bad class uuid", func(r *admitStudentRequest) { r.ClassID = "class-1" }},
		{"bad dorm uuid", func(r *admitStudentRequest) { r.DormitoryID = "dorm-1" }},
		{"missing guardian", func(r *admitStudentRequest) { r.GuardianName = "" }},
		{"bad guardian phone", func(r *admitStudentRequest) { r.GuardianPhone = "0712345678" }},
		{"bad dob", func(r *admitStudentRequest) { r.DOB = "not-a-date" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := valid
			tc.mutate(&request)
			if err := validateAdmitRequest(request); err == nil {
				t.Fatalf("invalid admit request accepted")
			}
		})
	}
}

func TestUUIDAndOTPPatterns(t *testing.T) {
	if !uuidPattern.MatchString("123e4567-e89b-12d3-a456-426614174000") {
		t.Fatalf("valid UUID rejected")
	}
	if uuidPattern.MatchString("not-a-uuid") {
		t.Fatalf("invalid UUID accepted")
	}
	if !otpPattern.MatchString("012345") {
		t.Fatalf("valid OTP rejected")
	}
	if otpPattern.MatchString("12345") || otpPattern.MatchString("abcdef") {
		t.Fatalf("invalid OTP accepted")
	}
}

func TestMapQRStatusError(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{ErrQRNotFound, "QR_INVALID"},
		{ErrQRRevoked, "QR_REVOKED"},
		{ErrQRExpired, "QR_EXPIRED"},
		{ErrQRNotStarted, "QR_NOT_ACTIVE"},
		{ErrParentQRRevoked, "QR_REVOKED"},
		{ErrParentQRExpired, "QR_EXPIRED"},
		{ErrParentQRNotYet, "QR_NOT_ACTIVE"},
	}
	for _, tc := range tests {
		mapped, ok := mapQRStatusError(tc.err)
		if !ok || mapped.code != tc.code {
			t.Fatalf("mapQRStatusError(%v) = %+v, %v; want %s", tc.err, mapped, ok, tc.code)
		}
	}
	if _, ok := mapQRStatusError(nil); ok {
		t.Fatalf("nil error should not map to a QR status")
	}
}
