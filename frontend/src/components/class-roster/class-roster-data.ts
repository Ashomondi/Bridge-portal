export type GuardianStatus = "verified" | "pending";

export type Guardian = {
  id: string;
  relationship: string;
  phoneMasked: string;
  status: GuardianStatus;
  verificationNote: string;
  isPrimary: boolean;
};

export type RosterAttendance = {
  overallPercent: number | null;
  totalDays: number | null;
  daysPresent: number | null;
  daysAbsent: number | null;
  absenceNoteStatus: string | null;
};

export type RosterStudent = {
  id: string;
  admissionNumber: string;
  name: string;
  initials?: string;
  house: string | null;
  gradeSection: string | null;
  linkedGuardians: number;
  guardianStatus: GuardianStatus;
  guardians: Guardian[];
  attendance: RosterAttendance;
  enrollmentType: string | null;
  avatarTone?: number;
};

export type ClassRosterData = {
  portal: {
    name: string;
    academicTerm: string;
  } | null;
  class: {
    id: string;
    name: string;
    gradeSection: string;
    classTeacherName: string | null;
    enrolledCount: number;
    cycleLabel: string | null;
  } | null;
  session: {
    userFullName: string;
    userRole: string;
    userSubject: string | null;
    initials: string;
  } | null;
  students: RosterStudent[];
  totalCount: number;
};

export const ROSTER_PAGE_SIZE = 5;

export const AVATAR_TONE_CLASSES = [
  "bg-[#E8F5F3] text-brand-700",
  "bg-[#F0EBF9] text-[#6D4FA3]",
  "bg-[#FFF6E5] text-[#9A6A10]",
  "bg-[#EAF2FB] text-[#356C9B]",
] as const;

export function getStudentInitials(student: RosterStudent) {
  if (student.initials?.trim()) {
    return student.initials.trim();
  }

  const parts = student.name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "—";
}

export function formatCount(value: number | null) {
  return value === null || value === undefined
    ? "—"
    : value.toLocaleString("en-GB");
}

export function formatPercent(value: number | null) {
  return value === null || value === undefined ? "Not recorded" : `${value}%`;
}

export function getPrimaryGuardian(student: RosterStudent) {
  return (
    student.guardians.find((guardian) => guardian.isPrimary) ??
    student.guardians[0] ??
    null
  );
}
