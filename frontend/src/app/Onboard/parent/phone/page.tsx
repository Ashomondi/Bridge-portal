'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Clock, 
  CheckCircle2, 
  AlertCircle, 
  Loader2, 
  ArrowRight, 
  Lock 
} from 'lucide-react';

interface ParentPhoneOnboardData {
  stepCurrent: number;
  stepTotal: number;
  linkActiveStatus: string;
  isLinkOpen: boolean;
  phoneNumber: string;
  isVerified: boolean;
  schoolName: string;
}

export default function ParentPhoneOnboardPage() {
  const router = useRouter();

  // Dynamic state replacing mock data
  const [data, setData] = useState<ParentPhoneOnboardData | null>(null);
  const [phoneNumber, setPhoneNumber] = useState('712345678');
  const [codeDigits, setCodeDigits] = useState(['4', '8', '2', '9', '1', '5']);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isResending, setIsResending] = useState(false);

  // References for 6-digit inputs
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const getErrorMessage = (err: unknown, fallback: string): string => {
    if (err instanceof Error) {
      return err.message;
    }

    if (typeof err === 'string') {
      return err;
    }

    return fallback;
  };

  // Fetch initial onboarding state from backend API
  useEffect(() => {
    const fetchOnboardState = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch('/api/v1/onboard/parent/phone', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize parent onboarding state from institutional server.');
        }

        const result = await res.json();
        setData(result);
        if (result.phoneNumber) {
          setPhoneNumber(result.phoneNumber);
        }
      } catch (err: unknown) {
        setErrorMessage(getErrorMessage(err, 'An error occurred while loading verification details.'));
      } finally {
        setIsLoading(false);
      }
    };

    fetchOnboardState();
  }, []);

  // Handle digit change for 6-digit code inputs
  const handleDigitChange = (index: number, value: string) => {
    const char = value.slice(-1);
    const newDigits = [...codeDigits];
    newDigits[index] = char;
    setCodeDigits(newDigits);

    // Auto-focus next input
    if (char && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  // Handle Resend SMS code action
  const handleResendSMS = async () => {
    setIsResending(true);
    try {
      const res = await fetch('/api/v1/onboard/parent/resend-code', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ phone: `+254${phoneNumber}` }),
      });

      if (!res.ok) throw new Error('Failed to dispatch verification code via SMS.');
      alert('Verification code successfully re-dispatched via SMS.');
    } catch (err: unknown) {
      alert(getErrorMessage(err, 'Could not resend SMS.'));
    } finally {
      setIsResending(false);
    }
  };

  // Handle Continue / Submit verification
  const handleContinue = async () => {
    setIsSubmitting(true);
    try {
      const fullCode = codeDigits.join('');
      const res = await fetch('/api/v1/onboard/parent/verify-code', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          phone: `+254${phoneNumber}`,
          code: fullCode,
        }),
      });

      if (!res.ok) throw new Error('Invalid verification code. Please check and try again.');

      // Proceed to Step 2 (About You) in parent onboarding flow
      router.push('/onboard/parent/about');
    } catch (err: unknown) {
      alert(getErrorMessage(err, 'Verification failed.'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#eaf4f3] via-[#f2f7f6] to-white flex flex-col items-center justify-center text-slate-800 p-4">
      
      {/* Centered Web Container */}
      <div className="max-w-xl w-full bg-white border border-slate-200/90 rounded-[32px] shadow-xl overflow-hidden my-6">
        
        {/* Top Demo Banner Bar */}
        <div className="w-full bg-[#f8faf9] border-b border-slate-200 px-6 py-2.5 flex items-center justify-between text-xs text-slate-500">
          <span className="font-medium">Demo environment • sample data only</span>
          <div className="text-center font-bold text-slate-900">
            Ikonex Academy <span className="block font-normal text-[10px] text-slate-400">Lakeview Secondary</span>
          </div>
        </div>

        {/* Content Body */}
        <div className="p-6 md:p-8 space-y-6">
          
          {/* Link Active Status Header */}
          <div className="bg-[#f2f8f7] border border-[#d0e8e4] rounded-2xl px-4 py-3 flex items-center justify-between text-xs font-semibold">
            <span className="flex items-center gap-2 text-slate-700">
              <Clock className="w-4 h-4 text-[#008272]" /> {data?.linkActiveStatus || 'This link works until 6:00 pm today.'}
            </span>
            <span className="text-[11px] text-slate-500 font-bold">
              Lakeview Sec.
            </span>
          </div>

          {/* Step Progress Tracker (5 Steps) */}
          <div className="space-y-3">
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-wider text-slate-400">
              <span>STEP 1 OF 5</span>
              <span className="text-[#008272] font-semibold lowercase">Phone Verification</span>
            </div>
            
            <div className="grid grid-cols-5 gap-2">
              <div className="h-1.5 rounded-full bg-[#008272]"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
            </div>

            {/* Step Pills Bar */}
            <div className="grid grid-cols-5 gap-1.5 pt-1 text-[11px] font-bold text-center">
              <span className="py-1 rounded-xl bg-[#008272] text-white">1. Phone</span>
              <span className="py-1 rounded-xl bg-slate-100 text-slate-400">2. About You</span>
              <span className="py-1 rounded-xl bg-slate-100 text-slate-400">3. Child</span>
              <span className="py-1 rounded-xl bg-slate-100 text-slate-400">4. ID Doc</span>
              <span className="py-1 rounded-xl bg-slate-100 text-slate-400">5. Review</span>
            </div>
          </div>

          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {isLoading ? (
            <div className="py-24 flex flex-col items-center justify-center gap-3 text-slate-400">
              <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
              <p className="text-xs font-medium">Loading authentication settings...</p>
            </div>
          ) : (
            <div className="space-y-6">
              
              {/* Heading Title */}
              <div className="space-y-1">
                <h1 className="text-2xl font-extrabold text-slate-900 tracking-tight">Your phone number</h1>
                <p className="text-xs text-slate-500 leading-relaxed">
                  We&apos;ll text a code to this number. It becomes the number you use to sign in.
                </p>
              </div>

              {/* Mobile Phone Input */}
              <div className="space-y-1">
                <label className="block text-xs font-bold text-slate-700">Mobile phone number</label>
                <div className="flex rounded-2xl overflow-hidden border border-slate-200 bg-[#f4f8f7]">
                  <div className="bg-emerald-50 text-[#008272] px-4 py-3 font-bold text-xs flex items-center border-r border-slate-200">
                    +254
                  </div>
                  <input 
                    type="text" 
                    value={phoneNumber}
                    onChange={(e) => setPhoneNumber(e.target.value)}
                    placeholder="712 345 678"
                    className="w-full bg-transparent px-4 py-3 text-xs font-bold text-slate-900 focus:outline-none"
                  />
                  <div className="px-4 flex items-center">
                    <CheckCircle2 className="w-4 h-4 text-[#008272]" />
                  </div>
                </div>
                <p className="text-[10px] text-slate-400 pl-1">Kenyan Safaricom, Airtel, or Telkom line</p>
              </div>

              {/* 6-Digit SMS Verification Code Section */}
              <div className="space-y-3 bg-[#f8faf9] border border-slate-200/80 rounded-2xl p-4">
                <div className="flex items-center justify-between text-xs">
                  <span className="font-bold text-slate-700 flex items-center gap-1.5">
                    💬 Enter 6-digit SMS code
                  </span>
                  <button 
                    onClick={handleResendSMS}
                    disabled={isResending}
                    className="font-bold text-[#008272] hover:underline"
                  >
                    {isResending ? 'Sending...' : 'Resend code'}
                  </button>
                </div>

                <div className="grid grid-cols-6 gap-2 pt-1">
                  {codeDigits.map((digit, idx) => (
                    <input 
                      key={idx}
                      ref={(el) => { inputRefs.current[idx] = el; }}
                      type="text"
                      maxLength={1}
                      value={digit}
                      onChange={(e) => handleDigitChange(idx, e.target.value)}
                      className="w-full h-12 bg-white border border-slate-200 rounded-xl text-center text-lg font-extrabold text-slate-900 focus:outline-none focus:border-[#008272] shadow-sm transition"
                    />
                  ))}
                </div>

                {/* Verified Match Badge */}
                <div className="bg-[#f2f8f7] border border-[#cbe6e2] rounded-xl p-3 flex items-center gap-2.5 text-xs text-emerald-900 mt-2">
                  <CheckCircle2 className="w-4 h-4 text-[#008272] shrink-0" />
                  <div>
                    <strong className="block font-bold text-slate-900">Number verified</strong>
                    <span className="text-[11px] text-slate-500">Matched with Lakeview records</span>
                  </div>
                </div>
              </div>

              {/* No Password Required Notice */}
              <div className="bg-slate-50 border border-slate-200 rounded-2xl p-3.5 flex items-center gap-3 text-xs text-slate-600">
                <Lock className="w-4 h-4 text-slate-400 shrink-0" />
                <span>No password required. You will always sign in with a single-use SMS security code.</span>
              </div>

            </div>
          )}

        </div>

        {/* Footer Navigation Buttons Bar */}
        <div className="bg-[#f8faf9] border-t border-slate-200 px-6 py-4 flex items-center justify-between">
          <button 
            onClick={() => router.back()}
            className="px-6 py-2.5 bg-white hover:bg-slate-100 border border-slate-200 font-semibold rounded-xl text-xs text-slate-400 transition shadow-sm"
          >
            Back
          </button>
          
          <button 
            onClick={handleContinue}
            disabled={isSubmitting || isLoading}
            className="px-8 py-2.5 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition flex items-center gap-2 shadow-md shadow-[#008272]/20"
          >
            {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
            Continue <ArrowRight className="w-3.5 h-3.5" />
          </button>
        </div>

      </div>

    </div>
  );
}