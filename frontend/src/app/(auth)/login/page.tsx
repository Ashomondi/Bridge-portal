'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ShieldCheck, Phone, Loader2, ArrowRight } from 'lucide-react';

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
    } catch (err: any) {
      setError(err.message || 'An error occurred. Please check your connection.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white flex flex-col items-center justify-center p-4 text-slate-800">
      
      {/* Top Demo Banner */}
      <div className="text-xs font-medium text-slate-600 bg-white/80 backdrop-blur px-4 py-1.5 rounded-full border border-slate-200/60 mb-6 shadow-sm">
        Demo environment - sample data only
      </div>

      {/* Outer Wrapper Box */}
      <div className="w-full max-w-md bg-white/60 backdrop-blur-md border border-slate-200/90 rounded-[32px] p-6 shadow-2xl space-y-6">
        
        {/* Header / Logo Section */}
        <div className="text-center space-y-2 pb-2 border-b border-dashed border-slate-300">
          <div className="inline-flex p-3 bg-[#008272] rounded-2xl text-white mb-1 shadow-md shadow-[#008272]/20">
            <ShieldCheck className="w-6 h-6" />
          </div>
          <h1 className="text-xl font-bold tracking-tight text-slate-900">Ikonex Academy</h1>
          <p className="text-[11px] font-semibold text-slate-500 tracking-wider">INSTITUTIONAL PORTAL</p>
        </div>

        {/* Main Sign-In Card */}
        <div className="bg-white border border-slate-200/80 rounded-[24px] p-6 shadow-sm space-y-5">
          <div className="space-y-1">
            <h2 className="text-xl font-bold text-slate-900">Sign in</h2>
            <p className="text-slate-600 text-sm leading-relaxed">
              Enter the mobile number linked to your school account. We'll send you a code by SMS.
            </p>
          </div>

          {error && (
            <div className="bg-rose-50 border border-rose-200 text-rose-600 p-3 rounded-xl text-sm text-center">
              {error}
            </div>
          )}

          {/* Input Form */}
          <form onSubmit={handleRequestOTP} className="space-y-4">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-600 mb-1.5">
                Mobile number
              </label>
              <div className="flex bg-[#f2f8f7] border border-slate-200 rounded-2xl overflow-hidden focus-within:border-[#008272] focus-within:ring-1 focus-within:ring-[#008272] transition">
                <div className="flex items-center gap-1 px-3.5 bg-white border-r border-slate-200 text-slate-700 font-medium text-sm select-none shadow-sm">
                  <span className="text-[#008272] font-bold">KE</span>
                  <span className="text-slate-700 font-semibold">+254</span>
                </div>
                <input 
                  type="tel" 
                  required
                  value={phoneNumber}
                  onChange={(e) => setPhoneNumber(e.target.value)}
                  placeholder="712 345 678"
                  className="w-full bg-transparent px-4 py-3 text-slate-900 placeholder-slate-400 focus:outline-none text-sm font-mono tracking-wide"
                />
              </div>
              <p className="text-xs text-slate-500 mt-2 flex items-center gap-1">
                <span>⏱️</span> {rateLimitActive ? `Resend available in ${countdown}s` : 'You can ask for a new code once a minute.'}
              </p>
            </div>

            <button 
              type="submit" 
              disabled={loading || rateLimitActive}
              className="w-full bg-[#008272] hover:bg-[#006e61] disabled:opacity-50 text-white font-semibold py-3.5 rounded-2xl transition flex items-center justify-center gap-2 shadow-lg shadow-[#008272]/25"
            >
              {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : <>Send code <ArrowRight className="w-4 h-4" /></>}
            </button>
          </form>

          {/* QR Registration Info Box */}
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

        {/* Footer Audit Strip */}
        <div className="bg-white/90 border border-slate-200/80 p-4 rounded-[20px] shadow-sm space-y-2 text-xs text-slate-600">
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-1.5 font-medium">
              <span className="w-2 h-2 rounded-full bg-purple-500"></span> 
              AUDIT & SECURITY RULES
            </span>
            <a href="#simulate" className="text-[#008272] font-semibold hover:underline">Simulate Rate Limit State</a>
          </div>
          <div className="flex items-center justify-center gap-2 text-slate-500 pt-2 border-t border-slate-100">
            <span> 256-bit SMS encrypted OTP</span>
            <span>•</span>
            <span>Secured by Ikonex OS</span>
          </div>
        </div>

        {/* Bottom Support Link */}
        <div className="text-center pt-1 pb-2 space-y-1 border-t border-dashed border-slate-300">
          <p className="text-xs text-slate-500 pt-3">Having trouble accessing your account?</p>
          <a href="#contact" className="text-xs font-semibold text-[#008272] hover:underline block">
            Contact Ikonex Academic Registry Office
          </a>
        </div>

      </div>

    </div>
  );
}