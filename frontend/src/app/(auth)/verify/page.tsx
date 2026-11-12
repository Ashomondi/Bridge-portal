'use client';

import React, { useState, useEffect, useRef, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ShieldCheck, ArrowRight, Edit3, AlertCircle, Loader2, CheckCircle2, LockKeyhole, BookOpenText } from 'lucide-react';

export default function VerifyPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center text-slate-600">Loading verification…</div>}>
      <VerifyPageContent />
    </Suspense>
  );
}

function VerifyPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const phone = searchParams.get('phone') || '+254 7•• ••• 678';

  const [code, setCode] = useState(['', '', '', '', '', '']);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(42);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    let timer: NodeJS.Timeout;
    if (countdown > 0) {
      timer = setInterval(() => setCountdown((prev) => prev - 1), 1000);
    }
    return () => clearInterval(timer);
  }, [countdown]);

  const handleChange = (index: number, value: string) => {
    // Only accept numeric digits
    if (value && !/^\d+$/.test(value)) return;

    if (value.length > 1) value = value.slice(-1);

    const newCode = [...code];
    newCode[index] = value;
    setCode(newCode);

    // Auto-advance focus to the next input box when a digit is entered
    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    const fullCode = code.join('');
    if (fullCode.length < 6) {
      setError('Please enter the complete 6-digit verification code.');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const res = await fetch('/api/v1/auth/otp/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ phone_number: phone, code: fullCode }),
      });

      const data = await res.json();
      if (!res.ok) {
        const errorMessage = typeof data?.error?.message === 'string' ? data.error.message : "That code didn't match. Check the SMS and try again.";
        throw new Error(errorMessage);
      }

      localStorage.setItem('auth_token', data.token);
      router.push('/dashboard');
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'That code did not match. Check the SMS and try again.';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white px-4 py-6 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center gap-6">
        <div className="grid w-full overflow-hidden rounded-[32px] border border-slate-200/90 bg-white/70 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur-md lg:grid-cols-[1.05fr_0.95fr]">
          {/* Left Decorative Side - Hidden on Small Screens */}
          <div className="relative hidden overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(0,130,114,0.18),_transparent_30%),linear-gradient(135deg,#0c2b35_0%,#113a3d_35%,#0d5258_100%)] p-6 text-white sm:p-8 lg:block">
            <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.06),transparent_45%,rgba(255,255,255,0.04))]" />
            <div className="relative z-10 flex h-full flex-col justify-between gap-8">
              {/* Header Section */}
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

              {/* Main Content Section */}
              <div className="space-y-6">
                <div className="space-y-4">
                  <span className="inline-flex rounded-full border border-emerald-300/40 bg-emerald-400/10 px-3 py-1 text-[10px] font-semibold uppercase tracking-[0.22em] text-emerald-100">
                    sms verification
                  </span>
                  <h1 className="max-w-md text-3xl font-bold tracking-tight text-white sm:text-4xl xl:text-5xl">
                    One-time code to keep your account secure.
                  </h1>
                </div>

                <p className="max-w-md text-sm leading-7 text-slate-200/85">
                  We sent a six-digit security code to your registered mobile number so your school dashboard stays protected.
                </p>
              </div>

              {/* Feature Icons Section */}
              <div className="grid gap-3 sm:grid-cols-3">
                {[
                  { icon: BookOpenText, label: 'Academic access' },
                  { icon: LockKeyhole, label: 'Private session' },
                  { icon: CheckCircle2, label: 'Otp verified' },
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

          {/* Right Form Side */}
          <div className="bg-white p-6 sm:p-8 lg:p-10">
            <div className="mx-auto max-w-md">
              {/* Header / Logo Section */}
              <div className="mb-8 text-center lg:text-left">
                <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-[#008272] text-white shadow-md shadow-[#008272]/20">
                  <ShieldCheck className="h-6 w-6" />
                </div>
                <h2 className="text-3xl font-bold tracking-tight text-slate-900">Enter your code</h2>
                <p className="mt-2 text-sm leading-6 text-slate-600">
                  We sent a 6-digit code to <span className="font-semibold text-slate-800">{phone}</span>
                </p>
              </div>

              {/* Error Message Display */}
              {error && (
                <div className="mb-5 rounded-2xl border border-rose-200 bg-rose-50 p-3.5 text-xs text-rose-700 flex items-start gap-2.5">
                  <AlertCircle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
                  <span>{error}</span>
                </div>
              )}

              {/* 6-Digit OTP Input Form */}
              <form onSubmit={handleVerify} className="space-y-5">
                <div className="flex justify-between gap-2 sm:gap-3">
                  {code.map((digit, idx) => (
                    <input
                      key={idx}
                      ref={(el) => {
                        inputRefs.current[idx] = el;
                      }}
                      type="text"
                      inputMode="numeric"
                      maxLength={1}
                      value={digit}
                      onChange={(e) => handleChange(idx, e.target.value)}
                      onKeyDown={(e) => handleKeyDown(idx, e)}
                      placeholder="•"
                      className="h-12 w-12 rounded-xl border bg-[#f8faf9] border-slate-200 text-center text-xl font-bold font-mono text-slate-900 focus:border-[#008272] focus:ring-1 focus:ring-[#008272] focus:outline-none transition sm:h-14 sm:w-14"
                    />
                  ))}
                </div>

                {/* Countdown Timer */}
                <div className="text-center text-xs text-slate-500">
                  <span>⏱ Send a new code in </span>
                  <span className="font-semibold font-mono text-slate-700">0:{countdown < 10 ? `0${countdown}` : countdown}</span>
                </div>

                {/* Verify Button */}
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-[#008272] hover:bg-[#006e61] disabled:opacity-50 text-white font-semibold py-3.5 rounded-2xl transition flex items-center justify-center gap-2 shadow-lg shadow-[#008272]/25"
                >
                  {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : <>Verify <ArrowRight className="w-4 h-4" /></>}
                </button>
              </form>

              {/* Use a different number link */}
              <div className="mt-6 text-center">
                <button
                  onClick={() => router.push('/login')}
                  className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#008272] hover:underline"
                >
                  <Edit3 className="w-3.5 h-3.5" /> Use a different number
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}