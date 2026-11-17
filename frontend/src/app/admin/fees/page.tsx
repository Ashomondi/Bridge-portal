import type { Metadata } from "next";
import AdminFeeStructureDashboard from "@/components/admin-fees/AdminFeeStructureDashboard";
import { loadAdminFees } from "@/components/admin-fees/admin-fees-source";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Fee Structures | Bridge Portal",
};

export default async function AdminFeesPage() {
  const data = await loadAdminFees();

  return <AdminFeeStructureDashboard data={data} />;
}
