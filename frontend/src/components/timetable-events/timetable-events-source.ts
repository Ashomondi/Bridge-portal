import type { ParentTimetableData } from "./timetable-events-data";

/**
 * Loads the selected child's timetable and key term dates.
 *
 * Integration point:
 *  - Resolve the guardian, linked children and selected academic term.
 *  - Call the timetable, calendar notice and key-date endpoints.
 *  - Map the real response into `ParentTimetableData`.
 *
 * The loader returns `null` until that context is available. The page keeps
 * its timetable and event structure and renders honest empty states rather
 * than sample periods or dates.
 */
export async function loadParentTimetable(): Promise<ParentTimetableData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated timetable request context is wired up yet.
  return null;
}
