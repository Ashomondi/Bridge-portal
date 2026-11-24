import type { TeacherOverviewData } from "./teacher-overview-data";

/**
 * Loads the authenticated teacher's overview dashboard.
 *
 * Integration point:
 *  - Resolve the teacher, active school and academic period from the session.
 *  - Load the next lesson, pending tasks, today's timetable and announcements.
 *  - Map the real response into `TeacherOverviewData`.
 *
 * The loader returns `null` until that context is available. The dashboard
 * keeps its full structure and renders honest empty states instead of sample
 * lessons, tasks or announcements.
 */
export async function loadTeacherOverview(): Promise<TeacherOverviewData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated teacher overview request context is wired up yet.
  return null;
}
