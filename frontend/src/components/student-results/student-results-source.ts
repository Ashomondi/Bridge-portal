import type { StudentResultsData } from "./student-results-data";

export type StudentResultsSelection = {
  studentId?: string;
  termId?: string;
};

/**
 * Loads released results for the selected child of the authenticated parent.
 *
 * Integration point:
 *  - Resolve the guardian and authorized children from the session.
 *  - Load the requested student/term (or session defaults) and return their IDs
 *    together with the matching results, adjustment and teacher comment.
 *  - Call the released-results, adjustment-note and teacher-comment endpoints.
 *  - Map the real response into `StudentResultsData`.
 *
 * The loader returns `null` until that context is available. The dashboard
 * keeps its structure and renders honest empty states rather than sample
 * results or comments.
 */
export async function loadStudentResults(
  selection: StudentResultsSelection = {},
): Promise<StudentResultsData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated results request context is wired up yet.
  // Do not return another selection's record as a fallback.
  void selection;
  return null;
}
