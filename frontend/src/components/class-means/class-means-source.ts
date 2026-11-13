import type { ClassMeansData } from "./class-means-data";

/**
 * Loads aggregate class means for the authenticated teacher.
 *
 * Integration point:
 *  - Resolve the active school, class stream, academic period and assessment
 *    series from the session.
 *  - Call the class means analytics endpoint.
 *  - Map the real response into `ClassMeansData`.
 *
 * The loader returns `null` until that context is available. The dashboard
 * keeps its layout and renders structured empty states rather than sample
 * scores, subjects or learner counts.
 */
export async function loadClassMeans(): Promise<ClassMeansData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated class means request context is wired up yet.
  return null;
}
