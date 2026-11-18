import type { Metadata } from "next";
import AdminConnectionsDashboard from "@/components/admin-connections/AdminConnectionsDashboard";
import { loadAdminConnections } from "@/components/admin-connections/admin-connections-source";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Connections | Bridge Portal",
};

export default async function AdminConnectionsPage() {
  const data = await loadAdminConnections();

  return <AdminConnectionsDashboard data={data} />;
}
