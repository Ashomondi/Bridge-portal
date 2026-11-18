import test from 'node:test';
import assert from 'node:assert/strict';
import { getApiError, normalizeKenyanPhone, readApiResponse, roleRoute } from './auth.ts';

test('normalizes accepted Kenyan phone formats', () => {
  assert.equal(normalizeKenyanPhone('0712 345 678'), '+254712345678');
  assert.equal(normalizeKenyanPhone('254712345678'), '+254712345678');
  assert.equal(normalizeKenyanPhone('+254 712-345-678'), '+254712345678');
  assert.equal(normalizeKenyanPhone('712345678'), '+254712345678');
  assert.equal(normalizeKenyanPhone('0112345678'), '+254112345678');
});

test('rejects invalid Kenyan phone formats', () => {
  assert.equal(normalizeKenyanPhone('254254712345678'), null);
  assert.equal(normalizeKenyanPhone('+255712345678'), null);
  assert.equal(normalizeKenyanPhone('0712ABC678'), null);
  assert.equal(normalizeKenyanPhone('71234567'), null);
});

test('returns cooldown details as a useful API error', () => {
  assert.equal(getApiError({ error: { message: 'Too many requests', details: { cooldown_seconds: 42 } } }, 'Failed'), 'Too many requests Please wait 42 seconds.');
  assert.equal(getApiError(null, 'Fallback'), 'Fallback');
});

test('handles JSON and non-JSON API responses', async () => {
  const jsonResponse = new Response(JSON.stringify({ error: { message: 'Invalid OTP' } }), { status: 400, headers: { 'content-type': 'application/json' } });
  const json = await readApiResponse(jsonResponse);
  assert.equal(json.ok, false);
  assert.equal(json.data.error.message, 'Invalid OTP');

  const htmlResponse = new Response('<html>Bad gateway</html>', { status: 502, headers: { 'content-type': 'text/html' } });
  const html = await readApiResponse(htmlResponse);
  assert.equal(html.ok, false);
  assert.equal(html.data.error.message, '<html>Bad gateway</html>');
});

test('maps recognized roles to existing portals', () => {
  assert.equal(roleRoute('teacher'), '/teacher');
  assert.equal(roleRoute('Principal'), '/admin');
  assert.equal(roleRoute('unknown'), null);
});
