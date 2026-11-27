/** Membership shape returned by POST /api/v1/auth/otp/verify. */
export type Membership = {
  id: string;
  school_id: string;
  school_name: string;
  role: string;
  badges?: string[];
};

export function isMembership(value: unknown): value is Membership {
  const record = asRecord(value);
  return Boolean(
    record &&
      typeof record.id === 'string' &&
      record.id &&
      typeof record.school_id === 'string' &&
      record.school_id &&
      typeof record.school_name === 'string' &&
      record.school_name &&
      typeof record.role === 'string' &&
      record.role,
  );
}

export function normalizeKenyanPhone(value: string): string | null {
  const cleaned = value.trim().replace(/[\s\-().]/g, '');

  let normalized: string;
  if (cleaned.startsWith('+254')) {
    normalized = cleaned;
  } else if (cleaned.startsWith('254')) {
    normalized = `+${cleaned}`;
  } else if (/^0[17]\d{8}$/.test(cleaned)) {
    normalized = `+254${cleaned.slice(1)}`;
  } else if (/^[17]\d{8}$/.test(cleaned)) {
    normalized = `+254${cleaned}`;
  } else {
    return null;
  }

  return /^\+254[17]\d{8}$/.test(normalized) ? normalized : null;
}

export function apiUrl(path: string): string {
  const rawBase = process.env.NEXT_PUBLIC_API_URL?.trim().replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/api/v1')
    ? path.slice('/api/v1'.length)
    : path;
  const suffix = normalizedPath.startsWith('/') ? normalizedPath : `/${normalizedPath}`;

  if (!rawBase) {
    return `/api/v1${suffix}`;
  }

  const base = rawBase.endsWith('/api/v1') ? rawBase : `${rawBase}/api/v1`;
  return `${base}${suffix}`;
}

export type ApiResponse = {
  data?: unknown;
  memberships?: unknown;
  error?: {
    message?: string;
    details?: {
      cooldown_seconds?: number;
      retry_after_seconds?: number;
      max_attempts?: number;
    };
  };
};

function asRecord(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null
    ? (value as Record<string, unknown>)
    : null;
}

/**
 * Extracts a retry delay from either the OTP cooldown, a sender/IP lockout, or
 * a successful OTP request (the backend returns all three in different places).
 */
export function getCooldownSeconds(data: unknown): number | null {
  const record = asRecord(data);
  if (!record) return null;

  const error = asRecord(record.error);
  const details = asRecord(error?.details);
  const success = asRecord(record.data);
  const cooldown =
    details?.cooldown_seconds ??
    details?.retry_after_seconds ??
    success?.cooldown_seconds ??
    record.cooldown_seconds;

  return typeof cooldown === 'number' && Number.isFinite(cooldown) && cooldown > 0
    ? Math.ceil(cooldown)
    : null;
}

export async function readApiResponse(
  res: Response,
): Promise<{ ok: boolean; data: ApiResponse }> {
  let data: ApiResponse = {};
  const contentType = res.headers.get('content-type') || '';

  try {
    if (contentType.includes('application/json')) {
      data = (await res.json()) as ApiResponse;
    } else {
      const text = await res.text();
      data = text ? { error: { message: text } } : {};
    }
  } catch {
    data = {};
  }

  return { ok: res.ok, data };
}

export function getApiError(data: unknown, fallback: string): string {
  const error = asRecord(asRecord(data)?.error);
  const message = typeof error?.message === 'string' && error.message ? error.message : fallback;
  const retryAfter = getCooldownSeconds(data);
  return retryAfter !== null
    ? `${message} Please wait ${retryAfter} seconds.`
    : message;
}

export function clearStoredAuth() {
  if (typeof window === "undefined") return;
  for (const storage of [window.sessionStorage, window.localStorage]) {
    for (const key of [
      "auth_token",
      "auth_expires_at",
      "auth_user",
      "auth_membership",
      "auth_memberships",
      "pending_otp_challenge",
    ]) {
      storage.removeItem(key);
    }
  }
}

export function roleRoute(role?: string | null): string | null {
  const normalized = role?.toLowerCase().trim();
  const routes: Record<string, string> = {
    admin: '/admin', administrator: '/admin', principal: '/admin', deputy: '/admin',
    dean: '/admin', dean_of_studies: '/admin',
    teacher: '/teacher', parent: '/parent', bursar: '/bursar', librarian: '/librarian', nurse: '/nurse',
  };
  return normalized ? routes[normalized] ?? null : null;
}
