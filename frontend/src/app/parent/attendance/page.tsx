import StudentAttendanceDashboard from "@/components/student-attendance/StudentAttendanceDashboard";
import { loadStudentAttendance } from "@/components/student-attendance/student-attendance-source";

export const dynamic = "force-dynamic";

export default async function ParentAttendancePage() {
  const data = await loadStudentAttendance();
  const now = new Date();

  return (
    <StudentAttendanceDashboard
      data={data}
      initialMonth={{ year: now.getFullYear(), month: now.getMonth() }}
    />
  );
}
