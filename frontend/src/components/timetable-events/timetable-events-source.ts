import { cookies } from "next/headers";
import type {
  AcademicNotice,
  KeyDate,
  KeyDateKind,
  ParentTimetableData,
  TimetablePeriod,
  TimetablePeriodKind,
  TimetableStudent,
} from "./timetable-events-data";

/**
 * Loads the authenticated guardian's timetable and key term dates.
 *
 * The guardian session cookie is forwarded to the parent timetable endpoint,
 * which resolves the guardian's linked students and returns each student's
 * published periods plus the school's notice, week/term labels and key dates.
 *
 * Returns `null` when the portal is not configured, the session is not
 * authorized, or the response cannot be used. The dashboard then renders
 * honest empty states instead of sample data.
 */
export async function loadParentTimetable(): Promise<ParentTimetableData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  const url = `${apiUrl.replace(/\/$/, "")}/parent/timetable`;
  const cookieHeader = cookies().toString();

  try {
    const response = await fetch(url, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
    });

    if (!response.ok) {
      return null;
    }

    const payload = (await response.json()) as Partial<ParentTimetableData>;

    return sanitizeParentTimetableData(payload);
  } catch {
    return null;
  }
}

/**
 * Validates the portal payload and normalizes it into the dashboard contract.
 *
 * Only records with a usable shape are kept. A linked student is always
 * selected, so the page never shows one student's periods under another
 * student's name.
 */
function sanitizeParentTimetableData(
  payload: Partial<ParentTimetableData>,
): ParentTimetableData | null {
  const students = Array.isArray(payload.students)
    ? payload.students.filter(isTimetableStudent)
    : [];

  if (students.length === 0) {
    return null;
  }

  const selectedStudentId =
    students.find((student) => student.id === payload.selectedStudentId)?.id ??
    students[0].id;

  return {
    portal: isPortal(payload.portal) ? payload.portal : null,
    session: isSession(payload.session) ? payload.session : null,
    students,
    selectedStudentId,
    weekLabel: typeof payload.weekLabel === "string" ? payload.weekLabel : null,
    termLabel: typeof payload.termLabel === "string" ? payload.termLabel : null,
    notice: isAcademicNotice(payload.notice) ? payload.notice : null,
    periods: Array.isArray(payload.periods)
      ? payload.periods.filter(isTimetablePeriod)
      : [],
    keyDates: Array.isArray(payload.keyDates)
      ? payload.keyDates.filter(isKeyDate)
      : [],
  };
}

/** Narrows an unknown payload entry to a usable linked student. */
function isTimetableStudent(value: unknown): value is TimetableStudent {
  if (!isRecord(value)) {
    return false;
  }

  return (
    typeof value.id === "string" &&
    value.id.length > 0 &&
    typeof value.name === "string" &&
    typeof value.initials === "string"
  );
}

/** Narrows an unknown payload entry to a timetable period. */
function isTimetablePeriod(value: unknown): value is TimetablePeriod {
  if (!isRecord(value)) {
    return false;
  }

  const kind = value.kind;

  return (
    typeof value.id === "string" &&
    value.id.length > 0 &&
    typeof value.studentId === "string" &&
    value.studentId.length > 0 &&
    typeof value.date === "string" &&
    value.date.length > 0 &&
    isTimetablePeriodKind(kind) &&
    typeof value.subject === "string" &&
    typeof value.label === "string"
  );
}

/** Checks a timetable period kind against the supported values. */
function isTimetablePeriodKind(value: unknown): value is TimetablePeriodKind {
  return (
    value === "completed" ||
    value === "break" ||
    value === "active" ||
    value === "upcoming"
  );
}

/** Narrows an unknown payload entry to a key term date. */
function isKeyDate(value: unknown): value is KeyDate {
  if (!isRecord(value)) {
    return false;
  }

  const kind = value.kind;

  return (
    typeof value.id === "string" &&
    value.id.length > 0 &&
    isKeyDateKind(kind) &&
    typeof value.month === "string" &&
    typeof value.day === "string" &&
    typeof value.title === "string"
  );
}

/** Checks a key date kind against the supported values. */
function isKeyDateKind(value: unknown): value is KeyDateKind {
  return value === "standard" || value === "notice" || value === "general";
}

/** Narrows the payload's portal identity object. */
function isPortal(
  value: unknown,
): value is NonNullable<ParentTimetableData["portal"]> {
  if (!isRecord(value)) {
    return false;
  }

  return value.name === null || typeof value.name === "string";
}

/** Narrows the payload's guardian session object. */
function isSession(
  value: unknown,
): value is NonNullable<ParentTimetableData["session"]> {
  if (!isRecord(value)) {
    return false;
  }

  return (
    typeof value.fullName === "string" &&
    typeof value.relationship === "string" &&
    typeof value.initials === "string"
  );
}

/** Narrows the payload's academic notice object. */
function isAcademicNotice(value: unknown): value is AcademicNotice {
  if (!isRecord(value)) {
    return false;
  }

  return (
    typeof value.authority === "string" && typeof value.note === "string"
  );
}

/** Checks that an unknown value is a plain object. */
function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object";
}
