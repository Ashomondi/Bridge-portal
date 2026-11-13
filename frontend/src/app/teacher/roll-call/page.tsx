import RollCallDashboard from "@/components/teacher-roll-call/RollCallDashboard";
import {
  ROLL_CALL_IS_PREVIEW,
  loadRollCall,
} from "@/components/teacher-roll-call/roll-call-source";

export const dynamic = "force-dynamic";

export default async function TeacherRollCallPage({
  searchParams,
}: {
  searchParams: { classId?: string | string[] };
}) {
  const data = await loadRollCall({
    classId:
      typeof searchParams.classId === "string"
        ? searchParams.classId
        : undefined,
  });

  return (
    <RollCallDashboard data={data} isPreview={ROLL_CALL_IS_PREVIEW} />
  );
}
