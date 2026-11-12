'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Home, 
  Settings, 
  Users, 
  GraduationCap, 
  BookOpen, 
  Clock, 
  Calendar as CalendarIcon, 
  Bell, 
  Search, 
  ShieldCheck, 
  LogOut, 
  Loader2, 
  AlertCircle, 
  QrCode, 
  Printer, 
  Copy, 
  Download, 
  XCircle, 
  ArrowRight, 
  Filter 
} from 'lucide-react';

interface ActiveQRCodeData {
  codeString: string;
  timeRemaining: string;
  issuedAt: string;
  expiresAt: string;
  createdBy: string;
  telemetry: {
    submitted: number;
    verified: number;
    waiting: number;
    throughputPercentage: number;
    clearedCount: number;
    pendingCount: number;
  };
  auditRegistry: Array<{
    id: string;
    created: string;
    createdBy: string;
    ended: string;
    status: 'Expired' | 'Revoked' | 'Active';
    registeredCount: number;
  }>;
}

export default function AdminParentsQRPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('parents');
  const [subTab, setSubTab] = useState<'qr' | 'queue'>('qr');

  // Dynamic state replacing hardcoded mocks
  const [qrData, setQrData] = useState<ActiveQRCodeData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch QR code telemetry and audit registry from backend API
  useEffect(() => {
    const fetchQRTelemetry = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch('/api/v1/admin/parents/qr-session', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize QR session telemetry from institutional registry.');
        }

        const data = await res.json();
        setQrData(data);
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading QR session data.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchQRTelemetry();
  }, []);

  // Handle Revoke QR Code Action
  const handleRevokeQR = async () => {
    if (!confirm('Are you sure you want to revoke this active parent QR code session?')) return;
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/admin/parents/qr-session/revoke', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) throw new Error('Failed to revoke QR code session.');

      alert('Active QR session successfully revoked.');
      window.location.reload();
    } catch (err: any) {
      alert(err.message || 'Revocation failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCopyLink = () => {
    if (qrData) {
      navigator.clipboard.writeText(`https://ikonex.ac.ke/onboard?pass=${qrData.codeString}`);
      alert('Secure induction pass link copied to clipboard.');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#eaf4f3] via-[#f2f7f6] to-white flex text-slate-800">
      
      {/* Left Institutional Sidebar */}
      <aside className="w-64 bg-white border-r border-slate-200/90 flex flex-col justify-between p-5 shrink-0 hidden md:flex shadow-sm">
        <div className="space-y-6">
          <div>
            <h1 className="font-bold text-slate-900 text-lg tracking-tight">Ikonex Academy</h1>
            <p className="text-xs text-slate-500 font-medium">Lakeview Secondary</p>
          </div>

          <nav className="space-y-1 text-sm font-medium text-slate-600">
            <button onClick={() => router.push('/admin')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Home className="w-4 h-4 text-slate-400" /> Home
            </button>
            <button onClick={() => router.push('/admin/setup')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Settings className="w-4 h-4 text-slate-400" /> Setup
            </button>
            <button onClick={() => router.push('/admin/staff')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Users className="w-4 h-4 text-slate-400" /> Staff
            </button>
            <button onClick={() => setActiveNav('parents')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <Users className="w-4 h-4 text-[#008272]" /> Parents
            </button>
            <button onClick={() => router.push('/admin/teaching')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <BookOpen className="w-4 h-4 text-slate-400" /> Teaching
            </button>
            <button onClick={() => router.push('/admin/timetables')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Clock className="w-4 h-4 text-slate-400" /> Timetables
            </button>
            <button onClick={() => router.push('/admin/calendar')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <CalendarIcon className="w-4 h-4 text-slate-400" /> Calendar
            </button>
            <button onClick={() => router.push('/admin/students')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <GraduationCap className="w-4 h-4 text-slate-400" /> Students
            </button>
            <button onClick={() => router.push('/notifications')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Bell className="w-4 h-4 text-slate-400" /> Notifications
            </button>
          </nav>
        </div>

        <div className="pt-4 border-t border-slate-200 space-y-3">
          <div className="flex items-center gap-3 bg-slate-50 p-2.5 rounded-2xl border border-slate-200">
            <div className="w-8 h-8 rounded-xl bg-[#008272] text-white flex items-center justify-center text-xs font-bold">
              DA
            </div>
            <div className="leading-tight">
              <p className="text-xs font-bold text-slate-900">Ms. Achieng</p>
              <p className="text-[10px] text-slate-500 font-medium">Principal</p>
            </div>
          </div>
          <button onClick={() => router.push('/login')} className="w-full flex items-center gap-2 text-xs font-semibold text-rose-600 hover:text-rose-700 px-2 transition">
            <LogOut className="w-3.5 h-3.5" /> Sign out
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-w-0 overflow-y-auto">
        
        {/* Top Header Bar */}
        <div className="w-full bg-white/60 backdrop-blur border-b border-slate-200 px-6 py-2.5 flex items-center justify-between text-xs text-slate-500">
          <span className="px-3 py-1 bg-purple-50 text-purple-700 rounded-full font-medium border border-purple-200/60">
            Principal • Ms. Achieng
          </span>
          <span className="bg-white/80 px-3 py-1 rounded-full border border-slate-200 font-medium">
            Demo environment • sample data only
          </span>
          <div className="flex items-center gap-4">
            <button className="p-1.5 hover:bg-slate-100 rounded-full text-slate-600 transition"><Search className="w-4 h-4" /></button>
            <button className="p-1.5 hover:bg-slate-100 rounded-full text-slate-600 transition relative">
              <Bell className="w-4 h-4" />
              <span className="absolute top-1 right-1 w-2 h-2 rounded-full bg-rose-500"></span>
            </button>
            <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center text-xs font-bold">
              DA
            </div>
          </div>
        </div>

        <div className="p-6 md:p-10 space-y-6 max-w-7xl w-full mx-auto">
          
          {/* Top Subheader & Status Bar */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-2 border-b border-slate-200">
            <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">Parents</h2>
            <div className="flex items-center gap-3">
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs font-semibold rounded-full">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span> Session Active
              </span>
              <span className="text-xs text-slate-500 bg-slate-100 px-3 py-1.5 rounded-full font-medium">
                Term 3 · Onboarding Window
              </span>
            </div>
          </div>

          {/* Sub Navigation Switcher */}
          <div className="flex items-center gap-4 border-b border-slate-200 pb-2 text-xs font-semibold">
            <button 
              onClick={() => setSubTab('qr')}
              className={`flex items-center gap-1.5 pb-2 transition ${subTab === 'qr' ? 'text-[#008272] border-b-2 border-[#008272]' : 'text-slate-500 hover:text-slate-900'}`}
            >
              <QrCode className="w-4 h-4" /> QR code
            </button>
            <button 
              onClick={() => {
                setSubTab('queue');
                router.push('/admin/parents');
              }}
              className="flex items-center gap-1.5 pb-2 text-slate-500 hover:text-slate-900 transition"
            >
              <Users className="w-4 h-4" /> Verification queue <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded-full text-[10px]">8</span>
            </button>
          </div>

          {/* Error Banner if API fails */}
          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {isLoading ? (
            <div className="py-48 flex flex-col items-center justify-center gap-3 text-slate-400">
              <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
              <p className="text-xs font-medium">Loading active QR session telemetry...</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
              
              {/* Left Column: Active Parent QR Code Card */}
              <div className="lg:col-span-6 bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm space-y-6">
                
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-bold text-slate-900 text-base">Active parent QR code</h3>
                    <p className="text-xs text-slate-500">Official induction pass for guardian onboarding</p>
                  </div>
                  <span className="px-2.5 py-1 bg-emerald-50 text-[#008272] border border-emerald-200 font-bold text-[10px] rounded-full flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-[#008272] animate-ping"></span> Live
                  </span>
                </div>

                {/* QR Code Container Box */}
                <div className="bg-gradient-to-b from-[#f8faf9] to-[#edf4f3] border border-slate-200/90 rounded-3xl p-8 flex flex-col items-center justify-center space-y-5">
                  <div className="w-56 h-56 bg-white border-4 border-slate-900/10 rounded-2xl p-4 shadow-md flex items-center justify-center">
                    {/* Simulated SVG QR Matrix */}
                    <div className="w-full h-full bg-[#008272]/10 rounded-xl flex items-center justify-center relative overflow-hidden">
                      <QrCode className="w-32 h-32 text-[#008272]" />
                    </div>
                  </div>

                  <div className="w-full space-y-2">
                    <div className="flex items-center justify-between text-xs font-bold">
                      <span className="text-slate-500">Time remaining</span>
                      <span className="text-[#008272]">Valid for {qrData?.timeRemaining || '7 h 12 min'}</span>
                    </div>
                    <div className="w-full bg-slate-200 h-2 rounded-full overflow-hidden">
                      <div className="bg-[#008272] h-full rounded-full w-3/4"></div>
                    </div>
                    <div className="flex items-center justify-between text-[11px] text-slate-400">
                      <span>Issued: {qrData?.issuedAt || '6:00 am'}</span>
                      <span>Expires: {qrData?.expiresAt || '6:00 pm (12h window)'}</span>
                    </div>
                  </div>
                </div>

                <div className="bg-slate-50 border border-slate-200 rounded-2xl p-3 flex items-center justify-between text-xs text-slate-600">
                  <span>Created by <strong className="text-slate-900">{qrData?.createdBy || 'Mr. Otieno (Dean)'}</strong> · 6:00 am today</span>
                </div>

                {/* Action Buttons */}
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs font-semibold">
                  <button className="py-2.5 bg-white hover:bg-slate-5