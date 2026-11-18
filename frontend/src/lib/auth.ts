export type Membership = {
  school_id?: string | null;
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
  const base = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, '');
  return base ? `${base}${path}` : path;
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
  const cooldown = data?.error?.details?.cooldown_seconds;
  if (typeof cooldown === 'number') {
    return `${message || fallback} Please wait ${cooldown} seconds.`;
  }
  return typeof message === 'string' && message ? message : fallback;
}

export function roleRoute(role?: string | null): string | null {
  const normalized = role?.toLowerCase().trim();
  const routes: Record<string, string> = {
    admin: '/admin', administrator: '/admin', principal: '/admin', deputy: '/admin',
    teacher: '/teacher', parent: '/parent', bursar: '/bursar', librarian: '/librarian', nurse: '/nurse',
  };
  return normalized ? routes[normalized] ?? null : null;
}
