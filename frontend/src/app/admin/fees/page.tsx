import AdminFeeStructureDashboard from "@/components/admin-fees/AdminFeeStructureDashboard";
import { loadAdminFees } from "@/components/admin-fees/admin-fees-source";

export const dynamic = "force-dynamic";

export default async function AdminFeesPage() {
  const data = await loadAdminFees();

  return <AdminFeeStructureDashboard data={data} />;
}
