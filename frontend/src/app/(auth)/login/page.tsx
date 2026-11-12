'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  ShieldCheck,
  Phone,
  Loader2,
  ArrowRight,
  CheckCircle2,
  LockKeyhole,
  BookOpenText,
} from 'lucide-react';

export default function LoginPage() {
  const [phoneNumber, setPhoneNumber] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [rateLimitActive, setRateLimitActive] = useState(false);
  const [countdown, setCountdown] = useState(60);
  const router = useRouter();

  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (rateLimitActive && countdown > 0) {
      timer = setInterval(() => setCountdown((prev) => prev - 1), 1000);
    } else if (countdown === 0) {
      setRateLimitActive(false);
      setCountdown(60);
    }
    return () => clearInterval(timer);
  }, [rateLimitActive, countdown]);

  const handleRequestOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    let normalizedPhone = phoneNumber.trim();
    if (normalizedPhone.startsWith('07') || normalizedPhone.startsWith('01')) {
      normalizedPhone = '+254' + normalizedPhone.slice(1);
    } else if (!normalizedPhone.startsWith('+')) {
      normalizedPhone = '+254' + normalizedPhone;
    }

    try {
      const res = await fetch('/api/v1/auth/otp/request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone_number: normalizedPhone }),
      });

      const data = await res.json();
      if (!res.ok) throw new Error(data.error?.message || 'Failed to dispatch SMS OTP');

      setRateLimitActive(true);
      router.push(`/verify?phone=${encodeURIComponent(normalizedPhone)}`);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'An error occurred. Please check your connection.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white px-4 py-6 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center gap-6">
        <div className="grid w-full overflow-hidden rounded-[32px] border border-slate-200/90 bg-white/70 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur-md lg:grid-cols-[1.2fr_0.8fr]">
          <div className="relative overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(0,130,114,0.18),_transparent_30%),linear-gradient(135deg,#0c2b35_0%,#113a3d_35%,#0d5258_100%)] p-6 text-white sm:p-8 lg:p-10">
            <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.06),transparent_45%,rgba(255,255,255,0.04))]" />
            <div className="relative z-10 flex h-full flex-col justify-between gap-8">
              <div className="flex items-center justify-between gap-4">
                <div className="inline-flex items-center gap-3 rounded-2xl border border-white/15 bg-white/5 px-3 py-2 backdrop-blur-sm">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#008272] text-white shadow-md shadow-[#008272]/30">
                    <ShieldCheck className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-xs uppercase tracking-[0.22em] text-emerald-100/80">Ikonex</p>
                    <p className="text-sm font-semibold">Academy Portal</p>
                  </div>
                </div>
                <div className="rounded-full border border-white/15 bg-white/5 px-3 py-1.5 text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-100">
                  secure access
                </div>
              </div>

              <div className="space-y-6">
                <div className="space-y-4">
                  <span className="inline-flex rounded-full border border-emerald-300/40 bg-emerald-400/10 px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.22em] text-emerald-100">
                    school operations
                  </span>
                  <h1 className="max-w-md text-4xl font-bold tracking-tight text-white xl:text-5xl">
                    One sign-in for every school touchpoint.
                  </h1>
                </div>

                <p className="max-w-md text-sm leading-7 text-slate-200/85">
                  Parents, teachers, and school staff access attendance, homework, fee updates, and school notices from a single secure dashboard.
                </p>
              </div>

              <div className="grid gap-3 sm:grid-cols-3">
                {[
                  { icon: BookOpenText, label: 'Academic records' },
                  { icon: LockKeyhole, label: 'Verified access' },
                  { icon: CheckCircle2, label: 'Real-time alerts' },
                ].map(({ icon: Icon, label }) => (
                  <div key={label} className="rounded-2xl border border-white/10 bg-white/5 p-3 backdrop-blur-sm">
                    <div className="mb-3 inline-flex rounded-xl bg-white/10 p-2 text-emerald-100">
                      <Icon className="h-4 w-4" />
                    </div>
                    <p className="text-xs font-medium text-slate-100/90">{label}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="bg-white p-6 sm:p-8 lg:p-10">
            <div className="mx-auto max-w-md">
              <div className="mb-8 text-center lg:text-left">
                <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-[#008272] text-white shadow-md shadow-[#008272]/20">
                  <ShieldCheck className="h-6 w-6" />
                </div>
                <h2 className="text-3xl font-bold tracking-tight text-slate-900">Welcome back</h2>
                <p className="mt-2 text-sm leading-6 text-slate-600">
                  Enter your mobile number to receive a secure OTP and continue into your school portal.
                </p>
              </div>

              {error && (
                <div className="mb-5 rounded-2xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-600">
                  {error}
                </div>
              )}

              <form onSubmit={handleRequestOTP} className="space-y-5">
                <div>
                  <label className="mb-2 block text-xs font-semibold uppercase tracking-[0.18em] text-slate-600">
                    Mobile number
                  </label>
                  <div className="flex overflow-hidden rounded-2xl border border-slate-200 bg-[#f2f8f7] transition focus-within:border-[#008272] focus-within:ring-1 focus-within:ring-[#008272]">
                    <div className="flex items-center gap-1 border-r border-slate-200 bg-white px-3.5 text-sm font-medium text-slate-700 shadow-sm">
                      <span className="font-bold text-[#008272]">KE</span>
                      <span>+254</span>
                    </div>
                    <input
                      type="tel"
                      required
                      value={phoneNumber}
                      onChange={(e) => setPhoneNumber(e.target.value)}
                      placeholder="712 345 678"
                      className="w-full bg-transparent px-4 py-3.5 text-sm font-mono tracking-wide text-slate-900 placeholder:text-slate-400 focus:outline-none"
                    />
                  </div>
                  <p className="mt-2 text-xs text-slate-500">
                    {rateLimitActive ? `Resend available in ${countdown}s` : 'You can ask for a new code once a minute.'}
                  </p>
                </div>

                <button
                  type="submit"
                  disabled={loading || rateLimitActive}
                  className="flex w-full items-center justify-center gap-2 rounded-2xl bg-[#008272] px-4 py-3.5 text-sm font-semibold text-white shadow-lg shadow-[#008272]/25 transition hover:bg-[#006e61] disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {loading ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    <>
                      Send code
                      <ArrowRight className="h-4 w-4" />
                    </>
                  )}
                </button>
              </form>

              <div className="mt-6 rounded-2xl border border-[#e7e1f2] bg-[#f5f3f9] p-4">
                <div className="flex items-start gap-3">
                  <div className="mt-0.5 flex h-10 w-10 items-center justify-center rounded-xl bg-purple-100 text-purple-700">
                    <Phone className="h-4 w-4" />
                  </div>
                  <div className="text-sm text-slate-700">
                    <p className="font-semibold text-slate-900">New here?</p>
                    <p className="mt-1 text-slate-600">Parents and staff register with a QR code from the school office.</p>
                  </div>
                </div>
              </div>

              <div className="mt-6 rounded-[20px] border border-slate-200 bg-white p-4 text-xs text-slate-600 shadow-sm">
                <div className="flex items-center justify-between gap-3">
                  <span className="flex items-center gap-2 font-medium text-slate-700">
                    <span className="h-2 w-2 rounded-full bg-purple-500" />
                    Audit & security rules
                  </span>
                  <a href="#simulate" className="font-semibold text-[#008272] hover:underline">
                    Simulate rate limit state
                  </a>
                </div>
                <div className="mt-3 flex items-center justify-center gap-2 border-t border-slate-100 pt-3 text-slate-500">
                  <span>256-bit SMS encrypted OTP</span>
                  <span>•</span>
                  <span>Secured by Ikonex OS</span>
                </div>
              </div>

              <div className="mt-6 border-t border-dashed border-slate-300 pt-4 text-center text-xs text-slate-500">
                <p>Having trouble accessing your account?</p>
                <a href="#contact" className="mt-1 inline-block font-semibold text-[#008272] hover:underline">
                  Contact Ikonex Academic Registry Office
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}