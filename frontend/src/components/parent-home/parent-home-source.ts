import type { ParentHomeData } from "./parent-home-data";

/**
 * Loads the authenticated guardian's home dashboard.
 *
 * Integration point:
 *  - Resolve the guardian, linked children and selected academic term from the
 *    session.
 *  - Call the fee balance, attendance snapshot, next lesson, released results
 *    and calendar endpoints.
 *  - Map the real response into `ParentHomeData`.
 *
 * The loader returns `null` until that context is available. The dashboard
 * keeps its full structure and renders honest empty states instead of sample
 * balances, results or events.
 */
export async function loadParentHome(): Promise<ParentHomeData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated parent home request context is wired up yet.
  return null;
}
