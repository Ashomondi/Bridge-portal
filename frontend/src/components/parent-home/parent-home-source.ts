import { cookies } from "next/headers";
import type { ParentHomeData, ParentHomeStudent } from "./parent-home-data";

export type ParentHomeSelection = {
  studentId?: string;
};

/**
 * Loads the authenticated guardian's home dashboard.
 *
 * The guardian session cookie is forwarded to the parent portal endpoint,
 * which resolves the guardian's linked students and returns the selected
 * student's fee, attendance, lesson, results and event summaries.
 *
 * Returns `null` when the portal is not configured, the session is not
 * authorized, or the response cannot be used. The dashboard then renders
 * honest empty states instead of sample data.
 */
export async function loadParentHome(
  selection: ParentHomeSelection = {},
): Promise<ParentHomeData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  const requestedStudentId = selection.studentId?.trim() || undefined;
  const params = new URLSearchParams();

  if (requestedStudentId) {
    params.set("studentId", requestedStudentId);
  }

  const query = params.toString();
  const url = `${apiUrl.replace(/\/$/, "")}/parent/home${
    query ? `?${query}` : ""
  }`;
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

    const payload = (await response.json()) as Partial<ParentHomeData>;

    return sanitizeParentHomeData(payload, requestedStudentId);
  } catch {
    return null;
  }
}

/**
 * Validates the portal payload and resolves it to one authorized student.
 *
 * A requested student that is not part of the guardian's linked students is
 * rejected rather than silently replaced with another student's dashboard.
 */
function sanitizeParentHomeData(
  payload: Partial<ParentHomeData>,
  requestedStudentId?: string,
): ParentHomeData | null {
  const students = Array.isArray(payload.students)
    ? payload.students.filter(isParentHomeStudent)
    : [];

  if (students.length === 0) {
    return null;
  }

  const selectedStudent =
    students.find((student) => student.id === requestedStudentId) ??
    students.find((student) => student.id === payload.selectedStudentId) ??
    students[0];

  if (requestedStudentId && selectedStudent.id !== requestedStudentId) {
    return null;
  }

  return {
    portal: payload.portal ?? null,
    session: payload.session ?? null,
    students,
    selectedStudentId: selectedStudent.id,
    fee: payload.fee ?? null,
    attendance: payload.attendance ?? null,
    nextLesson: payload.nextLesson ?? null,
    results: payload.results ?? null,
    events: Array.isArray(payload.events) ? payload.events : [],
  };
}

/** Narrows an unknown payload entry to a usable linked student. */
function isParentHomeStudent(value: unknown): value is ParentHomeStudent {
  if (!value || typeof value !== "object") {
    return false;
  }

  const student = value as Partial<ParentHomeStudent>;

  return (
    typeof student.id === "string" &&
    student.id.length > 0 &&
    typeof student.name === "string" &&
    typeof student.initials === "string"
  );
}
