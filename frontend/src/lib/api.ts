import { apiUrl } from "./auth";

/** Small typed fetch wrapper for API modules that do not need a full SDK. */
export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(apiUrl(path), init);
  let body: unknown = null;

  try {
    body = await response.json();
  } catch {
    body = null;
  }

  if (!response.ok) {
    const errorBody = body as { error?: { message?: unknown } } | null;
    const message = errorBody?.error?.message;
    throw new Error(
      typeof message === "string" && message
        ? message
        : `Request failed (${response.status})`,
    );
  }

  const payload = body as { data?: unknown } | null;
  return (payload?.data ?? body) as T;
}
