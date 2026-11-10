import ParentTimetableDashboard from "@/components/timetable-events/ParentTimetableDashboard";
import { buildWeek } from "@/components/timetable-events/timetable-events-data";

export const dynamic = "force-dynamic";

export default function ParentTimetablePage() {
  return <ParentTimetableDashboard week={buildWeek(new Date())} />;
}
