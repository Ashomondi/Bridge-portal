import type { Metadata } from "next";
import AdminUnmatchedPaymentsDashboard from "@/components/admin-unmatched-payments/AdminUnmatchedPaymentsDashboard";
import { loadAdminUnmatched } from "@/components/admin-unmatched-payments/admin-unmatched-source";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Unmatched Payments | Bridge Portal",
};

export default async function AdminUnmatchedPaymentsPage() {
  const data = await loadAdminUnmatched();

  return <AdminUnmatchedPaymentsDashboard data={data} />;
}
