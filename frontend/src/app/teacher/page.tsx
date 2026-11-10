// Teacher portal: personal timetable with weekly date selection and status-aware cards.
import TimetableDashboard from "@/components/timetable/TimetableDashboard";
import { buildWeek } from "@/components/timetable/timetable-data";

export const dynamic = "force-dynamic";

export default function TeacherPage() {
  return <TimetableDashboard week={buildWeek(new Date())} />;
}
