'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  ShieldCheck, 
  Bell, 
  AlertTriangle, 
  Compass, 
  Lock, 
  Clock, 
  WifiOff, 
  QrCode, 
  ArrowRight, 
  X, 
  Sliders, 
  FileText, 
  Palette, 
  Code2 
} from 'lucide-react';

export default function SystemHealthPage() {
  const router = useRouter();
  const [activeFrame, setActiveFrame] = useState<'404' | '403' | 'timeout' | 'offline' | 'qr'>('404');
  const [isOfflineBannerVisible, setIsOfflineBannerVisible] = useState(true);

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white flex flex-col items-center p-4 md:p-8 text-slate-800">
      
      {/* Top Demo Banner */}
      <div className="text-xs font-medium text-slate-600 bg-white/80 backdrop-blur px-4 py-1.5 rounded-full border border-slate-200/60 shadow-sm mb-4">
        Demo environment • sample data only
      </div>

      {/* Main Container Wrapper */}
      <div className="w-full max-w-6xl bg-white/90 backdrop-blur-md border border-slate-200/90 rounded-[32px] p-6 md:p-8 shadow-2xl space-y-6">
        
        {/* Top Navbar Header */}
        <div className="flex items-center justify-between pb-5 border-b border-slate-200">
          <div className="flex items-center gap-8">
            <div className="flex items-center gap-2.5">
              <div className="w-9 h-9 rounded-xl bg-[#008272] flex items-center justify-center text-white font-bold shadow-md shadow-[#008272]/20">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <span className="font-bold text-slate-900 text-lg tracking-tight">Ikonex Academy</span>
            </div>

            <nav className="hidden md:flex items-center gap-6 text-sm font-medium text-slate-600">
              <a href="#public" className="hover:text-slate-900 transition">Public Suite</a>
              <a href="#cbc" className="hover:text-slate-900 transition">CBC Strands</a>
              <a href="#resources" className="hover:text-slate-900 transition">Resources</a>
              <span className="px-3 py-1.5 bg-[#e0f2f1] text-[#008272] rounded-full font-semibold">System Health</span>
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

        {/* Offline Warning Banner */}
        {isOfflineBannerVisible && (
          <div className="bg-amber-50 border border-amber-200 text-amber-900 px-4 py-3 rounded-2xl flex items-center justify-between text-xs font-medium shadow-sm">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
              <span>You are offline. You can read pages you have already opened. Payments and changes need a connection.</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-[11px] font-semibold text-amber-700 hidden sm:inline">Live Sync Paused</span>
              <button 
                onClick={() => setIsOfflineBannerVisible(false)}
                className="p-1 hover:bg-amber-100 rounded-lg text-amber-700 transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* System States Inspector Header & Frame Tabs */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pt-2">
          <div className="flex items-center gap-2.5">
            <div className="p-2.5 bg-slate-100 rounded-xl text-slate-700">
              <Sliders className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-slate-900">S6: System States Inspector</h2>
              <p className="text-xs text-slate-500">Standard 460px state cards with atmospheric radial mesh canvas</p>
            </div>
          </div>

          {/* Frame Switcher Pills */}
          <div className="flex flex-wrap items-center gap-1.5 bg-slate-100 p-1.5 rounded-2xl border border-slate-200 text-xs font-semibold">
            <button 
              onClick={() => setActiveFrame('404')}
              className={`px-3 py-1.5 rounded-xl transition ${activeFrame === '404' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Frame 1 • 404
            </button>
            <button 
              onClick={() => setActiveFrame('403')}
              className={`px-3 py-1.5 rounded-xl transition ${activeFrame === '403' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Frame 2 • 403
            </button>
            <button 
              onClick={() => setActiveFrame('timeout')}
              className={`px-3 py-1.5 rounded-xl transition ${activeFrame === 'timeout' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Frame 3 • Session
            </button>
            <button 
              onClick={() => setActiveFrame('offline')}
              className={`px-3 py-1.5 rounded-xl transition ${activeFrame === 'offline' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Frame 4 • Offline
            </button>
            <button 
              onClick={() => setActiveFrame('qr')}
              className={`px-3 py-1.5 rounded-xl transition ${activeFrame === 'qr' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
            >
              Frame 5 • QR Link
            </button>
          </div>
        </div>

        {/* Active Frame Preview Canvas */}
        <div className="relative bg-gradient-to-b from-[#eaf4f3] via-[#f0f6f5] to-[#f8faf9] border border-slate-200/90 rounded-[28px] p-8 md:p-14 flex flex-col items-center justify-center overflow-hidden min-h-[500px]">
          
          {activeFrame === '404' && (
            <div className="w-full max-w-[460px] bg-white border border-slate-200/90 rounded-[28px] p-8 shadow-xl text-center space-y-6">
              <div className="w-14 h-14 bg-purple-100 rounded-2xl text-purple-700 flex items-center justify-center mx-auto shadow-inner">
                <Compass className="w-7 h-7" />
              </div>
              <div className="space-y-2">
                <h3 className="text-2xl font-bold tracking-tight text-slate-900">We can&apos;t find that page</h3>
                <p className="text-xs text-slate-600 leading-relaxed max-w-xs mx-auto">
                  The link may be old or mistyped. Check the URL or return to your student dashboard.
                </p>
              </div>
              <div className="space-y-3 pt-2">
                <button 
                  onClick={() => router.push('/dashboard')}
                  className="w-full bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3.5 rounded-2xl transition flex items-center justify-center gap-2 shadow-lg shadow-[#008272]/20 text-sm"
                >
                  Go to home <ArrowRight className="w-4 h-4" />
                </button>
                <button 
                  onClick={() => router.push('/login')}
                  className="text-xs font-semibold text-[#008272] hover:underline"
                >
                  Sign in again
                </button>
              </div>
              <div className="pt-2">
                <span className="inline-block px-3 py-1 bg-slate-100 text-slate-500 rounded-full text-[11px] font-mono">
                  • Error 404 • route not registered
                </span>
              </div>
            </div>
          )}

          {activeFrame === '403' && (
            <div className="w-full max-w-[460px] bg-white border border-slate-200/90 rounded-[28px] p-8 shadow-xl text-center space-y-6">
              <div className="w-14 h-14 bg-rose-100 rounded-2xl text-rose-700 flex items-center justify-center mx-auto shadow-inner">
                <Lock className="w-7 h-7" />
              </div>
              <div className="space-y-2">
                <h3 className="text-2xl font-bold tracking-tight text-slate-900">403 Forbidden</h3>
                <p className="text-xs text-slate-600 leading-relaxed max-w-xs mx-auto">
                  Requires Guardian / CBC Lead elevated credentials to access this registry zone.
                </p>
              </div>
              <button 
                onClick={() => router.push('/select-role')}
                className="w-full bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3.5 rounded-2xl transition text-sm shadow-md"
              >
                Switch Role / Request Access
              </button>
              <span className="inline-block px-3 py-1 bg-rose-50 text-rose-700 rounded-full text-[11px] font-mono">
                • Error 403 • insufficient permissions
              </span>
            </div>
          )}

          {activeFrame === 'timeout' && (
            <div className="w-full max-w-[460px] bg-white border border-slate-200/90 rounded-[28px] p-8 shadow-xl text-center space-y-6">
              <div className="w-14 h-14 bg-amber-100 rounded-2xl text-amber-700 flex items-center justify-center mx-auto shadow-inner">
                <Clock className="w-7 h-7" />
              </div>
              <div className="space-y-2">
                <h3 className="text-2xl font-bold tracking-tight text-slate-900">Session Timeout</h3>
                <p className="text-xs text-slate-600 leading-relaxed max-w-xs mx-auto">
                  Auto-logout protection triggered for unattended classroom terminals.
                </p>
              </div>
              <button 
                onClick={() => router.push('/login')}
                className="w-full bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3.5 rounded-2xl transition text-sm shadow-md"
              >
                Reconnect Session
              </button>
              <span className="inline-block px-3 py-1 bg-amber-50 text-amber-800 rounded-full text-[11px] font-mono">
                • Code 403 • session expired
              </span>
            </div>
          )}

          {activeFrame === 'offline' && (
            <div className="w-full max-w-[460px] bg-white border border-slate-200/90 rounded-[28px] p-8 shadow-xl text-center space-y-6">
              <div className="w-14 h-14 bg-slate-100 rounded-2xl text-slate-700 flex items-center justify-center mx-auto shadow-inner">
                <WifiOff className="w-7 h-7" />
              </div>
              <div className="space-y-2">
                <h3 className="text-2xl font-bold tracking-tight text-slate-900">Network Disconnect</h3>
                <p className="text-xs text-slate-600 leading-relaxed max-w-xs mx-auto">
                  Full offline fallback active with clear degraded capability notice.
                </p>
              </div>
              <button 
                onClick={() => window.location.reload()}
                className="w-full bg-slate-800 hover:bg-slate-900 text-white font-semibold py-3.5 rounded-2xl transition text-sm shadow-md"
              >
                Retry Connection
              </button>
              <span className="inline-block px-3 py-1 bg-slate-100 text-slate-600 rounded-full text-[11px] font-mono">
                • Offline Mode Active
              </span>
            </div>
          )}

          {activeFrame === 'qr' && (
            <div className="w-full max-w-[460px] bg-white border border-slate-200/90 rounded-[28px] p-8 shadow-xl text-center space-y-6">
              <div className="w-14 h-14 bg-purple-100 rounded-2xl text-purple-700 flex items-center justify-center mx-auto shadow-inner">
                <QrCode className="w-7 h-7" />
              </div>
              <div className="space-y-2">
                <h3 className="text-2xl font-bold tracking-tight text-slate-900">Invalid QR Token</h3>
                <p className="text-xs text-slate-600 leading-relaxed max-w-xs mx-auto">
                  School desk context card and direct telephone fallback required.
                </p>
              </div>
              <button 
                onClick={() => router.push('/login')}
                className="w-full bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3.5 rounded-2xl transition text-sm shadow-md"
              >
                Use Phone Sign-In
              </button>
              <span className="inline-block px-3 py-1 bg-purple-50 text-purple-700 rounded-full text-[11px] font-mono">
                • Token Validation Error
              </span>
            </div>
          )}

        </div>

        {/* Design System State Matrix Section */}
        <div className="space-y-4 pt-2">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-2">
            <div>
              <h3 className="text-lg font-bold text-slate-900">Design System State Matrix</h3>
              <p className="text-xs text-slate-500">Simultaneous comparative overview of desktop cards (440–480px nominal)</p>
            </div>
            <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-xs font-semibold self-start md:self-auto">
              Token Compliant • High Trust Mode
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div 
              onClick={() => setActiveFrame('403')}
              className="bg-white border border-slate-200 hover:border-[#008272] rounded-2xl p-5 shadow-sm space-y-3 cursor-pointer transition group"
            >
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-bold text-slate-400 uppercase">Frame 2</span>
                <div className="p-2 bg-rose-50 text-rose-600 rounded-xl group-hover:bg-rose-100 transition">
                  <Lock className="w-4 h-4" />
                </div>
              </div>
              <div>
                <h4 className="font-bold text-slate-900 text-sm">403 Forbidden</h4>
                <p className="text-xs text-slate-500 mt-1 leading-relaxed">Requires Guardian / CBC Lead elevated credentials.</p>
              </div>
              <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-xs font-semibold text-[#008272]">
                <span>Inspect Frame</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition" />
              </div>
            </div>

            <div 
              onClick={() => setActiveFrame('timeout')}
              className="bg-white border border-slate-200 hover:border-[#008272] rounded-2xl p-5 shadow-sm space-y-3 cursor-pointer transition group"
            >
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-bold text-slate-400 uppercase">Frame 3</span>
                <div className="p-2 bg-amber-50 text-amber-600 rounded-xl group-hover:bg-amber-100 transition">
                  <Clock className="w-4 h-4" />
                </div>
              </div>
              <div>
                <h4 className="font-bold text-slate-900 text-sm">Session Timeout</h4>
                <p className="text-xs text-slate-500 mt-1 leading-relaxed">Auto-logout protection for unattended classroom terminals.</p>
              </div>
              <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-xs font-semibold text-[#008272]">
                <span>Inspect Frame</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition" />
              </div>
            </div>

            <div 
              onClick={() => setActiveFrame('offline')}
              className="bg-white border border-slate-200 hover:border-[#008272] rounded-2xl p-5 shadow-sm space-y-3 cursor-pointer transition group"
            >
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-bold text-slate-400 uppercase">Frame 4</span>
                <div className="p-2 bg-purple-50 text-purple-600 rounded-xl group-hover:bg-purple-100 transition">
                  <WifiOff className="w-4 h-4" />
                </div>
              </div>
              <div>
                <h4 className="font-bold text-slate-900 text-sm">Network Disconnect</h4>
                <p className="text-xs text-slate-500 mt-1 leading-relaxed">Full offline fallback with clear degraded capability notice.</p>
              </div>
              <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-xs font-semibold text-[#008272]">
                <span>Inspect Frame</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition" />
              </div>
            </div>

            <div 
              onClick={() => setActiveFrame('qr')}
              className="bg-white border border-slate-200 hover:border-[#008272] rounded-2xl p-5 shadow-sm space-y-3 cursor-pointer transition group"
            >
              <div className="flex items-center justify-between">
                <span className="text-[11px] font-bold text-slate-400 uppercase">Frame 5</span>
                <div className="p-2 bg-emerald-50 text-emerald-600 rounded-xl group-hover:bg-emerald-100 transition">
                  <QrCode className="w-4 h-4" />
                </div>
              </div>
              <div>
                <h4 className="font-bold text-slate-900 text-sm">Invalid QR Token</h4>
                <p className="text-xs text-slate-500 mt-1 leading-relaxed">School desk context card and direct telephone fallback.</p>
              </div>
              <div className="pt-2 border-t border-slate-100 flex items-center justify-between text-xs font-semibold text-[#008272]">
                <span>Inspect Frame</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition" />
              </div>
            </div>
          </div>
        </div>

        {/* Implementation Specifications Footer Box */}
        <div className="bg-[#f2f8f7] border border-slate-200 rounded-2xl p-6 grid grid-cols-1 md:grid-cols-3 gap-6 text-xs text-slate-700">
          <div className="space-y-1">
            <h4 className="font-bold text-slate-900 flex items-center gap-1.5">
              <FileText className="w-4 h-4 text-[#008272]" /> Canvas Geometry
            </h4>
            <p className="text-slate-600 leading-relaxed">
              Card container constrained between 440px and 480px width (rendered at 460px optimal). Centered with 96px top breathing margin and 24px border radius.
            </p>
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-slate-900 flex items-center gap-1.5">
              <Palette className="w-4 h-4 text-[#008272]" /> Color Palette Compliance
            </h4>
            <p className="text-slate-600 leading-relaxed">
              Strict 60/30/10 ratio: 60% Atmospheric Canvas (#F2F6F6 with soft gradient wash), 30% Teal (#008080/#006565 buttons & links), 10% Lavender (#EEEAF8/#5B4F9B icon badge).
            </p>
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-slate-900 flex items-center gap-1.5">
              <Code2 className="w-4 h-4 text-[#008272]" /> QA Developer Metadata
            </h4>
            <p className="text-slate-600 leading-relaxed">
              Subtle footer tag provides error code and system state token to accelerate ticketing without alarming guardians or students.
            </p>
          </div>
        </div>

      </div>

    </div>
  );
}