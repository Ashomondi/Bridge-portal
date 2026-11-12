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
  Phone 
} from 'lucide-react';

interface ParentStatusData {
  reference: string;
  schoolName: string;
  status: 'Pending' | 'Verified' | 'Needs review' | 'Removed';
  stepCurrent: number;
  stepTotal: number;
  submittedAt: string;
  parent: {
    fullName: string;
    relationship: string;
    phone: string;
  };
  student: {
    fullName: string;
    admissionNumber: string;
    gradeClass: string;
  };
  attachment: {
    filename: string;
    statusText: string;
  };
  registry: {
    phone: string;
    hours: string;
  };
}

export default function ParentRegistrationStatusPage() {
  const router = useRouter();

  // Dynamic state replacing mock data
  const [statusData, setStatusData] = useState<ParentStatusData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Interactive Demo State Inspector override for testing UI states
  const [inspectorStateOverride, setInspectorStateOverride] = useState<'Pending' | 'Verified' | 'Needs review' | 'Removed' | null>(null);

  // Fetch parent status from backend API
  const fetchStatus = async (override?: string) => {
    setIsRefreshing(true);
    setErrorMessage(null);
    try {
      const query = override ? `?state=${override}` : '';
      const res = await fetch(`/api/v1/parent/registration-status${query}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) {
        throw new Error('Failed to synchronize parent registration status from institutional gateway.');
      }

      const data = await res.json();
      setStatusData(data);
    } catch (err: any) {
      setErrorMessage(err.message || 'An unexpected error occurred while fetching your registration status.');
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
      <div className="max-w-xl w-full bg-white border border-slate-200/90 rounded-[32px] shadow-xl overflow-hidden my-4">
        
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

        {/* Header & Body Content */}
        <div className="p-6 md:p-8 space-y-6">
          
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-4 border-b border-slate-100">
            <div>
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1">
                <span>Ikonex Academy</span> • <span className="text-[#008272] font-mono">REG-4F2K</span>
              </div>
              <h1 className="text-2xl font-extrabold text-slate-900 tracking-tight">Parent Registration Status</h1>
              <p className="text-xs text-slate-500">Lakeview Secondary · Parent Portal Access</p>
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
              <p className="text-xs font-medium">Synchronizing registration status...</p>
            </div>
          ) : (
            <div className="space-y-6">
              
              {/* Main Status Hero Card */}
              <div className="bg-gradient-to-br from-[#f8faf9] via-white to-[#edf4f3] border border-slate-200/90 rounded-[28px] p-6 md:p-7 shadow-sm space-y-6">
                
                <div className="flex items-center justify-between">
                  <span className="px-3 py-1 bg-purple-100 text-purple-800 font-bold text-xs rounded-full flex items-center gap-1.5">
                    <Clock className="w-3.5 h-3.5" /> Waiting for verification
                  </span>
                  <span className="text-xs font-bold text-slate-400">Step {statusData?.stepCurrent || 2} of {statusData?.stepTotal || 3}</span>
                </div>

                {/* Vertical Step Timeline */}
                <div className="space-y-6 relative before:absolute before:left-3.5 before:top-3 before:bottom-3 before:w-0.5 before:bg-slate-200">
                  
                  {/* Step 1: Submitted Online */}
                  <div className="relative flex items-start gap-3.5">
                    <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center z-10 shrink-0 shadow-sm">
                      <CheckCircle2 className="w-4 h-4" />
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="font-bold text-slate-900 text-sm">Submitted online</h4>
                      <p className="text-xs text-slate-500">{statusData?.submittedAt || 'Wed 29 Oct, 8:42 am'}</p>
                    </div>
                  </div>

                  {/* Step 2: Verified in person */}
                  <div className="relative flex items-start gap-3.5">
                    <div className="w-7 h-7 rounded-full bg-[#008272] text-white flex items-center justify-center z-10 shrink-0 ring-4 ring-[#eaf4f3] shadow-sm">
                      <span className="w-2 h-2 rounded-full bg-white animate-pulse"></span>
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="font-bold text-slate-900 text-sm">Verified in person</h4>
                      <p className="text-xs text-slate-600">School office desk · Pending attendance</p>
                    </div>
                  </div>

                  {/* Step 3: Access Granted */}
                  <div className="relative flex items-start gap-3.5">
                    <div className="w-7 h-7 rounded-full bg-slate-200 text-slate-400 flex items-center justify-center z-10 shrink-0">
                      <Lock className="w-3.5 h-3.5" />
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="font-bold text-slate-400 text-sm">Access granted</h4>
                      <p className="text-xs text-slate-400">Instant SMS notification & portal unblocked</p>
                    </div>
                  </div>

                </div>

                {/* Instruction Callout Box */}
                <div className="bg-[#f2f8f7] border border-[#d0e8e4] rounded-2xl p-4 text-xs text-slate-700 leading-relaxed flex items-start gap-3">
                  <ShieldCheck className="w-5 h-5 text-[#008272] shrink-0 mt-0.5" />
                  <p>
                    A staff member needs to meet you and check your ID. Please visit the school office today and present your physical national ID card.
                  </p>
                </div>

              </div>

              {/* Submitted Details Card */}
              <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
                <div className="flex items-center justify-between pb-2 border-b border-slate-100">
                  <h3 className="font-bold text-slate-900 text-sm">Submitted details</h3>
                  <Lock className="w-4 h-4 text-slate-400" />
                </div>

                <div className="space-y-3 text-xs">
                  
                  {/* Parent Row */}
                  <div className="bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-xl bg-purple-100 text-purple-700 font-bold flex items-center justify-center text-xs">
                        GM
                      </div>
                      <div>
                        <h4 className="font-bold text-slate-900">{statusData?.parent.fullName || 'Grace Mwangi'}</h4>
                        <p className="text-[11px] text-slate-500">{statusData?.parent.relationship || 'Mother'} · {statusData?.parent.phone || '+254 7 ••• 678'}</p>
                      </div>
                    </div>
                    <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                  </div>

                  {/* Student Row */}
                  <div className="bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-xl bg-emerald-100 text-[#008272] font-bold flex items-center justify-center text-xs">
                        AW
                      </div>
                      <div>
                        <h4 className="font-bold text-slate-900">{statusData?.student.fullName || 'Amani Wekesa'}</h4>
                        <p className="text-[11px] text-slate-500">Adm {statusData?.student.admissionNumber || '4521'} · {statusData?.student.gradeClass || 'Grade 8 East'}</p>
                      </div>
                    </div>
                    <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                  </div>

                  {/* Attachment Row */}
                  <div className="bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <FileText className="w-5 h-5 text-[#008272]" />
                      <div>
                        <h4 className="font-bold text-slate-900">{statusData?.attachment.filename || 'national_id_front.jpg'}</h4>
                        <p className="text-[10px] text-slate-500">Stored in vault · Will be purged after confirmation</p>
                      </div>
                    </div>
                    <span className="px-2.5 py-1 bg-amber-100 text-amber-800 font-bold text-[10px] rounded-md">
                      {statusData?.attachment.statusText || 'Pending Desk ID Match'}
                    </span>
                  </div>

                </div>
              </div>

              {/* What Happens Next Checklist */}
              <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
                <h3 className="font-bold text-slate-900 text-sm">What happens next</h3>

                <div className="space-y-3 text-xs">
                  <div className="flex items-start gap-3 bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl">
                    <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-800 font-bold text-[10px] flex items-center justify-center shrink-0">1</span>
                    <p className="text-slate-700 leading-snug"><strong className="text-slate-900 block font-bold mb-0.5">Visit school office</strong> Walk in during ordinary business hours. No prior appointment required.</p>
                  </div>
                  <div className="flex items-start gap-3 bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl">
                    <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-800 font-bold text-[10px] flex items-center justify-center shrink-0">2</span>
                    <p className="text-slate-700 leading-snug"><strong className="text-slate-900 block font-bold mb-0.5">Present National ID</strong> Show your original identification card to Mr. Otieno or desk registrars.</p>
                  </div>
                  <div className="flex items-start gap-3 bg-slate-50 border border-slate-200/80 p-3.5 rounded-2xl">
                    <span className="w-5 h-5 rounded-full bg-purple-100 text-purple-800 font-bold text-[10px] flex items-center justify-center shrink-0">3</span>
                    <p className="text-slate-700 leading-snug"><strong className="text-slate-900 block font-bold mb-0.5">Immediate activation</strong> Your account unlocks right away and receives an SMS confirmation pin.</p>
                  </div>
                </div>
              </div>

              {/* Lakeview Secondary Reception Footer Card */}
              <div className="bg-[#f8faf9] border border-slate-200 rounded-2xl p-4 flex items-center justify-between text-xs">
                <div className="flex items-center gap-2.5">
                  <div className="w-8 h-8 rounded-xl bg-[#008272] text-white flex items-center justify-center font-bold">
                    🏫
                  </div>
                  <div>
                    <h4 className="font-bold text-slate-900">Lakeview Secondary Reception</h4>
                    <p className="text-[11px] text-slate-500 font-mono">0700 000 000 · Open 8:00 am to 5:00 pm</p>
                  </div>
                </div>
                <button className="p-2.5 bg-white hover:bg-slate-100 border border-slate-200 rounded-xl text-[#008272] shadow-sm">
                  <Phone className="w-4 h-4" />
                </button>
              </div>

              {/* Refresh Footer & Timestamp */}
              <div className="flex items-center justify-between text-xs text-slate-400 pt-2">
                <span>Refreshes every 30s · Last checked just now</span>
                <button 
                  onClick={() => fetchStatus(inspectorStateOverride || undefined)}
                  disabled={isRefreshing}
                  className="font-bold text-[#008272] hover:underline flex items-center gap-1"
                >
                  <RefreshCw className={`w-3 h-3 ${isRefreshing ? 'animate-spin' : ''}`} /> Refresh
                </button>
              </div>

              {/* Demo State Inspector Bar (Interactive Testing) */}
              <div className="bg-slate-900 text-white rounded-2xl p-4 space-y-2 text-xs shadow-inner">
                <div className="flex items-center justify-between text-[11px] text-slate-400">
                  <span className="uppercase tracking-wider font-bold">Demo State Inspector</span>
                  <span className="text-emerald-400 font-mono">Switch view</span>
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
                    onClick={() => setInspectorStateOverride('Verified')}
                    className={`py-2 rounded-xl font-semibold transition ${
                      inspectorStateOverride === 'Verified' ? 'bg-[#008272] text-white shadow' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                    }`}
                  >
                    2. Verified
                  </button>
                  <button 
                    onClick={() => setInspectorStateOverride('Needs review')}
                    className={`py-2 rounded-xl font-semibold transition ${
                      inspectorStateOverride === 'Needs review' ? 'bg-[#008272] text-white shadow' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
                    }`}
                  >
                    3. Needs review
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