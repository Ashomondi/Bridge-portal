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
  XCircle, 
  Printer, 
  Flag, 
  QrCode, 
  FileText 
} from 'lucide-react';

interface ParentVerificationItem {
  id: string;
  fullName: string;
  relationship: string;
  childrenCount: number;
  registeredPhone: string;
  checkInTime: string;
  status: 'Waiting' | 'Verified' | 'Rejected';
  claimedStudent: {
    name: string;
    admissionNumber: string;
    gradeClass: string;
    classTeacher: string;
    matchStatus: string;
    note?: string;
  };
  identityDocument: {
    idNumber: string;
    imageUrl: string;
    fileSize: string;
    legalFullName: string;
    dateOfBirth: string;
    subCounty: string;
    opticalCheckPassed: boolean;
  };
}

export default function AdminParentsPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('parents');
  const [activeTab, setActiveTab] = useState<'waiting' | 'verified' | 'rejected'>('waiting');

  // Dynamic state replacing hardcoded mocks
  const [parentsQueue, setParentsQueue] = useState<ParentVerificationItem[]>([]);
  const [selectedParent, setSelectedParent] = useState<ParentVerificationItem | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Search & Filter state within queue
  const [searchQuery, setSearchQuery] = useState('');

  // Checklist state for active verification
  const [checklistPersonMatch, setChecklistPersonMatch] = useState(false);
  const [checklistStudentMatch, setChecklistStudentMatch] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch queue from backend API
  useEffect(() => {
    const fetchVerificationQueue = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch(`/api/v1/admin/parents/verification-queue?status=${activeTab}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize parental verification queue from institutional registry.');
        }

        const data = await res.json();
        setParentsQueue(data.queue || []);

        // Automatically select first item if available
        if (data.queue && data.queue.length > 0 && (!selectedParent || !data.queue.find((p: ParentVerificationItem) => p.id === selectedParent.id))) {
          setSelectedParent(data.queue[0]);
          setChecklistPersonMatch(false);
          setChecklistStudentMatch(false);
        } else if (data.queue.length === 0) {
          setSelectedParent(null);
        }
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading verification data.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchVerificationQueue();
  }, [activeTab]);

  // Handle Verify Parent action
  const handleVerifyParent = async (status: 'Verified' | 'Rejected') => {
    if (!selectedParent) return;
    setIsSubmitting(true);
    try {
      const res = await fetch(`/api/v1/admin/parents/verify/${selectedParent.id}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status,
          personMatchConfirmed: checklistPersonMatch,
          studentMatchConfirmed: checklistStudentMatch,
        }),
      });

      if (!res.ok) throw new Error('Failed to record verification decision on institutional server.');

      // Remove from current queue view and move to next
      const remaining = parentsQueue.filter(p => p.id !== selectedParent.id);
      setParentsQueue(remaining);
      setSelectedParent(remaining.length > 0 ? remaining[0] : null);
      setChecklistPersonMatch(false);
      setChecklistStudentMatch(false);

      alert(`Parent identity successfully marked as ${status}.`);
    } catch (err: any) {
      alert(err.message || 'Could not complete verification action.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Filtered queue based on search
  const filteredQueue = parentsQueue.filter(p => 
    p.fullName.toLowerCase().includes(searchQuery.toLowerCase()) ||
    p.claimedStudent.admissionNumber.toLowerCase().includes(searchQuery.toLowerCase()) ||
    p.claimedStudent.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

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
          
          {/* Top Subheader Bar */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-2 border-b border-slate-200">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                <span>PARENT REGISTRATION & IDENTITY CHECK</span> • <span>Academic Session Term 1 · 2025</span>
              </div>
              <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">Parents</h2>
            </div>
            <div className="flex items-center gap-3">
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs font-semibold rounded-full">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span> Dean Level Permissions Active
              </span>
              <button className="px-3 py-1.5 bg-white hover:bg-slate-50 border border-slate-200 text-slate-700 font-semibold rounded-xl text-xs transition shadow-sm flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5 text-slate-400" /> Audit Log
              </button>
            </div>
          </div>

          {/* Secondary Switcher: QR Session vs Verification Queue */}
          <div className="flex items-center gap-4 border-b border-slate-200 pb-2 text-xs font-semibold">
            <button className="flex items-center gap-1.5 pb-2 text-slate-500 hover:text-slate-900 transition">
              <QrCode className="w-4 h-4 text-slate-400" /> QR Code Session <span className="px-2 py-0.5 bg-slate-100 text-slate-700 rounded-full text-[10px]">Live</span>
            </button>
            <button className="flex items-center gap-1.5 pb-2 text-[#008272] border-b-2 border-[#008272] transition">
              <Users className="w-4 h-4" /> Verification Queue ({parentsQueue.length})
            </button>
          </div>

          {/* Main Content Layout: Queue List (Left) + Parent Verification Inspector (Right) */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Left Column: Queue List & Filters */}
            <div className="lg:col-span-5 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
              
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Queue</h3>
                  <p className="text-xs text-slate-500">Gate 1 Desk</p>
                </div>
              </div>

              {/* Status Tabs */}
              <div className="bg-slate-100 p-1 rounded-2xl border border-slate-200 flex text-xs font-semibold">
                <button 
                  onClick={() => setActiveTab('waiting')}
                  className={`flex-1 py-2 rounded-xl transition ${activeTab === 'waiting' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
                >
                  Waiting ({parentsQueue.filter(p => p.status === 'Waiting').length})
                </button>
                <button 
                  onClick={() => setActiveTab('verified')}
                  className={`flex-1 py-2 rounded-xl transition ${activeTab === 'verified' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
                >
                  Verified
                </button>
                <button 
                  onClick={() => setActiveTab('rejected')}
                  className={`flex-1 py-2 rounded-xl transition ${activeTab === 'rejected' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
                >
                  Rejected
                </button>
              </div>

              {/* Search bar inside queue */}
              <div className="relative">
                <Search className="absolute left-3.5 top-3 w-4 h-4 text-slate-400" />
                <input 
                  type="text" 
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search by name, ID or student"
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl pl-10 pr-4 py-2.5 text-xs text-slate-900 focus:outline-none focus:border-[#008272]"
                />
              </div>

              {/* Queue Items List */}
              {isLoading ? (
                <div className="py-24 flex flex-col items-center justify-center gap-3 text-slate-400">
                  <Loader2 className="w-7 h-7 animate-spin text-[#008272]" />
                  <p className="text-xs">Loading verification queue...</p>
                </div>
              ) : errorMessage ? (
                <div className="py-12 text-center text-rose-600 text-xs flex flex-col items-center gap-2">
                  <AlertCircle className="w-5 h-5" />
                  <p>{errorMessage}</p>
                </div>
              ) : filteredQueue.length === 0 ? (
                <div className="py-16 text-center text-slate-400 text-xs">
                  No parents found in this queue.
                </div>
              ) : (
                <div className="space-y-2.5 overflow-y-auto max-h-[500px]">
                  {filteredQueue.map((item) => {
                    const isSelected = selectedParent?.id === item.id;
                    return (
                      <div 
                        key={item.id}
                        onClick={() => setSelectedParent(item)}
                        className={`p-3.5 rounded-2xl border transition cursor-pointer flex items-center justify-between ${
                          isSelected ? 'bg-[#eaf4f3]/60 border-[#008272] shadow-sm' : 'bg-white border-slate-200 hover:bg-slate-50'
                        }`}
                      >
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="w-9 h-9 rounded-xl bg-slate-900 text-white font-bold flex items-center justify-center text-xs shrink-0">
                            {item.fullName.split(' ').map(n => n[0]).join('')}
                          </div>
                          <div className="min-w-0">
                            <div className="flex items-center gap-1.5">
                              <h4 className="font-bold text-slate-900 text-sm truncate">{item.fullName}</h4>
                              <span className="w-2 h-2 rounded-full bg-rose-500 shrink-0"></span>
                            </div>
                            <p className="text-xs text-slate-500 truncate">{item.relationship} · {item.childrenCount} child</p>
                          </div>
                        </div>
                        <div className="text-right shrink-0">
                          <span className="block text-xs font-bold text-slate-900">{item.checkInTime}</span>
                          <span className="text-[10px] text-slate-400">Queue #{item.id}</span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}

              <div className="bg-[#f4f8f7] border border-slate-200 rounded-2xl p-3 text-xs text-slate-600 flex items-center justify-between">
                <div>
                  <p className="font-bold text-slate-900">Rapid NFC Check Ready</p>
                  <p className="text-[11px] text-slate-500">Hold card near scanner to match</p>
                </div>
                <Users className="w-5 h-5 text-[#008272]" />
              </div>

            </div>

            {/* Right Column: Detailed Parent Inspector & Verification Checklist */}
            <div className="lg:col-span-7 bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm space-y-6">
              
              {selectedParent ? (
                <div className="space-y-6">
                  
                  {/* Header Profile Info */}
                  <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-slate-100">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <h3 className="text-2xl font-extrabold text-slate-900">{selectedParent.fullName}</h3>
                        <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-100 text-purple-700">Waiting for verification</span>
                      </div>
                      <p className="text-xs text-slate-500 font-mono">
                        Registered phone: <span className="font-bold text-slate-800">{selectedParent.registeredPhone}</span>
                      </p>
                      <p className="text-xs text-slate-500">
                        Relationship: <span className="font-semibold text-slate-800">{selectedParent.relationship}</span> · Self Check-in: <span className="font-bold text-slate-900">{selectedParent.checkInTime}</span>
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button className="p-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-600 transition"><Printer className="w-4 h-4" /></button>
                      <button className="p-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-600 transition"><Flag className="w-4 h-4" /></button>
                    </div>
                  </div>

                  {/* Claimed Student Relation Card */}
                  <div className="bg-[#f8faf9] border border-slate-200 rounded-2xl p-4 space-y-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-xl bg-slate-900 text-white font-bold flex items-center justify-center text-xs">
                          {selectedParent.claimedStudent.name.split(' ').map(n => n[0]).join('')}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <h4 className="font-bold text-slate-900 text-sm">{selectedParent.claimedStudent.name}</h4>
                            <span className="inline-flex items-center gap-1 text-[10px] font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-md">
                              <CheckCircle2 className="w-3 h-3" /> Matches an existing student record
                            </span>
                          </div>
                          <p className="text-xs text-slate-500">
                            Admission {selectedParent.claimedStudent.admissionNumber} · {selectedParent.claimedStudent.gradeClass} · Class Teacher: {selectedParent.claimedStudent.classTeacher}
                          </p>
                        </div>
                      </div>
                      <button className="px-3 py-1.5 bg-white hover:bg-slate-50 border border-slate-200 text-slate-700 font-semibold rounded-xl text-xs shadow-sm shrink-0">
                        View Student Dossier
                      </button>
                    </div>
                    {selectedParent.claimedStudent.note && (
                      <p className="text-[11px] text-amber-800 bg-amber-50 border border-amber-200 p-2.5 rounded-xl">
                        ⚠️ {selectedParent.claimedStudent.note}
                      </p>
                    )}
                  </div>

                  {/* Submitted Identity Document */}
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-slate-700 uppercase tracking-wider">Submitted Identity Document</span>
                      <span className="px-2.5 py-1 bg-slate-100 text-slate-600 text-[10px] font-semibold rounded-full flex items-center gap-1">
                        🔒 Encrypted Storage (No download permitted)
                      </span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {/* Image Preview Box */}
                      <div className="relative bg-slate-900 rounded-2xl overflow-hidden border border-slate-200 p-3 flex flex-col justify-between min-h-[160px]">
                        <div className="flex items-center justify-between text-white text-[11px]">
                          <span>National ID · Front</span>
                          <span className="font-mono text-slate-400">image.jpg · 1.2 MB</span>
                        </div>
                        <div className="flex items-center justify-center py-4">
                          <FileText className="w-10 h-10 text-slate-500" />
                        </div>
                        <div className="flex items-center justify-between text-[11px] text-slate-300 pt-2 border-t border-slate-800">
                          <span className="font-mono">ID Number: {selectedParent.identityDocument.idNumber}</span>
                          <span className="text-emerald-400 font-bold flex items-center gap-1">
                            <CheckCircle2 className="w-3 h-3" /> Optical Check Passed
                          </span>
                        </div>
                      </div>

                      {/* Field Comparison Summary */}
                      <div className="bg-slate-50 border border-slate-200 rounded-2xl p-4 space-y-2 text-xs">
                        <p className="font-bold text-slate-900 text-[11px] uppercase tracking-wider text-slate-400">Field Comparison Summary</p>
                        <div className="space-y-1.5">
                          <div className="flex justify-between py-1 border-b border-slate-200/60">
                            <span className="text-slate-500">Legal Full Name</span>
                            <span className="font-bold text-slate-900">{selectedParent.identityDocument.legalFullName}</span>
                          </div>
                          <div className="flex justify-between py-1 border-b border-slate-200/60">
                            <span className="text-slate-500">Date of Birth</span>
                            <span className="font-bold text-slate-900">{selectedParent.identityDocument.dateOfBirth}</span>
                          </div>
                          <div className="flex justify-between py-1">
                            <span className="text-slate-500">Sub-County of Reg.</span>
                            <span className="font-bold text-slate-900">{selectedParent.identityDocument.subCounty}</span>
                          </div>
                        </div>
                        <p className="text-[10px] text-[#008272] bg-[#eaf4f3] p-2 rounded-xl flex items-center gap-1 font-medium mt-2">
                          <ShieldCheck className="w-3.5 h-3.5 shrink-0" /> Verified authentic security pattern matching government national issuance registry standards.
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Mandatory In-Person Checklist */}
                  <div className="space-y-3 pt-2 border-t border-slate-100">
                    <div className="flex items-center justify-between">
                      <span className="text-xs font-bold text-slate-900 uppercase tracking-wider">Mandatory In-Person Checklist</span>
                      <span className="text-[11px] text-rose-600 font-semibold bg-rose-50 px-2 py-0.5 rounded-md">Required for verification</span>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
                      <label className="flex items-start gap-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 p-3.5 rounded-2xl cursor-pointer transition">
                        <input 
                          type="checkbox" 
                          checked={checklistPersonMatch}
                          onChange={(e) => setChecklistPersonMatch(e.target.checked)}
                          className="mt-0.5 rounded text-[#008272] focus:ring-[#008272]"
                        />
                        <span className="leading-snug text-slate-700">
                          <strong className="block text-slate-900 mb-0.5">The person here matches the document</strong>
                          Visual face-to-photo identification confirmed at reception.
                        </span>
                      </label>

                      <label className="flex items-start gap-3 bg-slate-50 hover:bg-slate-100 border border-slate-200 p-3.5 rounded-2xl cursor-pointer transition">
                        <input 
                          type="checkbox" 
                          checked={checklistStudentMatch}
                          onChange={(e) => setChecklistStudentMatch(e.target.checked)}
                          className="mt-0.5 rounded text-[#008272] focus:ring-[#008272]"
                        />
                        <span className="leading-snug text-slate-700">
                          <strong className="block text-slate-900 mb-0.5">The student exists at this school</strong>
                          Amani Wekesa verified in current Lakeview Secondary records.
                        </span>
                      </label>
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="pt-4 flex items-center justify-between gap-4">
                    <button 
                      onClick={() => handleVerifyParent('Rejected')}
                      disabled={isSubmitting}
                      className="px-6 py-3 bg-rose-50 hover:bg-rose-100 border border-rose-200 text-rose-700 font-semibold rounded-2xl text-xs transition flex items-center gap-1.5"
                    >
                      <XCircle className="w-4 h-4" /> Not verified
                    </button>
                    <div className="flex items-center gap-3">
                      <span className="text-[11px] text-slate-400">Tick both boxes after you've checked in person</span>
                      <button 
                        onClick={() => handleVerifyParent('Verified')}
                        disabled={isSubmitting || !checklistPersonMatch || !checklistStudentMatch}
                        className={`px-6 py-3 rounded-2xl text-xs font-semibold transition flex items-center gap-2 shadow-md ${
                          checklistPersonMatch && checklistStudentMatch 
                            ? 'bg-[#008272] hover:bg-[#006e61] text-white shadow-[#008272]/20' 
                            : 'bg-slate-200 text-slate-400 cursor-not-allowed'
                        }`}
                      >
                        {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                        Verify Parent
                      </button>
                    </div>
                  </div>

                  {/* Footer Authority Note */}
                  <div className="bg-[#f2f8f7] border border-slate-200/80 rounded-2xl p-4 flex items-center justify-between text-xs text-slate-600">
                    <div className="flex items-center gap-2">
                      <QrCode className="w-4 h-4 text-[#008272]" />
                      <span>Who can verify while this QR code is active</span>
                    </div>
                    <span className="font-bold text-[#008272]">Dean & Principal Authority</span>
                  </div>

                </div>
              ) : (
                <div className="py-32 text-center text-slate-400 text-xs">
                  Select a parent from the queue to review identification and complete verification.
                </div>
              )}

            </div>

          </div>

        </div>

      </main>

    </div>
  );
}