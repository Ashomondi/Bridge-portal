export type GuardianStatus = "verified" | "pending";

export type RosterStudent = {
  id: string;
  admissionNumber: string;
  name: string;
  initials: string;
  house: string;
  linkedGuardians: string;
  guardianStatus: GuardianStatus;
  guardianRelationship: string;
  guardianPhone: string;
  guardianVerificationNote: string;
  gradeSection: string;
  enrollmentType: string;
  overallAttendance: string;
  totalDays: string;
  daysPresent: string;
  daysAbsent: string;
  absenceNoteStatus: string;
  avatarTone: number;
};

const AVATAR_TONES = [
  "bg-[#E8F5F3] text-brand-700",
  "bg-[#F0EBF9] text-[#6D4FA3]",
  "bg-[#FFF6E5] text-[#9A6A10]",
  "bg-[#EAF2FB] text-[#356C9B]",
];

/**
 * Structural roster rows. Replace this collection with the authenticated
 * class roster response: admission number, learner name, house, guardian
 * summary, attendance totals and the selected learner's detail payload.
 */
export const ROSTER_STUDENTS: RosterStudent[] = Array.from(
  { length: 12 },
  (_, index) => ({
    id: `roster-row-${index + 1}`,
    admissionNumber: "[ID]",
    name: "[Name]",
    initials: "[AB]",
    house: "[Dormitory / House]",
    linkedGuardians: "[Count]",
    guardianStatus: index % 3 === 1 ? "pending" : "verified",
    guardianRelationship: "[Relationship]",
    guardianPhone: "[Masked Phone]",
    guardianVerificationNote:
      index % 3 === 1
        ? "Pending verification · [Timestamp]"
        : "Verified · [Timestamp]",
    gradeSection: "[Grade / Section]",
    enrollmentType: "[Enrollment Type]",
    overallAttendance: "[Overall %]",
    totalDays: "[Count]",
    daysPresent: "[Count]",
    daysAbsent: "[Count]",
    absenceNoteStatus: "[Status]",
    avatarTone: index % AVATAR_TONES.length,
  }),
);

export const AVATAR_TONE_CLASSES = AVATAR_TONES;

export const CLASS_PORTAL = {
  portalName: "[Portal Name]",
  academicTerm: "[Academic Term / Year]",
  className: "[Class Name / Grade]",
  classTeacher: "[Name]",
  enrolledCount: "[Count]",
  cycle: "[Cycle / Term]",
  totalRecords: "[Count]",
  userFullName: "[Full Name]",
  userRole: "[Role Title]",
  userSubject: "[Subject Assignment]",
};

export const ROSTER_PAGE_SIZE = 5;
