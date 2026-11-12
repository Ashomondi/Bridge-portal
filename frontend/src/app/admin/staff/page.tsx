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
  CheckCircle2, 
  Eye, 
  FileText 
} from 'lucide-react';

interface StaffApplication {
  id: string;
  fullName: string;
  role: string;
  tscNumber: string;
  subjects: string;
  appliedTime: string;
  formsLead: string;
  status: 'Pending' | 'Approved' | 'Rejected';
  credentials: Array<{
    id: string;
    name: string;
    filename: string;
    size: string;
    type: string;
  }>;
}

interface StaffQRData {
  token: string;
  timeRemaining: string;
  dailyValidityWindow: string;
  percentageRemaining: number;
  windowSettings: {
    schedule: string;
    autoResets: boolean;
    tlsSecured: boolean;
    status: 'Active' | 'Closed';
  };
  applications: StaffApplication[];
}

export default function AdminStaffPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('staff');
  const [activeTab, setActiveTab] = useState<'Pending' | 'Approved' | 'Rejected'>('Pending');

  // Dynamic state replacing hardcoded mocks
  const [staffData, setStaffData] = useState<StaffQRData | null>(null);
  const [selectedApplicant, setSelectedApplicant] = useState<StaffApplication | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // In-person verification checklist state
  const [verificationChecked, setVerificationChecked] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch staff QR and application queue from backend API
  useEffect(() => {
    const fetchStaffData = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch(`/api/v1/admin/staff/applications?status=${activeTab}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize staff onboarding queue from institutional registry.');
        }

        const data = await res.json();
        setStaffData(data);

        // Automatically select the first applicant if available
        if (data.applications && data.applications.length > 0 && (!selectedApplicant || !data.applications.find((a: StaffApplication) => a.id === selectedApplicant.id))) {
          setSelectedApplicant(data.applications[0]);
          setVerificationChecked(false);
        } else if (data.applications.length === 0) {
          setSelectedApplicant(null);
        }
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading staff applications.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchStaffData();
  }, [activeTab]);

  // Handle Approve or Reject staff application
  const handleApprovalAction = async (status: 'Approved' | 'Rejected') => {
    if (!selectedApplicant) return;
    setIsSubmitting(true);
    try {
      const res = await fetch(`/api/v1/admin/staff/applications/${selectedApplicant.id}/review`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status,
          inPersonVerified: verificationChecked,
        }),
      });

      if (!res.ok) throw new Error('Failed to record staff review decision.');

      // Remove from current queue
      if (staffData) {
        const remaining = staffData.applications.filter(a => a.id !== selectedApplicant.id);
        setStaffData({
          ...staffData,
          applications: remaining,
        });
        setSelectedApplicant(remaining.length > 0 ? remaining[0] : null);
        setVerificationChecked(false);
      }

      alert(`Staff application successfully marked as ${status}.`);
    } catch (err: any) {
      alert(err.message || 'Action failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRevokePass = async () => {
    if (!confirm('Revoke active staff onboarding pass?')) return;
    try {
      const res = await fetch('/api/v1/admin/staff/qr-session/revoke', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });
      if (!res.ok) throw new Error('Revocation failed.');
      alert('Staff QR session successfully revoked.');
    } catch (err: any) {
      alert(err.message);
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
            <button onClick={() => setActiveNav('staff')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <Users className="w-4 h-4 text-[#008272]" /> Staff
            </button>
            <button onClick={() => router.push('/admin/parents')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Users className="w-4 h-4 text-slate-400" /> Parents
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
          
          {/* Top Subheader & Dual Authority Bar */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-2 border-b border-slate-200">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                <span>Executive Access</span> • <span>Lakeview Secondary · Term 1</span>
              </div>
              <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">Staff</h2>
            </div>
            <span className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-purple-50 border border-purple-200 text-purple-800 text-xs font-semibold rounded-full self-start md:self-auto">
              <ShieldCheck className="w-3.5 h-3.5 text-purple-600" /> Dual Authority: Principal & Deputy Only
            </span>
          </div>

          {/* Error Banner if API fails */}
          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {isLoading ? (
            <div className="bg-white border border-slate-200 rounded-[28px] py-48 flex flex-col items-center justify-center gap-3 text-slate-400">
              <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
              <p className="text-xs font-medium">Loading staff onboarding portal...</p>
            </div>
          ) : (
            <div className="space-y-6">
              
              {/* Top Row: Staff Onboarding Pass (Left) + Window Settings (Right) */}
              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                
                {/* Left Column: QR Pass */}
                <div className="lg:col-span-7 bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm flex flex-col justify-between space-y-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-bold text-slate-900 text-base">Staff Onboarding Pass</h3>
                      <p className="text-xs text-slate-500">Instruct incoming teachers to scan via phone camera or official portal link to submit certified bio-data.</p>
                    </div>
                    <span className="px-2.5 py-1 bg-emerald-50 text-[#008272] border border-emerald-200 font-bold text-[10px] rounded-full flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-[#008272] animate-ping"></span> Valid 8:00 am to 6:00 pm today
                    </span>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-center bg-[#f8faf9] border border-slate-200/80 rounded-3xl p-6">
                    <div className="md:col-span-5 flex flex-col items-center justify-center">
                      <div className="w-36 h-36 bg-white border-2 border-slate-900/10 rounded-2xl p-3 shadow-sm flex items-center justify-center">
                        <QrCode className="w-28 h-28 text-[#008272]" />
                      </div>
                      <span className="text-[10px] font-mono text-slate-400 mt-2 font-bold">TOKEN: STF-2025–09A</span>
                    </div>

                    <div className="md:col-span-7 space-y-4 text-xs">
                      <div>
                        <div className="flex items-center justify-between font-bold text-slate-700 mb-1">
                          <span>Daily validity window</span>
                          <span className="text-[#008272]">{staffData?.percentageRemaining || 62}% remaining</span>
                        </div>
                        <div className="w-full bg-slate-200 h-2 rounded-full overflow-hidden">
                          <div className="bg-[#008272] h-full rounded-full" style={{ width: `${staffData?.percentageRemaining || 62}%` }}></div>
                        </div>
                        <p className="text-[10px] text-slate-400 mt-1">- 7 h 12 min left</p>
                      </div>

                      <div className="flex flex-wrap items-center gap-2">
                        <button onClick={() => navigator.clipboard.writeText('https://ikonex.ac.ke/staff-onboard')} className="px-3.5 py-2 bg-white hover:bg-slate-50 border border-slate-200 text-slate-700 font-semibold rounded-xl transition flex items-center gap-1.5 shadow-sm">
                          <Copy className="w-3.5 h-3.5 text-slate-500" /> Copy link
                        </button>
                        <button className="px-3.5 py-2 bg-white hover:bg-slate-50 border border-slate-200 text-slate-700 font-semibold rounded-xl transition flex items-center gap-1.5 shadow-sm">
                          <Download className="w-3.5 h-3.5 text-slate-500" /> Download SVG
                        </button>
                        <button onClick={handleRevokePass} className="px-3.5 py-2 bg-rose-50 hover:bg-rose-100 border border-rose-200 text-rose-700 font-semibold rounded-xl transition flex items-center gap-1.5">
                          <XCircle className="w-3.5 h-3.5" /> Revoke
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Right Column: Window Settings Card */}
                <div className="lg:col-span-5 bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm space-y-6 flex flex-col justify-between">
                  <div className="flex items-center justify-between">
                    <h3 className="font-bold text-slate-900 text-base">Window settings</h3>
                    <span className="px-2.5 py-1 bg-emerald-50 text-emerald-800 text-[10px] font-bold rounded-full">Active</span>
                  </div>

                  <div className="bg-slate-50 border border-slate-200 rounded-2xl p-4 space-y-2 text-xs">
                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Standard Policy Schedule</span>
                    <p className="font-bold text-slate-900 text-sm">Valid 8:00 am to 6:00 pm every day it is active</p>
                    <div className="flex items-center gap-4 text-[11px] text-slate-500 pt-1">
                      <span>🔄 Auto-resets daily</span>
                      <span>🔒 TLS 1.3 secured</span>
                    </div>
                  </div>

                  <div className="bg-amber-50 border border-amber-200 text-amber-900 p-3.5 rounded-2xl text-[11px] leading-relaxed">
                    ⚠️ The window closes automatically at 6:00 pm. Only the principal or deputy can issue or modify this window.
                  </div>

                  <button className="w-full bg-slate-100 hover:bg-slate-200 text-slate-800 font-semibold py-3 rounded-xl transition text-xs">
                    Change window
                  </button>
                </div>

              </div>

              {/* Bottom Row: Staff Application Queue & Inspector */}
              <div className="bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm space-y-6">
                
                {/* Status Tabs */}
                <div className="flex items-center gap-3 border-b border-slate-200 pb-3 text-xs font-semibold">
                  <button 
                    onClick={() => setActiveTab('Pending')}
                    className={`px-4 py-2 rounded-xl transition ${activeTab === 'Pending' ? 'bg-[#008272] text-white shadow-sm' : 'bg-slate-100 text-slate-600'}`}
                  >
                    Pending ({staffData?.applications.filter(a => a.status === 'Pending').length || 2})
                  </button>
                  <button 
                    onClick={() => setActiveTab('Approved')}
                    className={`px-4 py-2 rounded-xl transition ${activeTab === 'Approved' ? 'bg-[#008272] text-white shadow-sm' : 'bg-slate-100 text-slate-600'}`}
                  >
                    Approved
                  </button>
                  <button 
                    onClick={() => setActiveTab('Rejected')}
                    className={`px-4 py-2 rounded-xl transition ${activeTab === 'Rejected' ? 'bg-[#008272] text-white shadow-sm' : 'bg-slate-100 text-slate-600'}`}
                  >
                    Rejected
                  </button>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                  
                  {/* Left Column: Applications Table */}
                  <div className="lg:col-span-7 overflow-x-auto">
                    <table className="w-full text-left border-collapse text-xs">
                      <thead>
                        <tr className="border-b border-slate-100 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                          <th className="py-3 px-3">Name</th>
                          <th className="py-3 px-3">Role</th>
                          <th className="py-3 px-3">Subjects</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100 font-medium text-slate-700">
                        {staffData?.applications?.map((app) => {
                          const isSelected = selectedApplicant?.id === app.id;
                          return (
                            <tr 
                              key={app.id} 
                              onClick={() => {
                                setSelectedApplicant(app);
                                setVerificationChecked(false);
                              }}
                              className={`cursor-pointer transition hover:bg-slate-50 ${isSelected ? 'bg-[#eaf4f3]/60' : ''}`}
                            >
                              <td className="py-4 px-3">
                                <div className="flex items-center gap-2.5">
                                  <div className="w-8 h-8 rounded-xl bg-[#008272] text-white font-bold flex items-center justify-center text-xs">
                                    {app.fullName.split(' ').map(n => n[0]).join('')}
                                  </div>
                                  <div>
                                    <p className="font-bold text-slate-900">{app.fullName}</p>
                                    <p className="text-[10px] text-slate-400 font-mono">TSC Reg: #{app.tscNumber}</p>
                                  </div>
                                </div>
                              </td>
                              <td className="py-4 px-3">
                                <span className="px-2.5 py-1 rounded-lg bg-slate-100 text-slate-800 font-semibold text-[11px]">
                                  {app.role}
                                </span>
                              </td>
                              <td className="py-4 px-3 text-slate-600">{app.subjects}</td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>

                    <div className="pt-4 flex items-center justify-between text-xs text-slate-400">
                      <span>Displaying pending applications received today</span>
                      <span className="flex items-center gap-1.5"><span className="w-2 h-2 rounded-full bg-emerald-500"></span> System gateway synchronised</span>
                    </div>
                  </div>

                  {/* Right Column: Applicant Verification Inspector Panel */}
                  <div className="lg:col-span-5 bg-slate-50 border border-slate-200 rounded-2xl p-5 space-y-5 text-xs">
                    {selectedApplicant ? (
                      <div className="space-y-4">
                        <div className="flex items-center justify-between pb-3 border-b border-slate-200/60">
                          <div>
                            <div className="flex items-center gap-2">
                              <h4 className="font-bold text-slate-900 text-base">{selectedApplicant.fullName}</h4>
                              <span className="px-2 py-0.5 bg-purple-100 text-purple-700 font-semibold rounded text-[10px]">{selectedApplicant.role}</span>
                            </div>
                            <p className="text-[11px] text-slate-500">Applied today at 8:30 am · Form 3 & 4 Lead</p>
                          </div>
                        </div>

                        <div className="grid grid-cols-2 gap-3 bg-white p-3 rounded-xl border border-slate-200">
                          <div>
                            <span className="text-[10px] text-slate-400 block">Assigned Subjects</span>
                            <strong className="text-slate-900 font-semibold">{selectedApplicant.subjects}</strong>
                          </div>
                          <div>
                            <span className="text-[10px] text-slate-400 block">Contact Number</span>
                            <strong className="text-slate-900 font-mono">+254 733 ••• 111</strong>
                          </div>
                        </div>

                        {/* Uploaded Credentials */}
                        <div className="space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-slate-700 uppercase tracking-wider text-[10px]">Uploaded Credentials (2)</span>
                            <span className="text-[10px] text-[#008272] font-semibold flex items-center gap-1">🔒 Encrypted</span>
                          </div>

                          {selectedApplicant.credentials.map((cred) => (
                            <div key={cred.id} className="bg-white border border-slate-200 rounded-xl p-3 flex items-center justify-between">
                              <div className="flex items-center gap-2.5">
                                <FileText className="w-4 h-4 text-[#008272]" />
                                <div>
                                  <p className="font-bold text-slate-900 text-xs">{cred.name}</p>
                                  <p className="text-[10px] text-slate-400 font-mono">{cred.filename} ({cred.size})</p>
                                </div>
                              </div>
                              <button className="px-2.5 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold rounded-lg text-[11px] transition flex items-center gap-1">
                                <Eye className="w-3 h-3" /> View
                              </button>
                            </div>
                          ))}
                        </div>

                        {/* Physical In-Person Verification Checkbox */}
                        <div className="bg-white border border-slate-200 rounded-xl p-3.5 space-y-2">
                          <span className="font-bold text-slate-900 block text-[11px] uppercase tracking-wider text-slate-400">Physical In-Person Verification</span>
                          <label className="flex items-start gap-2.5 cursor-pointer">
                            <input 
                              type="checkbox" 
                              checked={verificationChecked}
                              onChange={(e) => setVerificationChecked(e.target.checked)}
                              className="mt-0.5 rounded text-[#008272]"
                            />
                            <span className="text-slate-700 leading-snug">
                              <strong className="block text-slate-900 font-bold mb-0.5">I met this person and their documents match them</strong>
                              Tick the box once you've checked the person and documents.
                            </span>
                          </label>
                        </div>

                        {/* Action Buttons */}
                        <div className="grid grid-cols-2 gap-3 pt-2">
                          <button 
                            onClick={() => handleApprovalAction('Rejected')}
                            disabled={isSubmitting}
                            className="py-3 bg-rose-50 hover:bg-rose-100 border border-rose-200 text-rose-700 font-semibold rounded-xl transition flex items-center justify-center gap-1.5"
                          >
                            <XCircle className="w-4 h-4" /> Reject
                          </button>
                          <button 
                            onClick={() => handleApprovalAction('Approved')}
                            disabled={isSubmitting || !verificationChecked}
                            className={`py-3 rounded-xl font-semibold transition flex items-center justify-center gap-2 shadow-md ${
                              verificationChecked 
                                ? 'bg-[#008272] hover:bg-[#006e61] text-white shadow-[#008272]/20' 
                                : 'bg-slate-200 text-slate-400 cursor-not-allowed'
                            }`}
                          >
                            {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                            <CheckCircle2 className="w-4 h-4" /> Approve staff
                          </button>
                        </div>

                        <p className="text-[10px] text-slate-400 text-center">
                          🔒 Verified documents are purged from storage immediately upon approval.
                        </p>
                      </div>
                    ) : (
                      <p className="text-slate-400 text-center py-24">Select an applicant to review credentials.</p>
                    )}
                  </div>

                </div>

              </div>

            </div>
          )}

        </div>

      </main>

    </div>
  );
}