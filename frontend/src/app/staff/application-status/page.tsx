'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  ArrowLeft, 
  User, 
  RefreshCw, 
  CheckCircle2, 
  Clock, 
  Lock, 
  FileText, 
  ShieldCheck, 
  AlertCircle, 
  Loader2, 
  Download, 
  ExternalLink 
} from 'lucide-react';

interface ApplicationStatusData {
  reference: string;
  schoolName: string;
  status: 'Pending' | 'Approved' | 'Declined' | 'Removed';
  stepCurrent: number;
  stepTotal: number;
  submittedAt: string;
  applicant: {
    fullName: string;
    phone: string;
    role: string;
    subjects: string;
    tscNumber: string;
  };
  attachments: Array<{
    id: string;
    name: string;
    filename: string;
    size: string;
  }>;
  registry: {
    phone: string;
    hours: string;
  };
}

export default function StaffApplicationStatusPage() {
  const router = useRouter();

  // Dynamic state replacing hardcoded mocks
  const [appData, setAppData] = useState<ApplicationStatusData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Interactive Demo State Inspector override for testing UI states
  const [inspectorStateOverride, setInspectorStateOverride] = useState<'Pending' | 'Approved' | 'Declined' | 'Removed' | null>(null);

  // Fetch application status from backend API
  const fetchStatus = async (override?: string) => {
    setIsRefreshing(true);
    setErrorMessage(null);
    try {
      const query = override ? `?state=${override}` : '';
      const res = await fetch(`/api/v1/staff/application-status${query}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) {
        throw new Error('Failed to synchronize application status from institutional gateway.');
      }

      const data = await res.json();
      setAppData(data);
    } catch (err: any) {
      setErrorMessage(err.message || 'An unexpected error occurred while fetching your application status.');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    fetchStatus(inspectorStateOverride || undefined);
  }, [inspectorStateOverride]);

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#eaf4f3] via-[#f2f7f6] to-white flex flex-col items-center justify-start text-slate-800 p-4 md:p-8">
      
      {/* Centered Web Container */}
      <div className="max-w-2xl w-full bg-white border border-slate-200/90 rounded-[32px] shadow-xl overflow-hidden my-4">
        
        {/* Top Demo Banner Bar */}
        <div className="w-full bg-[#f8faf9] border-b border-slate-200 px-6 py-2.5 flex items-center justify-between text-xs text-slate-500">
          <span className="font-medium">Demo environment • sample data only</span>
          <div className="flex items-center gap-3">
            <button 
              onClick={() => router.back()}
              className="p-1 hover:bg-slate-200/60 rounded-full text-slate-700 transition"
              title="Go back"
            >
              <ArrowLeft className="w-4 h-4" />
            </button>
            <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center font-bold text-xs shadow-sm">
              <User className="w-3.5 h-3.5" />
            </div>
          </div>
        </div>

        {/* Header & Refresh Bar */}
        <div className="p-6 md:p-8 space-y-6">
          
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-100">
            <div>
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1">
                <span>Staff Onboarding</span> • <span className="text-[#008272] font-mono">Ref: {appData?.reference || 'APP-91QD'}</span>
              </div>
              <h1 className="text-2xl font-extrabold text-slate-900 tracking-tight">Staff Application Status</h1>
              <p className="text-xs text-slate-500">{appData?.schoolName || 'Lakeview Secondary'}</p>
            </div>

            <div className="flex items-center gap-3">
              <span className="text-[11px] text-slate-400 hidden sm:inline">Refreshes every 30s · Just now</span>
              <button 
                onClick={() => fetchStatus(inspectorStateOverride || undefined)}
                disabled={isRefreshing}
                className="px-3.5 py-2 bg-slate-50 hover:bg-slate-100 border border-slate-200 text-slate-700 font-semibold rounded-xl text-xs transition flex items-center gap-1.5 shadow-sm"
              >
                <RefreshCw className={`w-3.5 h-3.5 text-slate-500 ${isRefreshing ? 'animate-spin' : ''}`} /> Refresh
              </button>
            </div>
          </div>

          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {isLoading ? (
            <div className="py-32 flex flex-col items-center justify-center gap-3 text-slate-400">
              <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
              <p className="text-xs font-medium">Synchronizing application status...</p>
            </div>
          ) : (
            <div className="space-y-6">
              
              {/* Main Status Hero Card */}
              <div className="bg-gradient-to-br from-[#f8faf9] via-white to-[#edf4f3] border border-slate-200/90 rounded-[28px] p-6 md:p-7 shadow-sm space-y-6">
                
                <div className="flex items-center justify-between">
                  <span className="px-3 py-1 bg-purple-100 text-purple-800 font-bold text-xs rounded-full flex items-center gap-1.5">
                    <Clock className="w-3.5 h-3.5" /> Waiting for approval
                  </span>
                  <span className="text-xs font-bold text-slate-400">Step {appData?.stepCurrent || 2} of {appData?.stepTotal || 3}</span>
                </div>

                {/* Vertical Step Timeline */}
                <div className="space-y-6 relative before:absolute before:left-3.5 before:top-3 before:bottom-3 before:w-0.5 before:bg-slate-200">
                  
                  {/* Step 1: Submitted */}
                  <div className="relative flex items-start gap-3.5">
                    <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center z-10 shrink-0 shadow-sm">
                      <CheckCircle2 className="w-4 h-4" />
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="font-bold text-slate-900 text-sm">Submitted</h4>
                      <p className="text-xs text-slate-500">{appData?.submittedAt || 'Wed 29 Oct, 8:30 am'}</p>
                    </div>
                  </div>

                  {/* Step 2: Checked in person */}
                  <div className="relative flex items-start gap-3.5">
                    <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center z-10 shrink-0 ring-4 ring-[#eaf4f3] shadow-sm">
                      <span className="w-2 h-2 rounded-full bg-white animate-pulse"></span>
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="font-bold text-slate-900 text-sm">Checked in person</h4>
                      <p className="text-xs text-slate-600">Principal's office · Pending document verification</p>
                    </div>
                  </div>

                  {/* Step 3: Approved */}
                  <div className="relative flex items-start gap-3.5">
                    <div className="w-7 h-7 rounded-full bg-slate-200 text-slate-400 flex items-center justify-center z-10 shrink-0">
                      <Lock className="w-3.5 h-3.5" />
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="font-bold text-slate-400 text-sm">Approved</h4>
                      <p className="text-xs text-slate-400">Staff portal access unlocked upon sign-off</p>
                    </div>
                  </div>

                </div>

                {/* Instruction Callout Box */}
                <div className="bg-[#f2f8f7] border border-[#d0e8e4] rounded-2xl p-4 text-xs text-slate-700 leading-relaxed flex items-start gap-3">
                  <ShieldCheck className="w-5 h-5 text-[#008272] shrink-0 mt-0.5" />
                  <p>
                    The principal, deputy principal, or dean will meet you to check your documents. Please go to the principal's office with your national ID.
                  </p>
                </div>

              </div>

              {/* Next Steps Checklist Card */}
              <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
                <h3 className="font-bold text-slate-900 text-sm flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-[#008272]" /> Next steps checklist
                </h3>

                <div className="space-y-3 text-xs">
                  <div className="flex items-start gap-3 bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl">
                    <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-800 font-bold text-[10px] flex items-center justify-center shrink-0">1</span>
                    <p className="text-slate-700 leading-snug">Report to the administration building main reception desk.</p>
                  </div>
                  <div className="flex items-start gap-3 bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl">
                    <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-800 font-bold text-[10px] flex items-center justify-center shrink-0">2</span>
                    <p className="text-slate-700 leading-snug">Present original National ID and TSC credentials to <strong className="text-slate-900">Ms. Achieng</strong>.</p>
                  </div>
                  <div className="flex items-start gap-3 bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl">
                    <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-800 font-bold text-[10px] flex items-center justify-center shrink-0">3</span>
                    <p className="text-slate-700 leading-snug">Receive immediate SMS credentials and digital role provisioning.</p>
                  </div>
                </div>
              </div>

              {/* Application Details Summary Card */}
              <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
                <div className="flex items-center justify-between pb-2 border-b border-slate-100">
                  <h3 className="font-bold text-slate-900 text-sm">Application details</h3>
                  <FileText className="w-4 h-4 text-slate-400" />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
                  <div className="bg-slate-50 border border-slate-200/70 p-3.5 rounded-2xl space-y-0.5">
                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Applicant</span>
                    <p className="font-bold text-slate-900 text-sm">{appData?.applicant.fullName || 'Mr. Peter Mutua'}</p>
                    <p className="text-slate-500 font-mono text-[11px]">{appData?.applicant.phone || '+254 733 ••• 111'}</p>
                  </div>

                  <div className="bg-slate-50 border border-slate-200/70 p-3.5 rounded-2xl space-y-0.5">
                    <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Requested Role</span>
                    <p className="font-bold text-slate-900 text-sm">{appData?.applicant.role || 'Teacher'}</p>
                    <p className="text-slate-500 text-[11px]">{appData?.applicant.subjects || 'Mathematics, Integrated Science'}</p>
                  </div>
                </div>

                <div className="bg-slate-50 border border-slate-200/70 p-3.5 rounded-2xl space-y-0.5 text-xs">
                  <span className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Professional Identifier</span>
                  <p className="font-bold text-slate-900 font-mono">{appData?.applicant.tscNumber || 'TSC - 0800800'}</p>
                </div>

                {/* Uploaded Attachments */}
                <div className="space-y-2 pt-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-bold text-slate-700">Uploaded Attachments</span>
                    <span className="text-[10px] text-[#008272] font-semibold flex items-center gap-1">🔒 Vault protected</span>
                  </div>

                  {appData?.attachments?.map((att) => (
                    <div key={att.id} className="bg-slate-50 border border-slate-200/80 rounded-xl p-3 flex items-center justify-between text-xs">
                      <div className="flex items-center gap-2.5">
                        <FileText className="w-4 h-4 text-[#008272]" />
                        <div>
                          <p className="font-bold text-slate-900">{att.name}</p>
                          <p className="text-[10px] text-slate-400 font-mono">{att.filename} · {att.size}</p>
                        </div>
                      </div>
                      <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                    </div>
                  ))}
                  
                  <p className="text-[10px] text-slate-400 pt-1">
                    Physical validation required. Digital attachments will be purged upon in-person check.
                  </p>
                </div>

              </div>

              {/* Lakeview Secondary Registry Footer Card */}
              <div className="bg-[#f8faf9] border border-slate-200 rounded-2xl p-4 flex items-center justify-between text-xs">
                <div className="flex items-center gap-2.5">
                  <UsersIcon className="w-4 h-4 text-[#008272]" />
                  <div>
                    <h4 className="font-bold text-slate-900">Lakeview Secondary Registry</h4>
                    <p className="text-[11px] text-slate-500 font-mono">0700 000 000</p>
                  </div>
                </div>
                <span className="text-slate-500 font-medium">Open 8:00 am to 5:00 pm</span>
              </div>

              {/* Demo State Inspector Bar (Interactive Testing) */}
              <div className="bg-slate-900 text-white rounded-2xl p-4 space-y-2 text-xs shadow-inner">
                <div className="flex items-center justify-between text-[11px] text-slate-400">
                  <span className="uppercase tracking-wider font-bold">Demo State Inspector</span>
                  <span className="text-emerald-400 font-mono">Interactive Preview</span>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                  <button 
                    onClick={() => setInspectorStateOverride('Pending')}
                    className={`py-2 rounded-xl font-semibold transition ${
                      (!inspectorStateOverride || inspectorStateOverride === 'Pending') ? 'bg-[#008272] text-white shadow' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                    }`}
                  >
                    1. Pending
                  </button>
                  <button 
                    onClick={() => setInspectorStateOverride('Approved')}
                    className={`py-2 rounded-xl font-semibold transition ${
                      inspectorStateOverride === 'Approved' ? 'bg-[#008272] text-white shadow' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                    }`}
                  >
                    2. Approved
                  </button>
                  <button 
                    onClick={() => setInspectorStateOverride('Declined')}
                    className={`py-2 rounded-xl font-semibold transition ${
                      inspectorStateOverride === 'Declined' ? 'bg-[#008272] text-white shadow' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                    }`}
                  >
                    3. Declined
                  </button>
                  <button 
                    onClick={() => setInspectorStateOverride('Removed')}
                    className={`py-2 rounded-xl font-semibold transition ${
                      inspectorStateOverride === 'Removed' ? 'bg-[#008272] text-white shadow' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                    }`}
                  >
                    4. Removed
                  </button>
                </div>
              </div>

            </div>
          )}

        </div>

      </div>

    </div>
  );
}

function UsersIcon(props: any) {
  return (
    <svg {...props} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
    </svg>
  );
}