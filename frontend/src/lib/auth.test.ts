import { describe, expect, it } from 'vitest';
import { apiUrl, getApiError, getCooldownSeconds, normalizeKenyanPhone, readApiResponse, roleRoute } from './auth';

describe('Kenyan phone normalization', () => {
  it('normalizes accepted Kenyan phone formats', () => {
    expect(normalizeKenyanPhone('0712 345 678')).toBe('+254712345678');
    expect(normalizeKenyanPhone('254712345678')).toBe('+254712345678');
    expect(normalizeKenyanPhone('+254 712-345-678')).toBe('+254712345678');
    expect(normalizeKenyanPhone('712345678')).toBe('+254712345678');
    expect(normalizeKenyanPhone('0112345678')).toBe('+254112345678');
  });

  it('rejects invalid Kenyan phone formats', () => {
    expect(normalizeKenyanPhone('254254712345678')).toBeNull();
    expect(normalizeKenyanPhone('+255712345678')).toBeNull();
    expect(normalizeKenyanPhone('0712ABC678')).toBeNull();
    expect(normalizeKenyanPhone('71234567')).toBeNull();
  });
});

describe('API response helpers', () => {
  it('returns cooldown details as a useful API error', () => {
    expect(getApiError({ error: { message: 'Too many requests', details: { cooldown_seconds: 42 } } }, 'Failed')).toBe('Too many requests Please wait 42 seconds.');
    expect(getApiError(null, 'Fallback')).toBe('Fallback');
  });

  it('handles JSON and non-JSON API responses', async () => {
    const jsonResponse = new Response(JSON.stringify({ error: { message: 'Invalid OTP' } }), { status: 400, headers: { 'content-type': 'application/json' } });
    const json = await readApiResponse(jsonResponse);
    expect(json.ok).toBe(false);
    expect(json.data.error?.message).toBe('Invalid OTP');

    const htmlResponse = new Response('<html>Bad gateway</html>', { status: 502, headers: { 'content-type': 'text/html' } });
    const html = await readApiResponse(htmlResponse);
    expect(html.ok).toBe(false);
    expect(html.data.error?.message).toBe('<html>Bad gateway</html>');
  });

  it('maps recognized roles to existing portals', () => {
    expect(roleRoute('teacher')).toBe('/teacher');
    expect(roleRoute('Principal')).toBe('/admin');
    expect(roleRoute('dean_of_studies')).toBe('/admin');
    expect(roleRoute('unknown')).toBeNull();
  });

  it('normalizes the configured API base without duplicating /api/v1', () => {
    const original = process.env.NEXT_PUBLIC_API_URL;
    try {
      process.env.NEXT_PUBLIC_API_URL = 'http://localhost:8080/api/v1';
      expect(apiUrl('/auth/otp/request')).toBe('http://localhost:8080/api/v1/auth/otp/request');
      expect(apiUrl('/api/v1/auth/otp/request')).toBe('http://localhost:8080/api/v1/auth/otp/request');

      process.env.NEXT_PUBLIC_API_URL = 'http://localhost:8080';
      expect(apiUrl('/auth/otp/request')).toBe('http://localhost:8080/api/v1/auth/otp/request');

      process.env.NEXT_PUBLIC_API_URL = 'http://localhost:8080/api/v1/';
      expect(apiUrl('/auth/otp/request')).toBe('http://localhost:8080/api/v1/auth/otp/request');

      delete process.env.NEXT_PUBLIC_API_URL;
      expect(apiUrl('/auth/otp/request')).toBe('/api/v1/auth/otp/request');
    } finally {
      if (original === undefined) delete process.env.NEXT_PUBLIC_API_URL;
      else process.env.NEXT_PUBLIC_API_URL = original;
    }
  });

  it('extracts a positive backend cooldown', () => {
    expect(getCooldownSeconds({ error: { details: { cooldown_seconds: 42 } } })).toBe(42);
    expect(getCooldownSeconds({ error: { details: { cooldown_seconds: 0 } } })).toBeNull();
    expect(getCooldownSeconds(null)).toBeNull();
  });
});
