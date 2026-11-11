import AdminUnmatchedPaymentsDashboard from "@/components/admin-unmatched-payments/AdminUnmatchedPaymentsDashboard";
import { loadAdminUnmatched } from "@/components/admin-unmatched-payments/admin-unmatched-source";

export const dynamic = "force-dynamic";

export default async function AdminUnmatchedPaymentsPage() {
  const data = await loadAdminUnmatched();

  return <AdminUnmatchedPaymentsDashboard data={data} />;
}
