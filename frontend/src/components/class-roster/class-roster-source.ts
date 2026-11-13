import type { ClassRosterData } from "./class-roster-data";

/**
 * The roster route renders a clearly labelled preview until the authenticated
 * roster endpoint is available. Flip this off once `loadClassRoster` returns
 * real data so the preview notice disappears automatically.
 */
export const CLASS_ROSTER_IS_PREVIEW = true;

/**
 * Loads the authenticated teacher's assigned class roster.
 *
 * Integration point:
 *  - Resolve the active school, class and teacher context from the session.
 *  - Call the roster, guardian and attendance endpoints.
 *  - Map the real response into `ClassRosterData`.
 *
 * No authenticated roster endpoint is wired up yet, so the loader returns
 * `null` and the dashboard keeps its structure and shows empty states rather
 * than inventing students, guardians or attendance records.
 */
export async function loadClassRoster(): Promise<ClassRosterData | null> {
  // No authenticated class roster request context is wired up yet.
  return null;
}
