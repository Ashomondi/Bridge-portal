import ParentHomeDashboard from "@/components/parent-home/ParentHomeDashboard";
import { loadParentHome } from "@/components/parent-home/parent-home-source";

export const dynamic = "force-dynamic";

export default async function ParentHomePage({
  searchParams,
}: {
  searchParams: { studentId?: string | string[] };
}) {
  const data = await loadParentHome({
    studentId:
      typeof searchParams.studentId === "string"
        ? searchParams.studentId
        : undefined,
  });

  return <ParentHomeDashboard data={data} />;
}
