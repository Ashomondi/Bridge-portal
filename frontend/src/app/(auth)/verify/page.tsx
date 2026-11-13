'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ShieldCheck, ArrowRight, Edit3, AlertCircle, Loader2 } from 'lucide-react';

export default function VerifyPage() {
  // Start with empty strings so the user can type freely
  const [code, setCode] = useState(['', '', '', '', '', '']);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(42);
  const router = useRouter();
  const searchParams = useSearchParams();
  const phone = searchParams.get('phone') || '+254 7•• ••• 678';
  
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
      if (!res.ok) throw new Error(data.error?.message || 'That code didn\'t match. Check the SMS and try again.');

      // Store token and route to dashboard / membership selector
      localStorage.setItem('auth_token', data.token);
      router.push('/dashboard');
    } catch (err: any) {
      setError(err.message);
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

        {/* Main Verification Card */}
        <div className="bg-white border border-slate-200/80 rounded-[24px] p-6 shadow-sm space-y-5">
          <div className="text-center space-y-1">
            <h2 className="text-xl font-bold text-slate-900">Enter your code</h2>
            <p className="text-slate-600 text-sm leading-relaxed">
              We sent a 6-digit code to <span className="font-semibold text-slate-800">{phone}</span>
            </p>
          </div>

          {error && (
            <div className="bg-rose-50 border border-rose-200 text-rose-700 p-3.5 rounded-xl text-xs flex items-start gap-2.5">
              <AlertCircle className="w-4 h-4 text-rose-500 shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* 6-Digit OTP Box Inputs */}
          <form onSubmit={handleVerify} className="space-y-5">
            <div className="flex justify-between gap-2">
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
                  className="w-12 h-14 text-center text-xl font-bold font-mono rounded-xl border bg-[#f8faf9] border-slate-200 text-slate-900 focus:border-[#008272] focus:ring-1 focus:ring-[#008272] focus:outline-none transition"
                />
              ))}
            </div>

            <div className="text-center text-xs text-slate-500">
              <span>⏱ Send a new code in </span>
              <span className="font-semibold font-mono text-slate-700">0:{countdown < 10 ? `0${countdown}` : countdown}</span>
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
              onClick={() => router.push('/login')}
              className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#008272] hover:underline"
            >
              <Edit3 className="w-3.5 h-3.5" /> Use a different number
            </button>
          </div>
        </div>

        {/* Footer Audit Strip */}
        <div className="bg-white/90 border border-slate-200/80 p-4 rounded-[20px] text-center text-[11px] text-slate-500 border-dashed border-slate-300">
          <span> Direct encrypted SMS gateway • CBC Academic Portal</span>
        </div>

      </div>

    </div>
  );
}