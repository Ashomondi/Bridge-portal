import type { StudentResultsData } from "./student-results-data";

/**
 * Loads released results for the selected child of the authenticated parent.
 *
 * Integration point:
 *  - Resolve the guardian, children and selected academic term from the session.
 *  - Call the released-results, adjustment-note and teacher-comment endpoints.
 *  - Map the real response into `StudentResultsData`.
 *
 * The loader returns `null` until that context is available. The dashboard
 * keeps its structure and renders honest empty states rather than sample
 * results or comments.
 */
export async function loadStudentResults(): Promise<StudentResultsData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated results request context is wired up yet.
  return null;
}
