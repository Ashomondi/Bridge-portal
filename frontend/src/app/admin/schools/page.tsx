import type { Metadata } from "next";
import AdminOnboardingDashboard from "@/components/admin-onboarding/AdminOnboardingDashboard";
import { loadAdminOnboarding } from "@/components/admin-onboarding/admin-onboarding-source";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "School Onboarding | Bridge Portal",
};

export default async function AdminSchoolsPage() {
  const data = await loadAdminOnboarding();

  return <AdminOnboardingDashboard data={data} />;
}
