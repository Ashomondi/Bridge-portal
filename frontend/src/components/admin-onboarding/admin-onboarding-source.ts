import type { AdminOnboardingData } from "./admin-onboarding-data";

/**
 * Loads the school onboarding console data.
 *
 * Integration point:
 *  - Resolve the operator session and active institution context.
 *  - Call the onboarding, gateway health, metrics and registry endpoints.
 *  - Map the real response into `AdminOnboardingData`.
 *
 * The loader returns `null` until that context is available. The console keeps
 * its structure and renders honest empty states rather than sample institutions
 * or onboarding progress.
 */
export async function loadAdminOnboarding(): Promise<AdminOnboardingData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated admin onboarding request context is wired up yet.
  return null;
}
