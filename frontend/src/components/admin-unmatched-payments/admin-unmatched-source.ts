import type { AdminUnmatchedData } from "./admin-unmatched-data";

/**
 * Loads the unmatched direct-payment queue and active reconciliation context.
 *
 * Integration point:
 *  - Resolve the operator, active institution and gateway context.
 *  - Call the unmatched transactions, reconciliation metrics and inspector
 *    endpoints.
 *  - Map the real response into `AdminUnmatchedData`.
 *
 * The loader returns `null` until that context is available. The console keeps
 * its table, inspector and metrics structure and renders honest empty states
 * rather than sample transactions.
 */
export async function loadAdminUnmatched(): Promise<AdminUnmatchedData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated unmatched-payments request context is wired up yet.
  return null;
}
