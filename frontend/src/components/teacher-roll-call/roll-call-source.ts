import { cookies } from "next/headers";
import type { RollCallData, RollCallStudent } from "./roll-call-data";

/**
 * The route remains a preview until attendance submission is connected. Flip
 * this off once the save endpoint is wired so the preview notice disappears.
 */
export const ROLL_CALL_IS_PREVIEW = true;

export type RollCallSelection = {
  /** Class whose roster should be loaded; resolved from the session. */
  classId?: string;
};

type RosterStudentPayload = {
  student_id?: unknown;
  full_name?: unknown;
  admission_number?: unknown;
};

type RosterPayload = {
  students?: unknown;
};

/**
 * Loads the class roster and current roll-call context for a teacher.
 *
 * Integration point:
 *  - Resolve the teacher, assigned class stream, term and date from the session.
 *  - Call the class roster and attendance endpoints.
 *  - Map the real response into `RollCallData`.
 *
 * The loader needs a resolved class id. It returns `null` when the API is not
 * configured or that context is unavailable; a successful empty roster is
 * returned as data so the dashboard can render its honest empty state.
 */
export async function loadRollCall(
  selection: RollCallSelection = {},
): Promise<RollCallData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;
  const classId = selection.classId?.trim();

  if (!apiUrl || !classId) {
    return null;
  }

  const url = `${apiUrl.replace(/\/$/, "")}/classes/${encodeURIComponent(
    classId,
  )}/roster`;
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

    const payload = (await response.json()) as RosterPayload;

    return {
      portal: null,
      session: null,
      context: {
        streamName: null,
        dateLabel: null,
        termLabel: null,
        registryRef: null,
      },
      students: mapRosterStudents(payload.students),
    };
  } catch {
    return null;
  }
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
  if (!value || typeof value !== "object") {
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
