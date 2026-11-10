import type { ClassRosterData } from "./class-roster-data";

/**
 * Loads the authenticated teacher's assigned class roster.
 *
 * Integration point:
 *  - Resolve the active school, class and teacher context from the session.
 *  - Call the roster, guardian and attendance endpoints.
 *  - Map the real response into `ClassRosterData`.
 *
 * The loader deliberately returns `null` until that context is available.
 * The dashboard keeps its structure and shows empty states rather than
 * inventing students, guardians or attendance records.
 */
export async function loadClassRoster(): Promise<ClassRosterData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated class roster request context is wired up yet.
  return null;
}
