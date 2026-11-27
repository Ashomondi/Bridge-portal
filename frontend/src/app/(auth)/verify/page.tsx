'use client';

import React, { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ShieldCheck, ArrowRight, Edit3, AlertCircle, Loader2 } from 'lucide-react';
import { apiUrl, clearStoredAuth, getApiError, getCooldownSeconds, isMembership, readApiResponse, roleRoute, Membership } from '@/lib/auth';

type Challenge = { phone: string; resend_at: number };

type VerifyPayload = {
  token?: unknown;
  expires_at?: unknown;
  user?: unknown;
  memberships?: unknown;
};

export default function VerifyPage() {
  const [code, setCode] = useState(['', '', '', '', '', '']);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [phone, setPhone] = useState('');
  const [countdown, setCountdown] = useState(0);
  const [memberships, setMemberships] = useState<Membership[] | null>(null);
  const router = useRouter();
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    const raw = sessionStorage.getItem('pending_otp_challenge');
    if (!raw) {
      router.replace('/login');
      return;
    }

    try {
      const challenge = JSON.parse(raw) as Partial<Challenge>;
      if (typeof challenge.phone !== 'string' || typeof challenge.resend_at !== 'number') throw new Error('Invalid challenge');
      const resendAt = challenge.resend_at;
      setPhone(challenge.phone);
      setCountdown(Math.max(0, Math.ceil((resendAt - Date.now()) / 1000)));

      const timer = window.setInterval(() => {
        setCountdown(Math.max(0, Math.ceil((resendAt - Date.now()) / 1000)));
      }, 1000);
      return () => window.clearInterval(timer);
    } catch {
      sessionStorage.removeItem('pending_otp_challenge');
      router.replace('/login');
    }
  }, [router]);

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    const digits = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    if (!digits) return;
    e.preventDefault();
    setCode(digits.padEnd(6, '').split('').slice(0, 6));
    inputRefs.current[Math.min(digits.length, 5)]?.focus();
  };

  const handleChange = (index: number, value: string) => {
    const digits = value.replace(/\D/g, '');
    if (!digits) {
      const next = [...code];
      next[index] = '';
      setCode(next);
      return;
    }

    const next = [...code];
    for (let i = 0; i < digits.length && index + i < 6; i++) next[index + i] = digits[i];
    setCode(next);
    inputRefs.current[Math.min(index + digits.length, 5)]?.focus();
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !code[index] && index > 0) inputRefs.current[index - 1]?.focus();
    if (e.key === 'ArrowLeft') { e.preventDefault(); inputRefs.current[Math.max(0, index - 1)]?.focus(); }
    if (e.key === 'ArrowRight') { e.preventDefault(); inputRefs.current[Math.min(5, index + 1)]?.focus(); }
    if (e.key === 'Home') { e.preventDefault(); inputRefs.current[0]?.focus(); }
    if (e.key === 'End') { e.preventDefault(); inputRefs.current[5]?.focus(); }
  };

  const requestOtp = async () => {
    if (!phone) return;
    setLoading(true);
    setError('');
    try {
      const res = await fetch(apiUrl('/auth/otp/request'), {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone_number: phone }),
      });
      const { ok, data } = await readApiResponse(res);
      if (!ok) {
        const cooldown = getCooldownSeconds(data);
        if (cooldown !== null) {
          const resendAt = Date.now() + cooldown * 1000;
          sessionStorage.setItem('pending_otp_challenge', JSON.stringify({ phone, resend_at: resendAt }));
          setCountdown(cooldown);
        }
        throw new Error(getApiError(data, 'Unable to resend the code'));
      }
      const cooldown = getCooldownSeconds(data) ?? 60;
      const resendAt = Date.now() + cooldown * 1000;
      sessionStorage.setItem('pending_otp_challenge', JSON.stringify({ phone, resend_at: resendAt }));
      setCountdown(cooldown);
      setCode(['', '', '', '', '', '']);
      inputRefs.current[0]?.focus();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to resend the code');
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    const fullCode = code.join('');
    if (fullCode.length !== 6) { setError('Please enter the complete 6-digit verification code.'); return; }

    setLoading(true);
    setError('');
    try {
      const res = await fetch(apiUrl('/auth/otp/verify'), {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone_number: phone, otp: fullCode }),
      });
      const { ok, data } = await readApiResponse(res);
      if (!ok) throw new Error(getApiError(data, 'That code didn\'t match. Check the SMS and try again.'));

      const payload = data?.data as VerifyPayload | undefined;
      if (!payload || typeof payload.token !== 'string' || !payload.token || typeof payload.expires_at !== 'string' || !payload.expires_at || !payload.user || typeof payload.user !== 'object') {
        throw new Error('The login response was incomplete. Please try again.');
      }
      const expiry = Date.parse(payload.expires_at);
      if (!Number.isFinite(expiry) || expiry <= Date.now()) throw new Error('The login session has expired. Please request a new code.');

      // The backend only returns active memberships; validate the contract
      // instead of relying on fields the response does not include.
      const activeMemberships: Membership[] = Array.isArray(payload.memberships)
        ? payload.memberships.filter(isMembership)
        : [];

      sessionStorage.removeItem('pending_otp_challenge');
      if (activeMemberships.length === 0) {
        clearStoredAuth();
        setError('No active school membership is available for this account.');
        return;
      }

      sessionStorage.setItem('auth_token', payload.token);
      sessionStorage.setItem('auth_expires_at', payload.expires_at);
      sessionStorage.setItem('auth_user', JSON.stringify(payload.user));
      sessionStorage.setItem('auth_memberships', JSON.stringify(activeMemberships));

      if (activeMemberships.length === 1) {
        const route = roleRoute(activeMemberships[0].role);
        if (route) {
          sessionStorage.setItem('auth_membership', JSON.stringify(activeMemberships[0]));
          router.replace(route);
        } else {
          clearStoredAuth();
          setError(`No portal is configured for the ${activeMemberships[0].role || 'assigned'} role.`);
        }
      } else {
        setMemberships(activeMemberships);
        if (!activeMemberships.some((membership) => roleRoute(membership.role))) {
          setError('No portal is configured for the roles on this account.');
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setLoading(false);
    }
  };

  const selectMembership = (membership: Membership) => {
    const route = roleRoute(membership.role);
    if (!route) {
      setError(`No portal is configured for the ${membership.role || 'assigned'} role.`);
      return;
    }
    sessionStorage.setItem('auth_membership', JSON.stringify(membership));
    router.replace(route);
  };

  return (
    <div className="min-h-screen bg-[#edf7f5] flex items-center justify-center p-4 sm:p-6 lg:p-10 text-slate-800">
      <div className="w-full max-w-6xl overflow-hidden rounded-[36px] border border-slate-200/80 bg-white/70 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur-md grid lg:grid-cols-[1.15fr_0.85fr]">
        {/* Left Decorative Panel - Hidden on Small Screens */}
        <div className="hidden lg:flex flex-col justify-between bg-[radial-gradient(circle_at_top,_rgba(0,130,114,0.18),_rgba(15,23,42,0)_50%),linear-gradient(135deg,#0f172a,#123b36_45%,#0f766e)] p-10 text-white">
          <div>
            <div className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.2em] text-emerald-200">
              <ShieldCheck className="w-4 h-4" />
              Secure access
            </div>
            <h1 className="mt-8 text-4xl font-bold tracking-tight">Verify your identity to continue.</h1>
            <p className="mt-4 max-w-md text-sm text-slate-200 leading-6">
              Your one-time code protects access to fee records, academic reports, attendance, and school operations.
            </p>
          </div>

          <div className="rounded-2xl border border-white/10 bg-white/5 p-4 text-sm text-slate-200">
            <p className="font-semibold text-white">Protected session</p>
            <p className="mt-1 text-slate-300">The verification step ensures only the correct school account can access the portal.</p>
          </div>
        </div>

        {/* Right Form Panel */}
        <div className="flex items-center justify-center p-5 sm:p-8 lg:p-10 bg-white/70">
          <div className="w-full max-w-xl">
            {memberships ? (
              <div className="bg-white border border-slate-200/80 rounded-[24px] p-6 shadow-sm space-y-4">
                <h2 className="text-2xl font-bold text-slate-900">Choose a school role</h2>
                {memberships.map((membership) => {
                  const route = roleRoute(membership.role);
                  return (
                    <button
                      key={membership.id}
                      type="button"
                      disabled={!route}
                      onClick={() => selectMembership(membership)}
                      className="w-full rounded-xl border border-slate-200 px-4 py-3 text-left hover:border-[#008272] disabled:opacity-50"
                    >
                      <span className="font-semibold text-slate-900">{membership.school_name}</span>
                      <span className="block text-xs text-slate-500">{membership.role.replace(/_/g, ' ')}</span>
                    </button>
                  );
                })}
              </div>
            ) : (
              <div className="bg-white border border-slate-200/80 rounded-[24px] p-5 sm:p-7 shadow-sm space-y-5">
                <div className="text-center space-y-1">
                  <div className="inline-flex p-3 bg-[#008272] rounded-2xl text-white mb-3">
                    <ShieldCheck className="w-6 h-6" />
                  </div>
                  <h3 className="text-2xl font-bold text-slate-900">Enter your code</h3>
                  <p className="text-slate-600 text-sm leading-relaxed">
                    We sent a 6-digit code to <span className="font-semibold text-slate-800">{phone}</span>
                  </p>
                </div>

                {error && (
                  <div role="alert" className="bg-rose-50 border border-rose-200 text-rose-700 p-3.5 rounded-xl text-xs flex items-start gap-2.5">
                    <AlertCircle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
                    <span>{error}</span>
                  </div>
                )}

                {/* 6-Digit OTP Box Inputs */}
                <form onSubmit={handleVerify} className="space-y-5">
                  <div className="grid grid-cols-6 gap-1.5 sm:gap-3">
                    {code.map((digit, idx) => (
                      <input
                        key={idx}
                        ref={(el) => {
                          inputRefs.current[idx] = el;
                        }}
                        aria-label={`Digit ${idx + 1} of 6`}
                        type="text"
                        inputMode="numeric"
                        autoComplete={idx === 0 ? 'one-time-code' : 'off'}
                        value={digit}
                        onChange={(e) => handleChange(idx, e.target.value)}
                        onPaste={handlePaste}
                        onKeyDown={(e) => handleKeyDown(idx, e)}
                        placeholder="•"
                        className="min-w-0 w-full h-12 sm:h-16 text-center text-lg sm:text-xl font-bold font-mono rounded-xl border bg-[#f8faf9] border-slate-200 text-slate-900 focus:border-[#008272] focus:ring-1 focus:ring-[#008272] focus:outline-none transition"
                      />
                    ))}
                  </div>

                  <div className="text-center text-xs text-slate-500" aria-live="polite">
                    {countdown > 0 ? (
                      <>Send a new code in <span className="font-semibold font-mono text-slate-700">0:{countdown < 10 ? `0${countdown}` : countdown}</span></>
                    ) : (
                      <button
                        type="button"
                        onClick={requestOtp}
                        disabled={loading}
                        className="font-semibold text-[#008272] hover:underline"
                      >
                        Resend code
                      </button>
                    )}
                  </div>

                  <button
                    type="submit"
                    disabled={loading}
                    className="w-full bg-[#008272] hover:bg-[#006e61] disabled:opacity-50 text-white font-semibold py-3.5 rounded-2xl transition flex items-center justify-center gap-2 shadow-lg shadow-[#008272]/25"
                  >
                    {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : <>Verify <ArrowRight className="w-4 h-4" /></>}
                  </button>
                </form>

                {/* Use a different number link */}
                <div className="text-center pt-2">
                  <button
                    type="button"
                    onClick={() => {
                      sessionStorage.removeItem('pending_otp_challenge');
                      router.replace('/login');
                    }}
                    className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#008272] hover:underline"
                  >
                    <Edit3 className="w-3.5 h-3.5" /> Use a different number
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}