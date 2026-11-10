import RollCallDashboard from "@/components/teacher-roll-call/RollCallDashboard";
import { loadRollCall } from "@/components/teacher-roll-call/roll-call-source";

export const dynamic = "force-dynamic";

export default async function TeacherRollCallPage() {
  const data = await loadRollCall();

  return <RollCallDashboard data={data} />;
}
