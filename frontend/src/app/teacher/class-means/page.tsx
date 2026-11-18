import ClassMeansDashboard from "@/components/class-means/ClassMeansDashboard";
import { loadClassMeans } from "@/components/class-means/class-means-source";

export const dynamic = "force-dynamic";

export default async function ClassMeansPage() {
  const data = await loadClassMeans();

  return <ClassMeansDashboard data={data} />;
}
