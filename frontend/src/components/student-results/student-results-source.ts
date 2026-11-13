import { cookies } from "next/headers";
import type {
  CbcTrendPoint,
  StudentResultsData,
  SubjectResult,
} from "./student-results-data";

export type StudentResultsSelection = {
  studentId?: string;
  termId?: string;
};

export type StudentResultsError =
  | "unauthorized"
  | "forbidden"
  | "server"
  | "network";

export type StudentResultsLoadResult = {
  data: StudentResultsData | null;
  error: StudentResultsError | null;
};

/**
 * Loads released results for the selected child of the authenticated parent.
 *
 * Integration point:
 *  - Send the guardian session cookie to the authenticated results endpoint.
 *  - Request only the selected student/term and let the backend enforce
 *    guardian-child access.
 *  - Strip unreleased terms and records before returning data to the client.
 */
export async function loadStudentResults(
  selection: StudentResultsSelection = {},
): Promise<StudentResultsLoadResult> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return { data: null, error: null };
  }

  const params = new URLSearchParams();

  if (selection.studentId) {
    params.set("studentId", selection.studentId);
  }

  if (selection.termId) {
    params.set("termId", selection.termId);
  }

  const cookieHeader = cookies().toString();
  const query = params.toString();
  const url = `${apiUrl.replace(/\/$/, "")}/parent/results${
    query ? `?${query}` : ""
  }`;

  try {
    const response = await fetch(url, {
      cache: "no-store",
      headers: {
        Accept: "application/json",
        ...(cookieHeader ? { Cookie: cookieHeader } : {}),
      },
    });

    if (response.status === 401) {
      return { data: null, error: "unauthorized" };
    }

    if (response.status === 403) {
      return { data: null, error: "forbidden" };
    }

    if (!response.ok) {
      return { data: null, error: "server" };
    }

    const payload = (await response.json()) as Partial<StudentResultsData>;

    return { data: sanitizeResultsData(payload), error: null };
  } catch {
    return { data: null, error: "network" };
  }
}

/**
 * Removes unreleased terms and result records before they reach the client.
 * The selected term always resolves to a released term when one exists.
 */
function sanitizeResultsData(
  payload: Partial<StudentResultsData>,
): StudentResultsData {
  const terms = Array.isArray(payload.terms) ? payload.terms : [];
  const releasedTerms = terms.filter((term) => term.released);
  const selectedTermId =
    releasedTerms.find((term) => term.id === payload.selectedTermId)?.id ??
    releasedTerms[0]?.id ??
    null;
  const selectedTerm = releasedTerms.find(
    (term) => term.id === selectedTermId,
  );
  const results =
    selectedTerm && Array.isArray(payload.results)
      ? payload.results.map(normalizeSubjectResult)
      : [];
  const trend =
    selectedTerm && Array.isArray(payload.trend)
      ? payload.trend
          .filter(isReleasedTrendPoint)
          .map((point) => ({
            ...point,
            level: point.level ?? null,
            value: point.value ?? null,
          }))
      : [];
  const students = Array.isArray(payload.students) ? payload.students : [];
  const selectedStudent = students.find(
    (student) => student.id === payload.selectedStudentId,
  );

  return {
    portal: payload.portal ?? null,
    session: payload.session ?? null,
    students,
    selectedStudentId: selectedStudent?.id ?? students[0]?.id ?? null,
    terms,
    selectedTermId,
    results,
    trend,
    adjustment: selectedTerm ? payload.adjustment ?? null : null,
    teacherComment: selectedTerm
      ? payload.teacherComment
        ? {
            ...payload.teacherComment,
            attribution: payload.teacherComment.attribution ?? null,
            verifiedAt: payload.teacherComment.verifiedAt ?? null,
          }
        : null
      : null,
    levels: Array.isArray(payload.levels) ? payload.levels : [],
  };
}

function normalizeSubjectResult(result: SubjectResult): SubjectResult {
  return {
    ...result,
    cbcLevel: result.cbcLevel ?? null,
    amendmentNote: result.amendmentNote ?? null,
    activeTerms: Array.isArray(result.activeTerms) ? result.activeTerms : [],
  };
}

function isReleasedTrendPoint(point: CbcTrendPoint) {
  return Boolean(point.released && point.level && point.value !== null);
}
