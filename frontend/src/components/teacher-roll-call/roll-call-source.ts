import type { RollCallData } from "./roll-call-data";

/**
 * Loads the class roster and current roll-call context for a teacher.
 *
 * Integration point:
 *  - Resolve the teacher, assigned class stream, term and date from the session.
 *  - Call the class roster and attendance endpoints.
 *  - Map the real response into `RollCallData`.
 *
 * The loader returns `null` until that context is available. The page keeps
 * its attendance structures and renders honest empty states rather than
 * inventing learner records.
 */
export async function loadRollCall(): Promise<RollCallData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated roll-call request context is wired up yet.
  return null;
}
