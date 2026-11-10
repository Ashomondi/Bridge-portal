import ParentHomeDashboard from "@/components/parent-home/ParentHomeDashboard";
import { loadParentHome } from "@/components/parent-home/parent-home-source";

export const dynamic = "force-dynamic";

export default async function ParentHomePage() {
  const data = await loadParentHome();

  return <ParentHomeDashboard data={data} />;
}
