import StudentResultsDashboard from "@/components/student-results/StudentResultsDashboard";
import { loadStudentResults } from "@/components/student-results/student-results-source";

export const dynamic = "force-dynamic";

export default async function ParentResultsPage() {
  const data = await loadStudentResults();

  return <StudentResultsDashboard data={data} />;
}
