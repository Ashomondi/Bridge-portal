import type { MyTimetableData } from "./timetable-data";

/**
 * Loads the authenticated teacher's timetable for the active school and term.
 *
 * Integration point:
 *  - Resolve the authenticated teacher, school, active term and assigned
 *    classes from the session.
 *  - Call the timetable API and map its slots into `MyTimetableData`.
 *
 * The loader deliberately returns `null` until that context is available. The
 * dashboard then renders an honest empty state instead of sample lessons.
 */
export async function loadMyTimetable(): Promise<MyTimetableData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated teacher/timetable request context is wired up yet.
  // Keep the real data contract ready rather than fabricating a schedule.
  return null;
}
