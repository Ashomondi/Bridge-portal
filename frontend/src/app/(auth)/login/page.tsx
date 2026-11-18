'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { ShieldCheck, Phone, Loader2, ArrowRight } from 'lucide-react';
import { apiUrl, getApiError, getCooldownSeconds, normalizeKenyanPhone, readApiResponse } from '@/lib/auth';

export default function LoginPage() {
  const [phoneNumber, setPhoneNumber] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);
  const router = useRouter();

  useEffect(() => {
    const updateCountdown = () => {
      const challenge = sessionStorage.getItem('pending_otp_challenge');
      if (!challenge) return setCountdown(0);
      try {
        const parsed = JSON.parse(challenge) as { resend_at?: number };
        setCountdown(Math.max(0, Math.ceil(((parsed.resend_at || 0) - Date.now()) / 1000)));
      } catch {
        setCountdown(0);
      }
    };
    updateCountdown();
    const timer = window.setInterval(updateCountdown, 1000);
    return () => window.clearInterval(timer);
  }, []);

  const handleRequestOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const normalizedPhone = normalizeKenyanPhone(phoneNumber);
    if (!normalizedPhone) {
      setError('Enter a valid Kenyan mobile number, for example 712 345 678.');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(apiUrl('/auth/otp/request'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone_number: normalizedPhone }),
      });
      const { ok, data } = await readApiResponse(res);
      if (!ok) {
        const cooldown = getCooldownSeconds(data);
        if (cooldown !== null) {
          const resendAt = Date.now() + cooldown * 1000;
          sessionStorage.setItem('pending_otp_challenge', JSON.stringify({ phone: normalizedPhone, resend_at: resendAt }));
          setCountdown(cooldown);
        }
        throw new Error(getApiError(data, 'Failed to dispatch SMS OTP'));
      }

      const resendAt = Date.now() + 60_000;
      sessionStorage.setItem('pending_otp_challenge', JSON.stringify({ phone: normalizedPhone, resend_at: resendAt }));
      setCountdown(60);
      router.replace('/verify');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred. Please check your connection.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#edf7f5] flex items-center justify-center p-4 sm:p-6 lg:p-10 text-slate-800">
      <div className="w-full max-w-6xl overflow-hidden rounded-[36px] border border-slate-200/80 bg-white/70 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur-md grid lg:grid-cols-[1.15fr_0.85fr]">
        <div className="hidden lg:flex flex-col justify-between bg-[radial-gradient(circle_at_top,_rgba(0,130,114,0.18),_rgba(15,23,42,0)_50%),linear-gradient(135deg,#0f172a,#123b36_45%,#0f766e)] p-10 text-white">
          <div>
            <div className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.2em] text-emerald-200">
              <ShieldCheck className="w-4 h-4" />
              Ikonex OS
            </div>
            <h1 className="mt-8 text-4xl font-bold tracking-tight">A modern school portal built for school operations.</h1>
            <p className="mt-4 max-w-md text-sm text-slate-200 leading-6">
              Secure access for parents, teachers, staff, and administration across one connected academy platform.
            </p>
          </div>

          <div className="grid gap-4 text-sm text-slate-200">
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
              <p className="font-semibold text-white">Security-first login</p>
              <p className="mt-1 text-slate-300">SMS verification with role-based access.</p>
            </div>
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
              <p className="font-semibold text-white">Connected workflows</p>
              <p className="mt-1 text-slate-300">Fees, attendance, health, library, and academic operations in one view.</p>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-center p-5 sm:p-8 lg:p-10 bg-white/70">
          <div className="w-full max-w-xl">
            <div className="text-center space-y-2 pb-6 border-b border-dashed border-slate-300">
              <div className="inline-flex p-3 bg-[#008272] rounded-2xl text-white mb-1 shadow-md shadow-[#008272]/20">
                <ShieldCheck className="w-6 h-6" />
              </div>
              <h2 className="text-2xl font-bold tracking-tight text-slate-900">Ikonex Academy</h2>
              <p className="text-[11px] font-semibold text-slate-500 tracking-wider">INSTITUTIONAL PORTAL</p>
            </div>

            <div className="bg-white border border-slate-200/80 rounded-[24px] p-5 sm:p-7 shadow-sm space-y-5 mt-6">
              <div className="space-y-1">
                <h3 className="text-2xl font-bold text-slate-900">Sign in</h3>
                <p className="text-slate-600 text-sm leading-relaxed">
                  Enter the mobile number linked to your school account. We&apos;ll send you a code by SMS.
                </p>
              </div>

              {error && (
                <div role="alert" className="bg-rose-50 border border-rose-200 text-rose-600 p-3 rounded-xl text-sm text-center">
                  {error}
                </div>
              )}

              <form onSubmit={handleRequestOTP} className="space-y-4">
                <div>
                  <label htmlFor="phone-number" className="block text-xs font-semibold uppercase tracking-wider text-slate-600 mb-1.5">
                    Mobile number
                  </label>
                  <div className="flex bg-[#f2f8f7] border border-slate-200 rounded-2xl overflow-hidden focus-within:border-[#008272] focus-within:ring-1 focus-within:ring-[#008272] transition">
                    <div className="flex items-center gap-1 px-3.5 bg-white border-r border-slate-200 text-slate-700 font-medium text-sm select-none shadow-sm">
                      <span className="text-[#008272] font-bold">KE</span>
                      <span className="text-slate-700 font-semibold">+254</span>
                    </div>
                    <input
                      id="phone-number"
                      type="tel"
                      inputMode="tel"
                      autoComplete="tel-national"
                      required
                      value={phoneNumber}
                      onChange={(e) => setPhoneNumber(e.target.value)}
                      placeholder="712 345 678"
                      className="min-w-0 w-full bg-transparent px-4 py-3 text-slate-900 placeholder-slate-400 focus:outline-none text-sm font-mono tracking-wide"
                    />
                  </div>
                  <p className="text-xs text-slate-500 mt-2" aria-live="polite">
                    {countdown > 0 ? `Resend available in ${countdown}s` : 'You can ask for a new code once a minute.'}
                  </p>
                </div>

                <button
                  type="submit"
                  disabled={loading || countdown > 0}
                  className="w-full bg-[#008272] hover:bg-[#006e61] disabled:opacity-50 text-white font-semibold py-3.5 rounded-2xl transition flex items-center justify-center gap-2 shadow-lg shadow-[#008272]/25"
                >
                  {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : <>Send code <ArrowRight className="w-4 h-4" /></>}
                </button>
              </form>

              <div className="bg-[#f5f3f9] border border-[#e8e2f3] rounded-2xl p-4 flex items-start gap-3">
                <div className="p-2.5 bg-purple-100 rounded-xl text-purple-700 shrink-0 mt-0.5">
                  <Phone className="w-4 h-4" />
                </div>
                <div className="text-xs text-slate-700 space-y-0.5">
                  <p className="font-semibold text-slate-900">New here?</p>
                  <p className="text-slate-600">Parents and staff register with a QR code from the school office.</p>
                </div>
              </div>
            </div>

            <div className="bg-white/90 border border-slate-200/80 p-4 rounded-[20px] shadow-sm text-xs text-slate-600 mt-6">
              <div className="flex items-center justify-center gap-2">
                <span className="w-2 h-2 rounded-full bg-purple-500"></span>
                <span className="font-medium">AUDIT & SECURITY RULES</span>
              </div>
              <div className="text-center text-slate-500 pt-2 mt-2 border-t border-slate-100">
                Secured by Ikonex OS
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}