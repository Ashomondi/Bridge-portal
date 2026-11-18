import { cookies } from "next/headers";
import type {
  RollCallData,
  RollCallError,
  RollCallStudent,
} from "./roll-call-data";

/**
 * The route remains a preview until attendance submission is connected. Flip
 * this off once the save endpoint is wired so the preview notice disappears.
 */
export const ROLL_CALL_IS_PREVIEW = true;

export type RollCallSelection = {
  /** Class whose roster should be loaded; resolved from the session. */
  classId?: string;
};

export type RollCallLoadResult = {
  data: RollCallData | null;
  error: RollCallError | null;
};

type DefaultClassLookup = {
  classId: string | null;
  streamName: string | null;
  error: RollCallError | null;
};

type RosterStudentPayload = {
  student_id?: unknown;
  full_name?: unknown;
  admission_number?: unknown;
};

/**
 * Loads the class roster and current roll-call context for a teacher.
 *
 * When the URL has no `classId`, the loader asks the academics service for the
 * teacher's class-teacher assignment and uses that class as the default. Roster
 * responses are read from the standard `{ status, data }` envelope and mapped
 * into `RollCallData`; HTTP failures are surfaced to the dashboard instead of
 * being disguised as an empty roster.
 */
export async function loadRollCall(
  selection: RollCallSelection = {},
): Promise<RollCallLoadResult> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return { data: null, error: null };
  }

  const baseUrl = apiUrl.replace(/\/$/, "");
  const cookieHeader = cookies().toString();
  let classId = selection.classId?.trim() || null;
  let streamName: string | null = null;

  if (!classId) {
    const lookup = await loadDefaultClass(baseUrl, cookieHeader);

    if (lookup.error) {
      return { data: null, error: lookup.error };
    }

    classId = lookup.classId;
    streamName = lookup.streamName;
  }

  if (!classId) {
    return { data: null, error: "class-required" };
  }

  const url = `${baseUrl}/classes/${encodeURIComponent(classId)}/roster`;

  try {
    const response = await fetch(url, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
    });

    if (!response.ok) {
      return { data: null, error: errorForStatus(response.status) };
    }

    let payload: unknown;

    try {
      payload = await response.json();
    } catch {
      return { data: null, error: "invalid-response" };
    }

    const data = unwrapData(payload);

    if (!isRecord(data)) {
      return { data: null, error: "invalid-response" };
    }

    const resolvedClassId = optionalString(data.class_id) ?? classId;

    return {
      data: {
        portal: normalizePortal(data.portal ?? getRecord(payload)?.portal),
        session: normalizeSession(data.session ?? getRecord(payload)?.session),
        context: {
          classId: resolvedClassId,
          streamName:
            optionalString(data.stream_name) ??
            optionalString(data.class_name) ??
            streamName,
          dateLabel: normalizeDateLabel(data.date ?? data.date_label),
          termLabel:
            optionalString(data.term_label) ?? optionalString(data.term),
          registryRef: optionalString(data.registry_ref),
        },
        students: mapRosterStudents(data.students),
      },
      error: null,
    };
  } catch {
    return { data: null, error: "network" };
  }
}

/** Resolves the class-teacher assignment when the URL has no class id. */
async function loadDefaultClass(
  baseUrl: string,
  cookieHeader: string,
): Promise<DefaultClassLookup> {
  try {
    const response = await fetch(`${baseUrl}/academics/my-classes`, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
    });

    if (!response.ok) {
      return {
        classId: null,
        streamName: null,
        error: errorForStatus(response.status),
      };
    }

    let payload: unknown;

    try {
      payload = await response.json();
    } catch {
      return {
        classId: null,
        streamName: null,
        error: "invalid-response",
      };
    }

    const data = unwrapData(payload);

    if (!isRecord(data)) {
      return {
        classId: null,
        streamName: null,
        error: "invalid-response",
      };
    }

    const classes = Array.isArray(data.classes)
      ? data.classes.filter(isTeacherClass)
      : [];
    const requestedDefault = optionalString(data.default_class_id);
    const defaultClass =
      classes.find((item) => item.id === requestedDefault) ?? classes[0];

    return {
      classId: defaultClass?.id ?? null,
      streamName: defaultClass?.stream_name ?? null,
      error: null,
    };
  } catch {
    return { classId: null, streamName: null, error: "network" };
  }
}

/** Maps an HTTP status into the dashboard's error contract. */
function errorForStatus(status: number): RollCallError {
  if (status === 401) {
    return "unauthorized";
  }

  if (status === 403) {
    return "forbidden";
  }

  if (status === 404) {
    return "not-found";
  }

  return "server";
}

/** Returns the standard envelope's data, or the payload itself for BFFs. */
function unwrapData(payload: unknown): unknown {
  const record = getRecord(payload);
  return record && "data" in record ? record.data : payload;
}

function getRecord(value: unknown): Record<string, unknown> | null {
  return isRecord(value) ? value : null;
}

/** Validates and maps one class-teacher assignment. */
function isTeacherClass(value: unknown): value is {
  id: string;
  stream_name: string;
} {
  if (!isRecord(value)) {
    return false;
  }

  return (
    typeof value.id === "string" &&
    value.id.trim().length > 0 &&
    typeof value.stream_name === "string" &&
    value.stream_name.trim().length > 0
  );
}

/** Maps the roster endpoint rows into the dashboard's student contract. */
function mapRosterStudents(value: unknown): RollCallStudent[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.filter(isRosterStudent).map((student) => ({
    id: student.student_id,
    name: student.full_name,
    admissionNumber: student.admission_number,
  }));
}

/** Narrows an unknown payload entry to a complete roster row. */
function isRosterStudent(value: unknown): value is {
  student_id: string;
  full_name: string;
  admission_number: string;
} {
  if (!isRecord(value)) {
    return false;
  }

  const student = value as RosterStudentPayload;

  return (
    typeof student.student_id === "string" &&
    student.student_id.length > 0 &&
    typeof student.full_name === "string" &&
    student.full_name.trim().length > 0 &&
    typeof student.admission_number === "string" &&
    student.admission_number.trim().length > 0
  );
}

/** Narrows and normalizes the optional portal identity object. */
function normalizePortal(
  value: unknown,
): RollCallData["portal"] {
  if (!isRecord(value)) {
    return null;
  }

  return { name: optionalString(value.name) };
}

/** Narrows and normalizes the optional teacher session object. */
function normalizeSession(
  value: unknown,
): RollCallData["session"] {
  if (!isRecord(value)) {
    return null;
  }

  const fullName = optionalString(value.fullName);
  const role = optionalString(value.role);
  const initials = optionalString(value.initials);

  if (!fullName || !role || !initials) {
    return null;
  }

  return {
    fullName,
    role,
    subject: optionalString(value.subject),
    initials,
  };
}

/** Formats an ISO date in the school's timezone, otherwise uses today. */
function normalizeDateLabel(value: unknown) {
  const text = optionalString(value);

  if (text && !/^\d{4}-\d{2}-\d{2}$/.test(text)) {
    return text;
  }

  const date = text ? new Date(`${text}T12:00:00.000Z`) : new Date();

  if (Number.isNaN(date.getTime())) {
    return text;
  }

  return new Intl.DateTimeFormat("en-KE", {
    timeZone: "Africa/Nairobi",
    weekday: "short",
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(date);
}

/** Trims a string and maps blank or non-string values to null. */
function optionalString(value: unknown): string | null {
  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : null;
}

/** Checks that an unknown value is a plain object rather than an array. */
function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
