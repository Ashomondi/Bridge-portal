import StudentResultsDashboard from "@/components/student-results/StudentResultsDashboard";
import { loadStudentResults } from "@/components/student-results/student-results-source";

export const dynamic = "force-dynamic";

export default async function ParentResultsPage({
  searchParams,
}: {
  searchParams: { studentId?: string | string[]; termId?: string | string[] };
}) {
  const data = await loadStudentResults({
    studentId:
      typeof searchParams.studentId === "string" ? searchParams.studentId : undefined,
    termId:
      typeof searchParams.termId === "string" ? searchParams.termId : undefined,
  });

  return <StudentResultsDashboard data={data} />;
}
