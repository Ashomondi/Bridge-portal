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
  CheckCircle2, 
  FileText, 
  UserCheck, 
  RefreshCw 
} from 'lucide-react';

interface TemporaryAssignment {
  id: string;
  subject: string;
  gradeClass: string;
  originalTeacher: string;
  coveredBy: string;
  coveredByInitials: string;
  startDate: string;
  reason: string;
  status: 'Active Cover' | 'Concluded';
}

interface TeachingData {
  academicTerm: string;
  auditingRole: 'Dean of Studies' | 'Principal / Deputy';
  auditStatus: {
    isValid: boolean;
    percentage: number;
    description: string;
    streamsCovered: number;
    totalStreams: number;
    activeOverrides: number;
  };
  activeReassignments: TemporaryAssignment[];
  coverageHistory: TemporaryAssignment[];
  coverPolicyGuide: string[];
}

export default function AdminTeachingPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('teaching');
  const [activeTab, setActiveTab] = useState<'marks-cover' | 'class-teachers' | 'subject-teachers'>('marks-cover');

  // Dynamic state replacing hardcoded mocks
  const [teachingData, setTeachingData] = useState<TeachingData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch teaching data from backend API
  useEffect(() => {
    const fetchTeachingData = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch('/api/v1/admin/teaching', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize teaching assignments from institutional registry.');
        }

        const data = await res.json();
        setTeachingData(data);
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading teaching assignments.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchTeachingData();
  }, []);

  // Handle ending or reassigning a mark cover
  const handleEndCover = async (assignmentId: string) => {
    setIsSubmitting(true);
    try {
      const res = await fetch(`/api/v1/admin/teaching/reassignments/${assignmentId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) throw new Error('Failed to terminate mark cover assignment.');

      // Refresh local state
      if (teachingData) {
        setTeachingData({
          ...teachingData,
          activeReassignments: teachingData.activeReassignments.filter(a => a.id !== assignmentId),
        });
      }
      alert('Mark cover successfully concluded and original teacher write-access restored.');
    } catch (err: any) {
      alert(err.message || 'Could not conclude cover.');
    } finally {
      setIsSubmitting(false);
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
            <button onClick={() => router.push('/admin/parents')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Users className="w-4 h-4 text-slate-400" /> Parents
            </button>
            <button onClick={() => setActiveNav('teaching')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <BookOpen className="w-4 h-4 text-[#008272]" /> Teaching
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
          
          {/* Top Subheader & Role Toggle Card */}
          <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-6">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                <span>Academic Governance</span> • <span className="text-[#008272]">Term 3 · 2024 CBC Cycle</span>
              </div>
              <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">Teaching</h2>
              <p className="text-sm text-slate-600">
                Assign class teachers, subject teachers, and temporary marks cover.
              </p>
            </div>

            {/* Auditing Role Selector */}
            <div className="flex items-center gap-3 bg-slate-50 p-2 rounded-2xl border border-slate-200 text-xs">
              <span className="font-bold text-slate-500 pl-2">Auditing role:</span>
              <div className="flex bg-white rounded-xl p-1 shadow-sm border border-slate-200">
                <button 
                  onClick={() => setTeachingData(prev => prev ? { ...prev, auditingRole: 'Dean of Studies' } : null)}
                  className={`px-3 py-1.5 rounded-lg font-semibold transition ${
                    teachingData?.auditingRole === 'Dean of Studies' ? 'bg-[#008272] text-white' : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Dean of Studies
                </button>
                <button 
                  onClick={() => setTeachingData(prev => prev ? { ...prev, auditingRole: 'Principal / Deputy' } : null)}
                  className={`px-3 py-1.5 rounded-lg font-semibold transition ${
                    teachingData?.auditingRole === 'Principal / Deputy' ? 'bg-[#008272] text-white' : 'text-slate-600 hover:text-slate-900'
                  }`}
                >
                  Principal / Deputy
                </button>
              </div>
            </div>
          </div>

          {/* Sub Navigation Tabs & KNEC Sync Status */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-200 pb-3">
            <div className="flex items-center gap-2 text-xs font-semibold">
              <button 
                onClick={() => setActiveTab('marks-cover')}
                className={`px-4 py-2 rounded-xl transition flex items-center gap-1.5 ${
                  activeTab === 'marks-cover' ? 'bg-[#008272] text-white shadow-sm' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-50'
                }`}
              >
                <RefreshCw className="w-3.5 h-3.5" /> Marks cover <span className="px-1.5 py-0.2 bg-white/20 rounded-full text-[10px]">1 active</span>
              </button>
              <button 
                onClick={() => setActiveTab('class-teachers')}
                className={`px-4 py-2 rounded-xl transition ${
                  activeTab === 'class-teachers' ? 'bg-[#008272] text-white shadow-sm' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-50'
                }`}
              >
                Class teachers
              </button>
              <button 
                onClick={() => setActiveTab('subject-teachers')}
                className={`px-4 py-2 rounded-xl transition ${
                  activeTab === 'subject-teachers' ? 'bg-[#008272] text-white shadow-sm' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-50'
                }`}
              >
                Subject teachers
              </button>
            </div>

            <div className="flex items-center gap-2 text-xs text-slate-500 font-medium">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              Synced with KNEC CBA portal
            </div>
          </div>

          {/* Error Banner if API fails */}
          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* Main Layout: Active Reassignments & Coverage History (Left) + Audit Status & Policy Guide (Right) */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Left Column: Reassignments & History */}
            <div className="lg:col-span-8 space-y-6">
              
              {isLoading ? (
                <div className="bg-white border border-slate-200 rounded-[28px] py-32 flex flex-col items-center justify-center gap-3 text-slate-400">
                  <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
                  <p className="text-xs font-medium">Loading teaching assignments...</p>
                </div>
              ) : (
                <>
                  {/* Active Temporary Reassignments Card */}
                  <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-5">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="font-bold text-slate-900 text-base">Active Temporary Reassignments</h3>
                        <p className="text-xs text-slate-500">Authoritative marks delegation overriding original teacher records.</p>
                      </div>
                      <button className="px-4 py-2 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition flex items-center gap-1.5 shadow-md shadow-[#008272]/20">
                        <RefreshCw className="w-3.5 h-3.5" /> Reassign marks entry
                      </button>
                    </div>

                    <div className="space-y-4">
                      {teachingData?.activeReassignments?.length === 0 ? (
                        <p className="text-xs text-slate-400 py-6 text-center">No active temporary reassignments.</p>
                      ) : (
                        teachingData?.activeReassignments?.map((item) => (
                          <div key={item.id} className="bg-[#f8faf9] border border-slate-200 rounded-2xl p-5 space-y-4">
                            <div className="flex items-center justify-between">
                              <div className="flex items-center gap-2">
                                <h4 className="font-extrabold text-slate-900 text-base">{item.subject} · {item.gradeClass}</h4>
                                <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800">
                                  {item.status}
                                </span>
                              </div>
                              <button 
                                onClick={() => handleEndCover(item.id)}
                                disabled={isSubmitting}
                                className="px-3 py-1.5 bg-white hover:bg-rose-50 border border-slate-200 hover:border-rose-200 text-rose-600 font-semibold rounded-xl text-xs transition shadow-sm"
                              >
                                End cover
                              </button>
                            </div>

                            <div className="flex flex-wrap items-center gap-4 text-xs text-slate-600">
                              <span>Original: <strong className="text-slate-900">{item.originalTeacher}</strong></span>
                              <span>→</span>
                              <span>Covered by:</span>
                              <div className="flex items-center gap-2 bg-white border border-slate-200 px-3 py-1 rounded-xl shadow-sm">
                                <span className="w-5 h-5 rounded-md bg-[#008272] text-white text-[10px] font-bold flex items-center justify-center">
                                  {item.coveredByInitials}
                                </span>
                                <strong className="text-slate-900">{item.coveredBy}</strong>
                              </div>
                            </div>

                            <div className="flex items-center gap-6 text-[11px] text-slate-500 pt-2 border-t border-slate-100">
                              <span>📅 Since {item.startDate}</span>
                              <span>⚠️ Reason: {item.reason}</span>
                            </div>
                          </div>
                        ))
                      )}
                    </div>

                    <div className="bg-[#f2f8f7] border border-slate-200/80 rounded-2xl p-4 text-xs text-slate-600 leading-relaxed">
                      ℹ️ Temporary reassignment delegates raw score entry, continuous assessment submission, and remark generation. Original teachers lose write-access until the cover is formally closed.
                    </div>
                  </div>

                  {/* Recent Coverage History */}
                  <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
                    <h3 className="font-bold text-slate-900 text-base">Recent Coverage History</h3>
                    <div className="divide-y divide-slate-100 text-xs">
                      {teachingData?.coverageHistory?.map((history) => (
                        <div key={history.id} className="py-3.5 flex items-center justify-between">
                          <div className="space-y-0.5">
                            <p className="font-bold text-slate-900">{history.subject} · {history.gradeClass}</p>
                            <p className="text-slate-500 text-[11px]">{history.originalTeacher} covered by {history.coveredBy} ({history.reason})</p>
                          </div>
                          <span className="text-slate-400 font-medium">{history.status}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}

            </div>

            {/* Right Column: Audit Status Card & Cover Policy Guide */}
            <div className="lg:col-span-4 space-y-6">
              
              {/* Audit Status Card */}
              <div className="bg-slate-900 text-white rounded-[28px] p-6 shadow-xl space-y-6 relative overflow-hidden">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Audit Status</span>
                  <ShieldCheck className="w-5 h-5 text-[#008272]" />
                </div>

                <div className="space-y-2">
                  <h3 className="text-3xl font-extrabold tracking-tight text-emerald-400">
                    {teachingData?.auditStatus.percentage || 100}% Valid
                  </h3>
                  <p className="text-xs text-slate-300 leading-relaxed">
                    {teachingData?.auditStatus.description || 'All 24 active assessment streams have assigned teachers.'}
                  </p>
                </div>

                {/* Progress bar */}
                <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                  <div className="bg-emerald-400 h-full rounded-full" style={{ width: `${teachingData?.auditStatus.percentage || 100}%` }}></div>
                </div>

                <div className="grid grid-cols-2 gap-4 pt-2 border-t border-slate-800 text-xs">
                  <div>
                    <span className="block text-slate-400 text-[10px]">Streams Covered</span>
                    <strong className="text-lg font-bold text-white">{teachingData?.auditStatus.streamsCovered || 24} / {teachingData?.auditStatus.totalStreams || 24}</strong>
                  </div>
                  <div>
                    <span className="block text-slate-400 text-[10px]">Active Overrides</span>
                    <strong className="text-lg font-bold text-amber-400">{teachingData?.auditStatus.activeOverrides || 1} Stream</strong>
                  </div>
                </div>
              </div>

              {/* Cover Policy Guide */}
              <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4 text-xs">
                <h3 className="font-bold text-slate-900 text-base">Cover Policy Guide</h3>
                <ul className="space-y-3 text-slate-600">
                  <li className="flex items-start gap-2.5">
                    <CheckCircle2 className="w-4 h-4 text-[#008272] shrink-0 mt-0.5" />
                    <span>Covers must be approved by the Dean of Studies before marks submission deadlines.</span>
                  </li>
                  <li className="flex items-start gap-2.5">
                    <CheckCircle2 className="w-4 h-4 text-[#008272] shrink-0 mt-0.5" />
                    <span>Parent notifications reflect original teacher on report cards, while system logs credit the cover teacher.</span>
                  </li>
                  <li className="flex items-start gap-2.5">
                    <CheckCircle2 className="w-4 h-4 text-[#008272] shrink-0 mt-0.5" />
                    <span>Covers auto-expire upon manual release by the Dean.</span>
                  </li>
                </ul>
              </div>

            </div>

          </div>

        </div>

      </main>

    </div>
  );
}