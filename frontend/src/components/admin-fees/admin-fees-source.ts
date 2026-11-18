import type { AdminFeesData } from "./admin-fees-data";

/**
 * Loads the active school's fee matrix and distribution data.
 *
 * Integration point:
 *  - Resolve the operator, active institution and academic release context.
 *  - Call the fee structure, cohort and analytics endpoints.
 *  - Map the real response into `AdminFeesData`.
 *
 * The loader returns `null` until that context is available. The console keeps
 * its fee matrix and analytics structure and renders honest empty states rather
 * than sample amounts or cohorts.
 */
export async function loadAdminFees(): Promise<AdminFeesData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated fee-structure request context is wired up yet.
  return null;
}
