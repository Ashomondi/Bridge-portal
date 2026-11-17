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
 * Integration note: this expects `GET {NEXT_PUBLIC_API_URL}/parent/timetable`
 * to exist on the parent BFF alongside `/parent` and `/parent/results`. The
 * academics service's `/api/v1/timetables` route is not a drop-in substitute
 * because it returns class slots for staff, not guardian-scoped student data.
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

    let payload: unknown;

    try {
      payload = await response.json();
    } catch {
      return null;
    }

    return sanitizeParentTimetableData(payload);
  } catch {
    return null;
  }
}

/**
 * Validates the portal payload and normalizes it into the dashboard contract.
 *
 * Only records with a usable shape are kept. A linked student is always
 * selected, and periods/key dates belonging to a student outside the
 * authorized `students` list are stripped before they reach the client.
 */
export function sanitizeParentTimetableData(
  payload: unknown,
): ParentTimetableData | null {
  if (!isRecord(payload)) {
    return null;
  }

  const students = Array.isArray(payload.students)
    ? payload.students
        .map(normalizeTimetableStudent)
        .filter((student): student is TimetableStudent => student !== null)
    : [];

  if (students.length === 0) {
    return null;
  }

  const authorizedStudentIds = new Set(students.map((student) => student.id));
  const requestedStudentId = optionalString(payload.selectedStudentId);
  const selectedStudentId =
    requestedStudentId && authorizedStudentIds.has(requestedStudentId)
      ? requestedStudentId
      : students[0].id;

  const periods = Array.isArray(payload.periods)
    ? payload.periods
        .map(normalizeTimetablePeriod)
        .filter(
          (period): period is TimetablePeriod =>
            period !== null && authorizedStudentIds.has(period.studentId),
        )
    : [];

  const keyDates = Array.isArray(payload.keyDates)
    ? payload.keyDates
        .map(normalizeKeyDate)
        .filter(
          (date): date is KeyDate =>
            date !== null &&
            (!date.studentId || authorizedStudentIds.has(date.studentId)),
        )
    : [];

  return {
    portal: normalizePortal(payload.portal),
    session: normalizeSession(payload.session),
    students,
    selectedStudentId,
    weekLabel: optionalString(payload.weekLabel),
    termLabel: optionalString(payload.termLabel),
    notice: normalizeAcademicNotice(payload.notice),
    periods,
    keyDates,
  };
}

/** Narrows and normalizes an unknown payload entry to a linked student. */
function normalizeTimetableStudent(value: unknown): TimetableStudent | null {
  if (!isRecord(value)) {
    return null;
  }

  const id = optionalString(value.id);
  const name = optionalString(value.name);

  if (!id || !name) {
    return null;
  }

  return {
    id,
    name,
    initials: optionalString(value.initials) ?? initialsFromName(name),
    stream: optionalString(value.stream),
  };
}

/** Narrows and normalizes an unknown payload entry to a timetable period. */
function normalizeTimetablePeriod(value: unknown): TimetablePeriod | null {
  if (!isRecord(value)) {
    return null;
  }

  const id = optionalString(value.id);
  const studentId = optionalString(value.studentId);
  const date = optionalString(value.date);
  const subject = optionalString(value.subject);
  const label = optionalString(value.label);
  const kind = value.kind;

  if (
    !id ||
    !studentId ||
    !date ||
    !isTimetablePeriodKind(kind) ||
    !subject ||
    !label
  ) {
    return null;
  }

  return {
    id,
    studentId,
    date,
    kind,
    subject,
    teacher: optionalString(value.teacher),
    room: optionalString(value.room),
    start: optionalString(value.start),
    end: optionalString(value.end),
    label,
    duration: optionalString(value.duration),
  };
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

/** Narrows and normalizes an unknown payload entry to a key term date. */
function normalizeKeyDate(value: unknown): KeyDate | null {
  if (!isRecord(value)) {
    return null;
  }

  const id = optionalString(value.id);
  const month = optionalString(value.month);
  const day = optionalString(value.day);
  const title = optionalString(value.title);
  const kind = value.kind;

  if (!id || !isKeyDateKind(kind) || !month || !day || !title) {
    return null;
  }

  return {
    id,
    kind,
    month,
    day,
    title,
    meta: optionalString(value.meta),
    badge: optionalString(value.badge),
    studentId: optionalString(value.studentId),
  };
}

/** Checks a key date kind against the supported values. */
function isKeyDateKind(value: unknown): value is KeyDateKind {
  return value === "standard" || value === "notice" || value === "general";
}

/** Narrows and normalizes the payload's portal identity object. */
function normalizePortal(
  value: unknown,
): NonNullable<ParentTimetableData["portal"]> | null {
  if (!isRecord(value)) {
    return null;
  }

  return { name: optionalString(value.name) };
}

/** Narrows and normalizes the payload's guardian session object. */
function normalizeSession(
  value: unknown,
): NonNullable<ParentTimetableData["session"]> | null {
  if (!isRecord(value)) {
    return null;
  }

  const fullName = optionalString(value.fullName);
  const relationship = optionalString(value.relationship);
  const initials = optionalString(value.initials);

  if (!fullName || !relationship || !initials) {
    return null;
  }

  return { fullName, relationship, initials };
}

/** Narrows and normalizes the payload's academic notice object. */
function normalizeAcademicNotice(value: unknown): AcademicNotice | null {
  if (!isRecord(value)) {
    return null;
  }

  const authority = optionalString(value.authority);
  const note = optionalString(value.note);

  if (!authority && !note) {
    return null;
  }

  return {
    authority: authority ?? "Academic office",
    note: note ?? "",
  };
}

/** Trims a string value and maps blank or non-string values to null. */
function optionalString(value: unknown): string | null {
  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();

  return trimmed.length > 0 ? trimmed : null;
}

/** Derives up to two initials from a student's display name. */
function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "S";
}

/** Checks that an unknown value is a plain object rather than an array. */
function isRecord(value: unknown): value is Record<string, unknown> {
  return (
    Boolean(value) && typeof value === "object" && !Array.isArray(value)
  );
}
