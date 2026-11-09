package onboarding

import (
	"testing"
)

func validCreateSchoolRequest() createSchoolRequest {
	return createSchoolRequest{
		Code:            "IKX-001",
		Name:            "Alliance High",
		County:          "Kiambu",
		PrincipalName:   "Jane Doe",
		PrincipalPhone:  "+254712345678",
		ContactPhone:    "+254712345679",
		ContactEmail:    "office@alliance.ac.ke",
		OfficeHours:     "Mon-Fri 08:00-17:00",
		PhysicalAddress: "P.O. Box 1",
	}
}

func TestValidateCreateSchool(t *testing.T) {
	if err := validateCreateSchool(validCreateSchoolRequest()); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*createSchoolRequest)
	}{
		{"missing code", func(r *createSchoolRequest) { r.Code = "" }},
		{"missing name", func(r *createSchoolRequest) { r.Name = "" }},
		{"missing principal", func(r *createSchoolRequest) { r.PrincipalName = "" }},
		{"bad principal phone", func(r *createSchoolRequest) { r.PrincipalPhone = "0712345678" }},
		{"bad contact phone", func(r *createSchoolRequest) { r.ContactPhone = "+2547" }},
		{"bad email", func(r *createSchoolRequest) { r.ContactEmail = "not-an-email" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := validCreateSchoolRequest()
			tc.mutate(&req)
			if err := validateCreateSchool(req); err == nil {
				t.Fatalf("invalid request accepted")
			}
		})
	}
}

func validSetupWizardRequest() setupWizardRequest {
	return setupWizardRequest{
		Contact: contactDetails{
			Phone: "+254712345678",
			Email: "office@school.ac.ke",
		},
		AcademicYear: 2026,
		Terms: []termRequest{
			{TermNumber: 1, StartDate: "2026-01-05", EndDate: "2026-04-03"},
			{TermNumber: 2, StartDate: "2026-05-04", EndDate: "2026-08-07"},
			{TermNumber: 3, StartDate: "2026-09-07", EndDate: "2026-11-27"},
		},
		GradeLevels: []int{7, 8, 9, 10},
		Classes:     []classRequest{{GradeLevel: 7, StreamName: "North"}},
		Dormitories: []dormitoryRequest{{Name: "Elgon", Gender: "male", Capacity: 60}},
		GradingScale: gradingScaleRequest{
			EEMin: 80, MEMin: 65, AEMin: 50, BEMin: 0,
		},
	}
}

func TestValidateSetupWizard(t *testing.T) {
	parsed, err := validateSetupWizard(validSetupWizardRequest())
	if err != nil {
		t.Fatalf("valid wizard rejected: %v", err)
	}
	if parsed.academicYear != 2026 || len(parsed.terms) != 3 || len(parsed.grades) != 4 {
		t.Fatalf("unexpected parsed setup: %+v", parsed)
	}

	tests := []struct {
		name   string
		mutate func(*setupWizardRequest)
	}{
		{"two terms", func(r *setupWizardRequest) { r.Terms = r.Terms[:2] }},
		{"duplicate term", func(r *setupWizardRequest) { r.Terms[2].TermNumber = 1 }},
		{"term end before start", func(r *setupWizardRequest) { r.Terms[0].EndDate = "2025-12-01" }},
		{"grade out of range", func(r *setupWizardRequest) { r.GradeLevels = []int{6, 7} }},
		{"duplicate grade", func(r *setupWizardRequest) { r.GradeLevels = []int{7, 7} }},
		{"class grade missing", func(r *setupWizardRequest) { r.Classes = []classRequest{{GradeLevel: 11, StreamName: "North"}} }},
		{"empty stream", func(r *setupWizardRequest) { r.Classes = []classRequest{{GradeLevel: 7, StreamName: " "}} }},
		{"bad dorm gender", func(r *setupWizardRequest) { r.Dormitories[0].Gender = "mixed" }},
		{"scale ordering", func(r *setupWizardRequest) {
			r.GradingScale = gradingScaleRequest{EEMin: 50, MEMin: 65, AEMin: 40, BEMin: 0}
		}},
		{"scale out of range", func(r *setupWizardRequest) { r.GradingScale.EEMin = 120 }},
		{"bad contact phone", func(r *setupWizardRequest) { r.Contact.Phone = "12345" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := validSetupWizardRequest()
			tc.mutate(&req)
			if _, err := validateSetupWizard(req); err == nil {
				t.Fatalf("invalid wizard accepted")
			}
		})
	}
}
