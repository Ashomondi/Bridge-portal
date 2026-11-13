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

export type RosterFilters = {
  query: string;
  guardianFilter: string;
  attendanceFilter: string;
};

export type RosterPage = {
  totalPages: number;
  safePage: number;
  startIndex: number;
  pageRows: RosterStudent[];
};

export const ROSTER_PAGE_SIZE = 5;

export const AVATAR_TONE_CLASSES = [
  "bg-[#E8F5F3] text-brand-700",
  "bg-[#F0EBF9] text-[#6D4FA3]",
  "bg-[#FFF6E5] text-[#9A6A10]",
  "bg-[#EAF2FB] text-[#356C9B]",
] as const;

/** Returns the learner's initials, falling back to the first letters of their name. */
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

/** Formats a count using British English grouping, or an em dash when absent. */
export function formatCount(value: number | null) {
  return value === null || value === undefined
    ? "—"
    : value.toLocaleString("en-GB");
}

/** Formats an attendance percentage, or a "not recorded" label when absent. */
export function formatPercent(value: number | null) {
  return value === null || value === undefined ? "Not recorded" : `${value}%`;
}

/** Returns the learner's primary guardian, falling back to the first linked guardian. */
export function getPrimaryGuardian(student: RosterStudent) {
  return (
    student.guardians.find((guardian) => guardian.isPrimary) ??
    student.guardians[0] ??
    null
  );
}

/** Filters learners by search text, guardian status and attendance band. */
export function filterRosterStudents(
  students: RosterStudent[],
  { query, guardianFilter, attendanceFilter }: RosterFilters,
) {
  const normalizedQuery = query.trim().toLowerCase();

  return students.filter((student) => {
    const matchesQuery =
      normalizedQuery.length === 0 ||
      [
        student.name,
        student.admissionNumber,
        student.house ?? "",
        student.gradeSection ?? "",
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery);
    const matchesGuardian =
      guardianFilter === "all" || student.guardianStatus === guardianFilter;
    const percentage = student.attendance.overallPercent;
    const matchesAttendance =
      attendanceFilter === "all" ||
      (attendanceFilter === "above" &&
        percentage !== null &&
        percentage >= 90) ||
      (attendanceFilter === "below" &&
        percentage !== null &&
        percentage < 90) ||
      (attendanceFilter === "not-recorded" && percentage === null);

    return matchesQuery && matchesGuardian && matchesAttendance;
  });
}

/** Resolves a requested page to a bounded page window and its visible rows. */
export function paginateRoster(
  students: RosterStudent[],
  page: number,
  pageSize = ROSTER_PAGE_SIZE,
): RosterPage {
  const totalPages = Math.max(1, Math.ceil(students.length / pageSize));
  const safePage = Math.min(Math.max(page, 1), totalPages);
  const startIndex = (safePage - 1) * pageSize;

  return {
    totalPages,
    safePage,
    startIndex,
    pageRows: students.slice(startIndex, startIndex + pageSize),
  };
}

/** Keeps learner details aligned to a row visible on the active page. */
export function resolveVisibleStudent(
  pageRows: RosterStudent[],
  selectedId: string,
) {
  return (
    pageRows.find((student) => student.id === selectedId) ??
    pageRows[0] ??
    null
  );
}

/** Formats a count with the correct singular or plural noun. */
export function formatCountLabel(
  count: number,
  singular: string,
  plural = `${singular}s`,
) {
  return `${count.toLocaleString("en-GB")} ${count === 1 ? singular : plural}`;
}
