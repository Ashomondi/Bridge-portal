// Teacher portal: class roster with learner details and guardian/attendance context.
import ClassRosterDashboard from "@/components/class-roster/ClassRosterDashboard";
import { loadClassRoster } from "@/components/class-roster/class-roster-source";

export const dynamic = "force-dynamic";

export const metadata = {
  title: "My Class Roster | Ikonex Academy",
};

export default async function ClassRosterPage() {
  const roster = await loadClassRoster();

  return <ClassRosterDashboard roster={roster} />;
}
