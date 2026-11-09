'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  ShieldCheck, 
  Bell, 
  Lock, 
  Smartphone, 
  ArrowRight, 
  RefreshCw, 
  LogOut, 
  Globe
} from 'lucide-react';

export default function AccountSettingsPage() {
  const router = useRouter();
  const [newPhone, setNewPhone] = useState('');
  const [isChangingPhone, setIsChangingPhone] = useState(false);

  const handleSendNewCode = (e: React.FormEvent) => {
    e.preventDefault();
    router.push(`/verify?phone=${encodeURIComponent(newPhone)}`);
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white flex flex-col items-center p-4 md:p-8 text-slate-800">
      
      {/* Top Demo Banner */}
      <div className="text-xs font-medium text-slate-600 bg-white/80 backdrop-blur px-4 py-1.5 rounded-full border border-slate-200/60 shadow-sm mb-6">
        Demo environment • sample data only
      </div>

      {/* Main Container Wrapper */}
      <div className="w-full max-w-4xl bg-white/80 backdrop-blur-md border border-slate-200/90 rounded-[32px] p-6 md:p-10 shadow-2xl space-y-8">
        
        {/* Top Navbar Header */}
        <div className="flex items-center justify-between pb-6 border-b border-slate-200">
          <div className="flex items-center gap-8">
            <div className="flex items-center gap-2.5">
              <div className="w-9 h-9 rounded-xl bg-[#008272] flex items-center justify-center text-white font-bold shadow-md shadow-[#008272]/20">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <span className="font-bold text-slate-900 text-lg tracking-tight">Ikonex Academy</span>
            </div>

            {/* Desktop Navigation Links */}
            <nav className="hidden md:flex items-center gap-6 text-sm font-medium text-slate-600">
              <span className="px-3 py-1.5 bg-[#e0f2f1] text-[#008272] rounded-full font-semibold">Public Suite</span>
              <a href="#cbc" className="hover:text-slate-900 transition">CBC Strands</a>
              <a href="#resources" className="hover:text-slate-900 transition">Resources</a>
              <a href="#health" className="hover:text-slate-900 transition">System Health</a>
            </nav>
          </div>

          <div className="flex items-center gap-4">
            <button className="p-2.5 text-slate-600 hover:bg-slate-100 rounded-full transition relative">
              <Bell className="w-5 h-5 text-slate-700" />
              <span className="absolute top-2 right-2 w-2 h-2 rounded-full bg-rose-500"></span>
            </button>
            <div className="flex items-center gap-2.5 bg-slate-100 pl-3 pr-4 py-1.5 rounded-full border border-slate-200">
              <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center text-xs font-bold">
                GM
              </div>
              <div className="text-left leading-tight">
                <p className="text-xs font-bold text-slate-900">Grace Mwangi</p>
                <p className="text-[10px] text-slate-500">CBC Lead Admin</p>
              </div>
            </div>
          </div>
        </div>

        {/* Page Title & Subtitle */}
        <div className="space-y-1">
          <div className="flex items-center gap-1.5 text-xs font-semibold text-[#008272] uppercase tracking-wider">
            <ShieldCheck className="w-4 h-4" /> SECURITY & INSTITUTIONAL REGISTRY
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-slate-900">Account Settings</h1>
          <p className="text-slate-600 text-sm">
            Manage your institutional profile, phone number, and security settings.
          </p>
        </div>

        {/* User Profile Card */}
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-4">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <div className="w-14 h-14 rounded-2xl bg-[#008272] text-white font-bold text-xl flex items-center justify-center shadow-md shadow-[#008272]/20">
                GM
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <h2 className="text-xl font-bold text-slate-900">Grace Mwangi</h2>
                  <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-100 text-purple-700">
                    Parent
                  </span>
                </div>
                <p className="text-xs text-slate-500 font-medium">Lakeview Secondary</p>
              </div>
            </div>
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 border border-emerald-200 text-emerald-700 text-xs font-semibold rounded-full self-start md:self-auto">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              This device • Active session
            </div>
          </div>

          <div className="bg-[#f8faf9] border border-slate-200/80 rounded-xl p-3.5 text-xs text-slate-600 leading-relaxed">
            Official details are managed directly by Lakeview Secondary administration. To update legal guardianship or institutional identity records, visit the administration office.
          </div>
        </div>

        {/* Registered Number Management Card */}
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm space-y-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-sm font-bold text-slate-900">
              <Lock className="w-4 h-4 text-slate-500" /> Registered number
            </div>
            <Smartphone className="w-4 h-4 text-slate-400" />
          </div>

          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="space-y-1">
              <div className="text-xl font-mono font-bold text-slate-900 tracking-wide">
                +254 7•• ••• 678
              </div>
              <div className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-amber-50 border border-amber-200 text-amber-800 text-xs font-medium rounded-lg">
                <RefreshCw className="w-3 h-3 text-amber-600" /> Changed today • payments paused until 10:15 am
              </div>
            </div>

            <button 
              onClick={() => setIsChangingPhone(!isChangingPhone)}
              className="px-4 py-2 bg-purple-50 hover:bg-purple-100 border border-purple-200 text-purple-700 font-semibold rounded-xl text-xs transition flex items-center gap-1.5 self-start md:self-auto"
            >
              <Smartphone className="w-3.5 h-3.5" /> Change number
            </button>
          </div>

          <p className="text-xs text-slate-500">
            Codes for signing in and for the assistant are sent to this number.
          </p>

          {/* Change Number Step 1 of 3 Inline Flow */}
          {isChangingPhone && (
            <div className="bg-[#f4f8f7] border border-slate-200 rounded-2xl p-5 space-y-4 mt-4">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-[#008272] uppercase tracking-wider flex items-center gap-1">
                  <span className="w-2 h-2 rounded-full bg-[#008272]"></span> STEP 1 OF 3 • CHANGE YOUR NUMBER
                </span>
                <div className="flex gap-1">
                  <span className="w-2 h-2 rounded-full bg-[#008272]"></span>
                  <span className="w-2 h-2 rounded-full bg-slate-300"></span>
                  <span className="w-2 h-2 rounded-full bg-slate-300"></span>
                </div>
              </div>

              <p className="text-xs text-slate-600">
                We'll send a code to your new number. We'll also send a text to your old number so you know about the change.
              </p>

              <form onSubmit={handleSendNewCode} className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold uppercase tracking-wider text-slate-600 mb-1.5">
                    New mobile number
                  </label>
                  <div className="flex bg-white border border-slate-200 rounded-xl overflow-hidden focus-within:border-[#008272] transition">
                    <div className="flex items-center gap-1 px-3 bg-slate-50 border-r border-slate-200 text-slate-700 font-semibold text-sm">
                      <span className="text-[#008272]">KE</span>
                      <span>+254</span>
                    </div>
                    <input 
                      type="tel" 
                      required
                      value={newPhone}
                      onChange={(e) => setNewPhone(e.target.value)}
                      placeholder="712 345 678"
                      className="w-full bg-transparent px-4 py-3 text-slate-900 placeholder-slate-400 focus:outline-none text-sm font-mono"
                    />
                  </div>
                  <p className="text-[11px] text-slate-500 mt-1.5">
                    Enter Safaricom, Airtel, or Telkom line.
                  </p>
                </div>

                <button 
                  type="submit" 
                  className="w-full md:w-auto px-6 bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3 rounded-xl transition flex items-center justify-center gap-2 text-xs shadow-md shadow-[#008272]/20"
                >
                  Send code <ArrowRight className="w-4 h-4" />
                </button>
              </form>
            </div>
          )}

          <div className="text-[11px] text-slate-400 pt-2 border-t border-slate-100">
            Can't receive texts on your current number? Visit the school office for help.
          </div>
        </div>

        {/* Switch Role or School Card */}
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2 font-bold text-slate-900 text-base">
              🔄 Switch role or school
            </div>
            <p className="text-xs text-slate-500">
              Currently active as Grace Mwangi at Lakeview.
            </p>
          </div>
          <button 
            onClick={() => router.push('/select-role')}
            className="px-4 py-2 bg-white hover:bg-slate-50 border border-slate-200 text-slate-700 font-semibold rounded-xl text-xs transition shadow-sm"
          >
            Choose another role
          </button>
        </div>

        {/* Privacy & Consent Card */}
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2 font-bold text-slate-900 text-base">
              <Lock className="w-4 h-4 text-slate-500" /> Privacy & consent
            </div>
            <p className="text-xs text-slate-500">
              You'll be able to see and change what the school may do with your family's information.
            </p>
          </div>
          <span className="px-3 py-1 bg-slate-100 text-slate-600 font-semibold rounded-full text-[11px] self-start md:self-auto">
            COMING SOON
          </span>
        </div>

        {/* Language & Regional Settings Card */}
        <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-slate-100 rounded-xl text-slate-700">
              <Globe className="w-5 h-5" />
            </div>
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <h3 className="font-bold text-slate-900 text-base">English</h3>
                <span className="px-2 py-0.5 rounded-md text-xs font-semibold bg-[#e0f2f1] text-[#008272]">
                  Default
                </span>
              </div>
              <p className="text-xs text-slate-500">More languages are coming.</p>
            </div>
          </div>
          <Lock className="w-4 h-4 text-slate-400" />
        </div>

        {/* Center Sign Out Button & Session ID Footer */}
        <div className="pt-6 flex flex-col items-center justify-center space-y-3">
          <button 
            onClick={() => router.push('/login')}
            className="px-6 py-3 bg-white hover:bg-rose-50 border border-slate-200 hover:border-rose-200 text-rose-600 font-semibold rounded-2xl transition shadow-sm flex items-center gap-2 text-xs"
          >
            <LogOut className="w-4 h-4" /> Sign out
          </button>
          <p className="text-[11px] font-mono text-slate-400">
            Session ID: IKX-49201–KE · Lakeview
          </p>
        </div>

      </div>

    </div>
  );
}