import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import VerifyPage from '../page';

const nav = vi.hoisted(() => ({
  push: vi.fn(),
  replace: vi.fn(),
}));

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: nav.push, replace: nav.replace }),
  usePathname: () => '/verify',
}));

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json' },
  });
}

describe('VerifyPage cooldown and lockout handling', () => {
  beforeEach(() => {
    nav.push.mockReset();
    nav.replace.mockReset();
    sessionStorage.clear();
    sessionStorage.setItem(
      'pending_otp_challenge',
      JSON.stringify({ phone: '+254712345678', resend_at: Date.now() - 1000 }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('uses the cooldown returned by the backend after resending', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse({ data: { message: 'OTP sent', cooldown_seconds: 42 } }),
      ),
    );
    const user = userEvent.setup();
    render(<VerifyPage />);

    await user.click(await screen.findByRole('button', { name: 'Resend code' }));

    expect(await screen.findByText(/Send a new code in/)).toHaveTextContent('0:42');
  });

  it('surfaces the backend lockout retry_after and holds the resend timer', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(
          {
            error: {
              code: 'SENDER_BLOCKED',
              message: 'Too many requests. Try again later.',
              details: { retry_after_seconds: 120 },
            },
          },
          429,
        ),
      ),
    );
    const user = userEvent.setup();
    render(<VerifyPage />);

    await user.click(await screen.findByRole('button', { name: 'Resend code' }));

    expect(
      await screen.findByRole('alert'),
    ).toHaveTextContent('Too many requests. Try again later. Please wait 120 seconds.');
    expect(screen.getByText(/Send a new code in/)).toHaveTextContent('0:120');
  });
});
