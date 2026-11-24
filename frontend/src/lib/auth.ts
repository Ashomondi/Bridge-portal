export type Membership = {
  school_id?: string | null;
  school_name?: string | null;
  role?: string | null;
  active?: boolean | null;
  status?: string | null;
};

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

/** Extracts a backend cooldown from error details or a successful response. */
export function getCooldownSeconds(data: any): number | null {
  const candidates = [
    data?.error?.details?.cooldown_seconds,
    data?.error?.details?.retry_after_seconds,
    data?.data?.cooldown_seconds,
    data?.cooldown_seconds,
  ];
  for (const cooldown of candidates) {
    if (typeof cooldown === 'number' && Number.isFinite(cooldown) && cooldown > 0) {
      return Math.ceil(cooldown);
    }
  }
  return null;
}

export async function readApiResponse(res: Response): Promise<{ ok: boolean; data: any }> {
  let data: any = null;
  const contentType = res.headers.get('content-type') || '';

  try {
    if (contentType.includes('application/json')) {
      data = await res.json();
    } else {
      const text = await res.text();
      data = text ? { error: { message: text } } : null;
    }
  } catch {
    data = null;
  }

  return { ok: res.ok, data };
}

export function getApiError(data: any, fallback: string): string {
  const message = data?.error?.message;
  const cooldown = getCooldownSeconds(data);
  if (cooldown !== null) {
    return `${message || fallback} Please wait ${cooldown} seconds.`;
  }
  return typeof message === 'string' && message ? message : fallback;
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
