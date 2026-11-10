import ParentTimetableDashboard from "@/components/timetable-events/ParentTimetableDashboard";
import { buildWeek } from "@/components/timetable-events/timetable-events-data";
import { loadParentTimetable } from "@/components/timetable-events/timetable-events-source";

export const dynamic = "force-dynamic";

export default async function ParentTimetablePage() {
  const data = await loadParentTimetable();

  return (
    <ParentTimetableDashboard
      data={data}
      week={buildWeek(new Date())}
    />
  );
}
