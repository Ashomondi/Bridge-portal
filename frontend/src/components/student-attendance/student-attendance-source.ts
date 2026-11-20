import type { StudentAttendanceData } from "./student-attendance-data";

/**
 * Loads the selected child's roll-call and attendance history.
 *
 * Integration point:
 *  - Resolve the guardian, children and selected academic term from the session.
 *  - Call the attendance summary, calendar-status and absence-record endpoints.
 *  - Map each child's summary, calendar and absence records into the
 *    student-keyed `attendanceByStudent` map.
 *
 * The loader returns `null` until that context is available. The dashboard
 * keeps its calendar and record structure and renders honest empty states
 * rather than sample attendance.
 */
export async function loadStudentAttendance(): Promise<StudentAttendanceData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated attendance request context is wired up yet.
  return null;
}
