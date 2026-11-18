// Teacher portal: personal timetable rendered from school-configured data.
import TimetableDashboard from "@/components/timetable/TimetableDashboard";
import { buildWeek } from "@/components/timetable/timetable-data";
import { loadMyTimetable } from "@/components/timetable/timetable-source";

export const dynamic = "force-dynamic";

export default async function TeacherPage() {
  const timetable = await loadMyTimetable();
  const week = buildWeek(new Date(), timetable?.settings ?? null);

  return <TimetableDashboard timetable={timetable} week={week} />;
}
