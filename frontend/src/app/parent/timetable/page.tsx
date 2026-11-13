import ParentTimetableDashboard from "@/components/timetable-events/ParentTimetableDashboard";
import { buildWeek } from "@/components/timetable-events/timetable-events-data";
import { loadParentTimetable } from "@/components/timetable-events/timetable-events-source";

export const dynamic = "force-dynamic";

/**
 * Loads the authenticated parent timetable and renders it for the current
 * school week. Week boundaries are resolved in the school's timezone by
 * `buildWeek`, not the frontend server's local timezone.
 */
export default async function ParentTimetablePage() {
  const data = await loadParentTimetable();

  return (
    <ParentTimetableDashboard
      data={data}
      week={buildWeek(new Date())}
    />
  );
}
