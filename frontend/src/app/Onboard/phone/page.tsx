'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { 
  ShieldCheck, 
  Clock, 
  CheckCircle2, 
  AlertCircle, 
  Loader2, 
  ArrowRight, 
  ArrowLeft 
} from 'lucide-react';

interface PhoneOnboardData {
  stepCurrent: number;
  stepTotal: number;
  linkActiveStatus: string;
  isLinkOpen: boolean;
  phoneNumber: string;
  isVerified: boolean;
  schoolName: string;
}

export default function PhoneOnboardPage() {
  const router = useRouter();

  // Dynamic state replacing mock data
  const [data, setData] = useState<PhoneOnboardData | null>(null);
  const [phoneNumber, setPhoneNumber] = useState('733000111');
  const [codeDigits, setCodeDigits] = useState(['4', '8', '2', '9', '1', '5']);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isResending, setIsResending] = useState(false);

  // References for 6-digit inputs
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  // Fetch initial onboarding state from backend API
  useEffect(() => {
    const fetchOnboardState = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch('/api/v1/onboard/phone', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize onboarding state from institutional server.');
        }

        const result = await res.json();
        setData(result);
        if (result.phoneNumber) {
          setPhoneNumber(result.phoneNumber);
        }
      } catch (err: any) {
        setErrorMessage(err.message || 'An error occurred while loading verification details.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchOnboardState();
  }, []);

  // Handle digit change for 6-digit code inputs
  const handleDigitChange = (index: number, value: string) => {
    const char = value.slice(-1); // Take only the last character
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
      const res = await fetch('/api/v1/onboard/resend-code', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ phone: `+254${phoneNumber}` }),
      });

      if (!res.ok) throw new Error('Failed to dispatch verification code via SMS.');
      alert('Verification code successfully re-dispatched via SMS.');
    } catch (err: any) {
      alert(err.message || 'Could not resend SMS.');
    } finally {
      setIsResending(false);
    }
  };

  // Handle Continue / Submit verification
  const handleContinue = async () => {
    setIsSubmitting(true);
    try {
      const fullCode = codeDigits.join('');
      const res = await fetch('/api/v1/onboard/verify-code', {
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

      // Proceed to next step in onboarding flow
      router.push('/onboard/profile');
    } catch (err: any) {
      alert(err.message || 'Verification failed.');
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
              <Clock className="w-4 h-4 text-[#008272]" /> {data?.linkActiveStatus || 'Link active: 8:00 am – 6:00 pm today'}
            </span>
            <span className="px-2.5 py-0.5 rounded-full bg-emerald-100 text-[#008272] font-bold text-[10px]">
              {data?.isLinkOpen !== false ? 'Open' : 'Closed'}
            </span>
          </div>

          {/* Step Progress Tracker */}
          <div className="space-y-2">
            <div className="flex items-center justify-between text-xs font-bold uppercase tracking-wider text-slate-400">
              <span>Step {data?.stepCurrent || 1} of {data?.stepTotal || 4}</span>
              <span className="text-slate-700 font-semibold lowercase">Phone verification</span>
            </div>
            
            <div className="grid grid-cols-4 gap-2">
              <div className="h-1.5 rounded-full bg-[#008272]"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
              <div className="h-1.5 rounded-full bg-slate-200"></div>
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
                  We will send an encrypted authentication code for your Lakeview portal credentials.
                </p>
              </div>

              {/* Mobile Phone Input */}
              <div className="space-y-1.5">
                <label className="block text-xs font-bold text-slate-700">Mobile phone (Kenya)</label>
                <div className="flex rounded-2xl overflow-hidden border border-slate-200 bg-[#f4f8f7]">
                  <div className="bg-slate-100 px-4 py-3 text-slate-600 font-semibold text-xs flex items-center border-r border-slate-200">
                    +254
                  </div>
                  <input 
                    type="text" 
                    value={phoneNumber}
                    onChange={(e) => setPhoneNumber(e.target.value)}
                    placeholder="733000111"
                    className="w-full bg-transparent px-4 py-3 text-xs font-semibold text-slate-900 focus:outline-none"
                  />
                </div>
              </div>

              {/* 6-Digit Verification Code Section */}
              <div className="space-y-3 pt-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-slate-400">Code dispatched via SMS</span>
                  <button 
                    onClick={handleResendSMS}
                    disabled={isResending}
                    className="font-bold text-[#008272] hover:underline"
                  >
                    {isResending ? 'Sending...' : 'Resend SMS'}
                  </button>
                </div>

                <label className="block text-xs font-bold text-slate-700">6–Digit Verification Code</label>
                
                <div className="grid grid-cols-6 gap-2 sm:gap-3">
                  {codeDigits.map((digit, idx) => (
                    <input 
                      key={idx}
                      ref={(el) => { inputRefs.current[idx] = el; }}
                      type="text"
                      maxLength={1}
                      value={digit}
                      onChange={(e) => handleDigitChange(idx, e.target.value)}
                      className="w-full h-12 bg-[#f4f8f7] border border-slate-200 rounded-2xl text-center text-lg font-extrabold text-slate-900 focus:outline-none focus:border-[#008272] focus:bg-white shadow-sm transition"
                    />
                  ))}
                </div>

                {/* Verified Success Badge */}
                <div className="bg-[#f2f8f7] border border-[#cbe6e2] rounded-2xl p-3 flex items-center gap-2.5 text-xs text-emerald-900">
                  <CheckCircle2 className="w-4 h-4 text-[#008272] shrink-0" />
                  <div>
                    <strong className="block font-bold text-slate-900">Number verified</strong>
                    <span className="text-[11px] text-slate-500">Lakeview secondary directory check cleared</span>
                  </div>
                </div>
              </div>

            </div>
          )}

        </div>

        {/* Footer Navigation Buttons Bar */}
        <div className="bg-[#f8faf9] border-t border-slate-200 px-6 py-4 flex items-center justify-between">
          <button 
            onClick={() => router.back()}
            className="px-6 py-2.5 bg-white hover:bg-slate-100 border border-slate-200 font-semibold rounded-xl text-xs text-slate-700 transition shadow-sm"
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