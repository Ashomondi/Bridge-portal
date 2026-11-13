import TeacherOverviewDashboard from "@/components/teacher-overview/TeacherOverviewDashboard";
import { loadTeacherOverview } from "@/components/teacher-overview/teacher-overview-source";

export const dynamic = "force-dynamic";

export default async function TeacherOverviewPage() {
  const overview = await loadTeacherOverview();

  return <TeacherOverviewDashboard overview={overview} />;
}
