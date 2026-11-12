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
  AlertTriangle,
  CheckCircle2, 
  ArrowRight, 
  QrCode, 
  UserPlus, 
  BarChart3, 
  History, 
  Plus, 
  Send, 
  RefreshCw 
} from 'lucide-react';

interface ClassAssessmentProgress {
  id: string;
  gradeClass: string;
  classMaster: string;
  enteredSubjects: number;
  totalSubjects: number;
  status: 'Waiting on marks' | 'Ready to release' | 'Action required';
}

interface BlockerItem {
  id: string;
  subject: string;
  gradeClass: string;
  instructor: string;
  issueDescription: string;
}

interface UpcomingEventItem {
  id: string;
  month: string;
  day: string;
  title: string;
  dateFormatted: string;
  time: string;
  location: string;
  audienceTag: string;
}

interface AuditLogEntry {
  id: string;
  actor: string;
  actorInitials: string;
  actionDescription: string;
  timestamp: string;
  category: string;
}

interface LeadershipDashboardData {
  schoolName: string;
  currentTerm: string;
  currentDateFormatted: string;
  user: {
    name: string;
    initials: string;
    role: string;
  };
  deanUser: {
    name: string;
    role: string;
  };
  setupProgress: {
    completedSteps: number;
    totalSteps: number;
    description: string;
    isComplete: boolean;
  };
  kpis: {
    waitingVerificationCount: number;
    qrWindowRemaining: string;
    qrExpiresAt: string;
    qrSessionActive: boolean;
    staffApplicationsCount: number;
  };
  assessmentProgress: {
    classes: ClassAssessmentProgress[];
    submissionDeadline: string;
  };
  needsAttention: {
    blockerCount: number;
    blockers: BlockerItem[];
    feeReconciliationStatus: string;
    complianceAuditsUpToDate: boolean;
  };
  comingUp: {
    events: UpcomingEventItem[];
    summaryText: string;
  };
  recentChanges: {
    logs: AuditLogEntry[];
    retentionPolicyText: string;
  };
}

export default function AdminLeadershipHomePage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('home');
  const [viewingRole, setViewingRole] = useState<'principal' | 'dean'>('principal');

  // Dynamic state replacing hardcoded mocks
  const [dashboardData, setDashboardData] = useState<LeadershipDashboardData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  // Fetch dashboard metrics from backend API
  useEffect(() => {
    const fetchDashboard = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch(`/api/v1/admin/dashboard?perspective=${viewingRole}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize leadership dashboard telemetry from institutional registry.');
        }

        const data = await res.json();
        setDashboardData(data);
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading dashboard overview.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchDashboard();
  }, [viewingRole]);

  // Handle sending reminder to an instructor with missing marks
  const handleSendReminder = async (blockerId: string) => {
    setActionLoadingId(blockerId);
    try {
      const res = await fetch(`/api/v1/admin/assessments/blockers/${blockerId}/remind`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) throw new Error('Failed to dispatch reminder notification.');
      alert('Reminder notification dispatched to instructor.');
    } catch (err: any) {
      alert(err.message || 'Could not send reminder.');
    } finally {
      setActionLoadingId(null);
    }
  };

  const setupPercent = dashboardData?.setupProgress
    ? Math.round((dashboardData.setupProgress.completedSteps / dashboardData.setupProgress.totalSteps) * 100)
    : 0;

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#eaf4f3] via-[#f2f7f6] to-white flex text-slate-800">
      
      {/* Left Institutional Sidebar */}
      <aside className="w-64 bg-white border-r border-slate-200/90 flex flex-col justify-between p-5 shrink-0 hidden md:flex shadow-sm">
        <div className="space-y-6">
          <div>
            <h1 className="font-bold text-slate-900 text-lg tracking-tight">Ikonex Academy</h1>
            <p className="text-xs text-slate-500 font-medium">{dashboardData?.schoolName || 'Institutional Portal'}</p>
          </div>

          <nav className="space-y-1 text-sm font-medium text-slate-600">
            <button onClick={() => setActiveNav('home')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <Home className="w-4 h-4 text-[#008272]" /> Home
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
              {dashboardData?.user.initials || 'PR'}
            </div>
            <div className="leading-tight">
              <p className="text-xs font-bold text-slate-900">{dashboardData?.user.name || 'Leadership User'}</p>
              <p className="text-[10px] text-slate-500 font-medium">{dashboardData?.user.role || 'Principal'}</p>
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
            {dashboardData?.user.role || 'Principal'} • {dashboardData?.user.name || 'Authenticated'}
          </span>
          <span className="bg-white/80 px-3 py-1 rounded-full border border-slate-200 font-medium">
            Live Institutional Environment
          </span>
          <div className="flex items-