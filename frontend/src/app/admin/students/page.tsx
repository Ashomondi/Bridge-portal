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
  Plus, 
  Download, 
  ShieldCheck, 
  LogOut, 
  Loader2, 
  AlertCircle, 
  X, 
  ChevronRight, 
  ChevronLeft, 
  HelpCircle, 
  FileText, 
  CheckCircle2 
} from 'lucide-react';

interface Student {
  id: string;
  admissionNumber: string;
  fullName: string;
  cbcLevel: string;
  track: string;
  gradeClass: string;
  dormitory: string;
  guardianCount: number;
  enrollmentDate: string;
  scholarType: string;
  guardians: Array<{
    name: string;
    relationship: string;
    phone: string;
    status: 'Linked' | 'Pending';
    verifiedBy: string;
    verifiedAt: string;
  }>;
  assessmentProfile?: {
    target: string;
    exceeding: string;
  };
}

export default function AdminStudentsPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('students');

  // Dynamic state replacing hardcoded mocks
  const [students, setStudents] = useState<Student[]>([]);
  const [selectedStudent, setSelectedStudent] = useState<Student | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Filters & Search State
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedClassFilter, setSelectedClassFilter] = useState('All classes');
  const [selectedStatusFilter, setSelectedStatusFilter] = useState('All status');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  // Modals state
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [isRollForwardModalOpen, setIsRollForwardModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Add student form state
  const [newAdmissionNumber, setNewAdmissionNumber] = useState('');
  const [newFullName, setNewFullName] = useState('');
  const [newGradeClass, setNewGradeClass] = useState('Grade 8 East');
  const [newDormitory, setNewDormitory] = useState('Kilimanjaro House');
  const [newScholarType, setNewScholarType] = useState('Boarding Scholar');

  // Fetch students from backend API with query filters
  useEffect(() => {
    const fetchStudents = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const queryParams = new URLSearchParams({
          page: currentPage.toString(),
          search: searchQuery,
          class: selectedClassFilter,
          status: selectedStatusFilter,
        });

        const res = await fetch(`/api/v1/admin/students?${queryParams.toString()}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize student directory from institutional registry.');
        }

        const data = await res.json();
        setStudents(data.students || []);
        setTotalPages(data.totalPages || 1);

        // Automatically select the first student if none selected
        if (data.students && data.students.length > 0 && !selectedStudent) {
          setSelectedStudent(data.students[0]);
        }
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading student records.');
      } finally {
        setIsLoading(false);
      }
    };

    const debounceTimer = setTimeout(fetchStudents, 300);
    return () => clearTimeout(debounceTimer);
  }, [searchQuery, selectedClassFilter, selectedStatusFilter, currentPage]);

  // Handle adding a new student
  const handleAddStudent = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/admin/students', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          admissionNumber: newAdmissionNumber,
          fullName: newFullName,
          gradeClass: newGradeClass,
          dormitory: newDormitory,
          scholarType: newScholarType,
        }),
      });

      if (!res.ok) throw new Error('Failed to register new student.');

      const createdStudent = await res.json();
      setStudents((prev) => [createdStudent, ...prev]);
      setSelectedStudent(createdStudent);
      setIsAddModalOpen(false);

      // Reset form
      setNewAdmissionNumber('');
      setNewFullName('');
    } catch (err: any) {
      alert(err.message || 'Could not add student.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle saving changes made in the Dean Inspector panel
  const handleSaveInspectorChanges = async () => {
    if (!selectedStudent) return;
    setIsSubmitting(true);
    try {
      const res = await fetch(`/api/v1/admin/students/${selectedStudent.id}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          gradeClass: selectedStudent.gradeClass,
          dormitory: selectedStudent.dormitory,
        }),
      });

      if (!res.ok) throw new Error('Failed to update student academic record.');

      const updated = await res.json();
      setStudents((prev) => prev.map((s) => (s.id === updated.id ? updated : s)));
      alert('Student record successfully updated and synchronized.');
    } catch (err: any) {
      alert(err.message || 'Save failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Grade Roll-Forward action
  const handleRollForwardAction = async () => {
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/admin/students/roll-forward', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) throw new Error('Roll-forward batch execution failed.');

      alert('Grade roll-forward successfully executed for all eligible scholars.');
      setIsRollForwardModalOpen(false);
    } catch (err: any) {
      alert(err.message || 'Roll-forward failed.');
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
            <button onClick={() => router.push('/admin/teaching')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <BookOpen className="w-4 h-4 text-slate-400" /> Teaching
            </button>
            <button onClick={() => router.push('/admin/timetables')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <Clock className="w-4 h-4 text-slate-400" /> Timetables
            </button>
            <button onClick={() => router.push('/admin/calendar')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 transition">
              <CalendarIcon className="w-4 h-4 text-slate-400" /> Calendar
            </button>
            <button onClick={() => setActiveNav('students')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <GraduationCap className="w-4 h-4 text-[#008272]" /> Students
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
            Academic Year 2025 · Term 1 • Updated 8m ago
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
          
          {/* Page Title Header */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="space-y-1">
              <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">Students</h2>
              <p className="text-sm text-slate-600">
                Manage Lakeview Secondary student directory and parental linkages.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <button className="px-4 py-2 bg-white hover:bg-slate-50 border border-slate-200 font-semibold rounded-xl text-xs transition flex items-center gap-1.5 shadow-sm">
                <Download className="w-3.5 h-3.5 text-slate-500" /> Export Directory
              </button>
              <button 
                onClick={() => setIsAddModalOpen(true)}
                className="px-4 py-2 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition flex items-center gap-1.5 shadow-md shadow-[#008272]/20"
              >
                <Plus className="w-3.5 h-3.5" /> Add Student
              </button>
            </div>
          </div>

          {/* Roll-Forward & Linkage Guidance Banner Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Left Roll-Forward Banner */}
            <div className="lg:col-span-7 bg-gradient-to-r from-purple-50/80 via-[#f4f2f9] to-white border border-purple-200/80 rounded-[28px] p-6 shadow-sm flex flex-col justify-between space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="px-2.5 py-1 bg-purple-200/70 text-purple-900 font-bold text-[10px] rounded-lg">
                    Grade Roll-Forward Ready
                  </span>
                  <span className="text-xs text-purple-700 font-medium">Year-Start Action</span>
                </div>
                <span className="text-[11px] text-slate-500 font-medium">Audit lock closes Friday 5:00 PM</span>
              </div>

              <div className="space-y-1">
                <h3 className="text-lg font-bold text-slate-900">
                  <span className="text-purple-700">142 students</span> are ready to move up to their next grade level based on Term 3 CBC evaluations.
                </h3>
              </div>

              <div className="flex items-center gap-3 pt-1">
                <button 
                  onClick={() => setIsRollForwardModalOpen(true)}
                  className="px-4 py-2 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition shadow-sm"
                >
                  Carry Forward
                </button>
                <button className="px-4 py-2 bg-white hover:bg-slate-50 border border-slate-200 text-slate-700 font-semibold rounded-xl text-xs transition shadow-sm">
                  Review List
                </button>
              </div>
            </div>

            {/* Right Linkage Guidance Card */}
            <div className="lg:col-span-5 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="font-bold text-slate-900 text-sm flex items-center gap-2">
                  🔗 How parents are linked
                </h3>
                <HelpCircle className="w-4 h-4 text-slate-400" />
              </div>

              <div className="grid grid-cols-2 gap-2 bg-slate-50 p-1.5 rounded-2xl border border-slate-200 text-xs font-semibold text-center">
                <div className="py-2 bg-white rounded-xl shadow-sm text-slate-900 border border-slate-200/60">
                  Teachers enter & notify
                </div>
                <div className="py-2 text-slate-500 hover:text-slate-900 transition cursor-pointer">
                  Opening day self-register
                </div>
              </div>

              <p className="text-[11px] text-slate-500 leading-relaxed">
                When both methods are active, opening day registrations are automatically matched to admission numbers.
              </p>
            </div>

          </div>

          {/* Search & Filter Bar */}
          <div className="flex flex-col md:flex-row items-center gap-4 bg-white border border-slate-200 rounded-2xl p-4 shadow-sm">
            <div className="relative flex-1 w-full">
              <Search className="absolute left-3.5 top-3 w-4 h-4 text-slate-400" />
              <input 
                type="text" 
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search by admission or name..."
                className="w-full bg-slate-50 border border-slate-200 rounded-xl pl-10 pr-4 py-2.5 text-xs text-slate-900 focus:outline-none focus:border-[#008272]"
              />
            </div>
            <div className="flex items-center gap-3 w-full md:w-auto">
              <select 
                value={selectedClassFilter} 
                onChange={(e) => setSelectedClassFilter(e.target.value)}
                className="bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-xs text-slate-700 font-medium focus:outline-none"
              >
                <option>All classes</option>
                <option>Grade 8 East</option>
                <option>Grade 8 West</option>
                <option>CBC Level 3</option>
              </select>
              <select 
                value={selectedStatusFilter} 
                onChange={(e) => setSelectedStatusFilter(e.target.value)}
                className="bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-xs text-slate-700 font-medium focus:outline-none"
              >
                <option>All status</option>
                <option>Active</option>
                <option>Pending</option>
              </select>
            </div>
          </div>

          {/* Main Content Layout: Table (Left) + Dean Inspector (Right) */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Left Column: Enrolled Learners Table */}
            <div className="lg:col-span-8 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-4">
              
              <div className="flex items-center justify-between pb-3 border-b border-slate-100">
                <div className="flex items-center gap-3">
                  <h3 className="font-bold text-slate-900 text-base">Enrolled Learners</h3>
                  <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-100 text-[#008272]">
                    420 Active
                  </span>
                </div>
                <span className="text-xs text-slate-400">Showing 1–3 of 420 students</span>
              </div>

              {isLoading ? (
                <div className="py-32 flex flex-col items-center justify-center gap-3 text-slate-400">
                  <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
                  <p className="text-xs font-medium">Loading student directory...</p>
                </div>
              ) : errorMessage ? (
                <div className="py-12 text-center text-rose-600 text-xs flex flex-col items-center gap-2">
                  <AlertCircle className="w-6 h-6" />
                  <p>{errorMessage}</p>
                </div>
              ) : students.length === 0 ? (
                <div className="py-20 text-center text-slate-400 text-xs">
                  No students matched your search criteria.
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse text-xs">
                    <thead>
                      <tr className="border-b border-slate-100 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                        <th className="py-3 px-3">Admission #</th>
                        <th className="py-3 px-3">Full Name</th>
                        <th className="py-3 px-3">Class</th>
                        <th className="py-3 px-3">Dormitory</th>
                        <th className="py-3 px-3">Guardians</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 font-medium text-slate-700">
                      {students.map((student) => {
                        const isSelected = selectedStudent?.id === student.id;
                        return (
                          <tr 
                            key={student.id} 
                            onClick={() => setSelectedStudent(student)}
                            className={`cursor-pointer transition hover:bg-slate-50 ${isSelected ? 'bg-[#eaf4f3]/50' : ''}`}
                          >
                            <td className="py-4 px-3 font-mono font-bold text-[#008272]">{student.admissionNumber}</td>
                            <td className="py-4 px-3">
                              <div className="flex items-center gap-2.5">
                                <div className="w-7 h-7 rounded-lg bg-slate-900 text-white font-bold flex items-center justify-center text-[10px]">
                                  {student.fullName.split(' ').map(n => n[0]).join('')}
                                </div>
                                <div>
                                  <p className="font-bold text-slate-900">{student.fullName}</p>
                                  <p className="text-[10px] text-slate-400">{student.cbcLevel} • {student.track}</p>
                                </div>
                              </div>
                            </td>
                            <td className="py-4 px-3">
                              <span className="px-2.5 py-1 rounded-lg bg-slate-100 text-slate-800 font-semibold text-[11px]">
                                {student.gradeClass}
                              </span>
                            </td>
                            <td className="py-4 px-3 text-slate-500 italic">{student.dormitory}</td>
                            <td className="py-4 px-3">
                              <span className="inline-flex items-center gap-1 text-slate-700 font-semibold">
                                <Users className="w-3.5 h-3.5 text-slate-400" /> {student.guardianCount} guardian{student.guardianCount === 1 ? '' : 's'}
                              </span>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Table Pagination Footer */}
              <div className="pt-4 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
                <span>Direct parent notifications enabled for 389 of 420 scholars</span>
                <div className="flex items-center gap-3">
                  <button onClick={() => setCurrentPage(p => Math.max(1, p - 1))} className="p-1 hover:bg-slate-100 rounded-lg"><ChevronLeft className="w-4 h-4" /></button>
                  <span className="font-semibold text-slate-900">Page {currentPage} of {totalPages}</span>
                  <button onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))} className="p-1 hover:bg-slate-100 rounded-lg"><ChevronRight className="w-4 h-4" /></button>
                </div>
              </div>

            </div>

            {/* Right Column: Dean Record Inspector Panel */}
            <div className="lg:col-span-4 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-6">
              
              <div className="flex items-center justify-between pb-3 border-b border-slate-100">
                <div className="flex items-center gap-2">
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-[#008272]">Active</span>
                  <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Dean Record Inspector</span>
                </div>
              </div>

              {selectedStudent ? (
                <div className="space-y-5 text-xs">
                  <div>
                    <h3 className="text-lg font-bold text-slate-900">{selectedStudent.fullName} · {selectedStudent.admissionNumber}</h3>
                    <p className="text-slate-400">Enrolled Term 1, 2024 · {selectedStudent.scholarType}</p>
                  </div>

                  {/* Academic Class Dropdown */}
                  <div className="space-y-1">
                    <label className="block font-bold text-slate-700">Academic Class</label>
                    <select 
                      value={selectedStudent.gradeClass}
                      onChange={(e) => setSelectedStudent({ ...selectedStudent, gradeClass: e.target.value })}
                      className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-semibold text-slate-900 focus:outline-none"
                    >
                      <option value="Grade 8 East">Grade 8 East</option>
                      <option value="Grade 8 West">Grade 8 West</option>
                      <option value="Grade 7 North">Grade 7 North</option>
                    </select>
                    <p className="text-[10px] text-slate-400">Controlled by Dean of Studies</p>
                  </div>

                  {/* Dormitory Allocation Dropdown */}
                  <div className="space-y-1">
                    <label className="block font-bold text-slate-700">Dormitory Allocation</label>
                    <select 
                      value={selectedStudent.dormitory}
                      onChange={(e) => setSelectedStudent({ ...selectedStudent, dormitory: e.target.value })}
                      className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-semibold text-slate-900 focus:outline-none"
                    >
                      <option value="Kilimanjaro House">Kilimanjaro House</option>
                      <option value="Mount Kenya House">Mount Kenya House</option>
                      <option value="Day scholar">Day scholar</option>
                    </select>
                  </div>

                  {/* Linked Guardians */}
                  <div className="space-y-2 pt-2 border-t border-slate-100">
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-slate-900">Linked Guardians</span>
                      <button className="text-[11px] font-semibold text-[#008272] hover:underline">+ Add Link</button>
                    </div>

                    {selectedStudent.guardians && selectedStudent.guardians.map((g, idx) => (
                      <div key={idx} className="bg-[#f8faf9] border border-slate-200 rounded-xl p-3 space-y-1">
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-slate-900">{g.name} ({g.relationship})</span>
                          <span className="text-[10px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full">Linked</span>
                        </div>
                        <p className="font-mono text-[11px] text-slate-600">{g.phone}</p>
                        <p className="text-[10px] text-slate-400">Verified by {g.verifiedBy} · {g.verifiedAt}</p>
                      </div>
                    ))}
                  </div>

                  {/* Assessment Profile Card */}
                  <div className="bg-[#eaf4f3] border border-[#d0e8e4] rounded-2xl p-4 space-y-1">
                    <div className="flex items-center gap-1.5 font-bold text-[#008272] text-xs">
                      <FileText className="w-4 h-4" /> Term 1 Assessment Profile
                    </div>
                    <p className="text-slate-700 font-medium">8 Core Learning Areas · Target Exceeding</p>
                  </div>

                  {/* Action Buttons */}
                  <div className="pt-2 flex items-center gap-2">
                    <button 
                      onClick={handleSaveInspectorChanges}
                      disabled={isSubmitting}
                      className="flex-1 bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3 rounded-xl transition shadow-md shadow-[#008272]/20 flex items-center justify-center gap-2 text-xs"
                    >
                      {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                      Save Record Changes
                    </button>
                    <button className="px-4 py-3 bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold rounded-xl transition text-xs">
                      Activity Log
                    </button>
                  </div>
                </div>
              ) : (
                <p className="text-xs text-slate-400 py-12 text-center">Select a learner from the directory to inspect records.</p>
              )}

            </div>

          </div>

        </div>

      </main>

      {/* Add Student Modal */}
      {isAddModalOpen && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-[28px] max-w-md w-full p-6 shadow-2xl space-y-5">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <h3 className="font-bold text-slate-900 text-lg">Register New Student</h3>
              <button onClick={() => setIsAddModalOpen(false)} className="p-1 hover:bg-slate-100 rounded-full text-slate-500">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleAddStudent} className="space-y-4 text-xs">
              <div>
                <label className="block font-semibold text-slate-700 mb-1">Admission Number</label>
                <input 
                  type="text" 
                  required
                  value={newAdmissionNumber}
                  onChange={(e) => setNewAdmissionNumber(e.target.value)}
                  placeholder="e.g. #4524" 
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-mono text-slate-900 focus:outline-none focus:border-[#008272]"
                />
              </div>

              <div>
                <label className="block font-semibold text-slate-700 mb-1">Full Name</label>
                <input 
                  type="text" 
                  required
                  value={newFullName}
                  onChange={(e) => setNewFullName(e.target.value)}
                  placeholder="e.g. Brian Kiprono" 
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none focus:border-[#008272]"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold text-slate-700 mb-1">Grade Class</label>
                  <select 
                    value={newGradeClass}
                    onChange={(e) => setNewGradeClass(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none"
                  >
                    <option value="Grade 8 East">Grade 8 East</option>
                    <option value="Grade 8 West">Grade 8 West</option>
                  </select>
                </div>
                <div>
                  <label className="block font-semibold text-slate-700 mb-1">Scholar Type</label>
                  <select 
                    value={newScholarType}
                    onChange={(e) => setNewScholarType(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none"
                  >
                    <option value="Boarding Scholar">Boarding Scholar</option>
                    <option value="Day scholar">Day scholar</option>
                  </select>
                </div>
              </div>

              <div className="pt-4 flex items-center justify-end gap-3">
                <button 
                  type="button" 
                  onClick={() => setIsAddModalOpen(false)}
                  className="px-4 py-2.5 border border-slate-200 rounded-xl font-semibold text-slate-600 hover:bg-slate-50"
                >
                  Cancel
                </button>
                <button 
                  type="submit" 
                  disabled={isSubmitting}
                  className="px-5 py-2.5 bg-[#008272] text-white rounded-xl font-semibold hover:bg-[#006e61] transition flex items-center gap-2"
                >
                  {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                  Register Scholar
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Roll-Forward Confirmation Modal */}
      {isRollForwardModalOpen && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-[28px] max-w-md w-full p-6 shadow-2xl space-y-5 text-center">
            <div className="w-12 h-12 bg-purple-100 text-purple-700 rounded-2xl flex items-center justify-center mx-auto">
              <GraduationCap className="w-6 h-6" />
            </div>
            <div className="space-y-2">
              <h3 className="font-bold text-slate-900 text-lg">Execute Grade Roll-Forward?</h3>
              <p className="text-xs text-slate-600 leading-relaxed">
                This will automatically advance 142 eligible scholars to their next grade level based on completed Term 3 CBC evaluations. Audit lock closes Friday at 5:00 PM.
              </p>
            </div>
            <div className="flex items-center gap-3 pt-2">
              <button 
                onClick={() => setIsRollForwardModalOpen(false)}
                className="flex-1 py-3 border border-slate-200 rounded-xl font-semibold text-slate-600 hover:bg-slate-50 text-xs"
              >
                Cancel
              </button>
              <button 
                onClick={handleRollForwardAction}
                disabled={isSubmitting}
                className="flex-1 py-3 bg-[#008272] text-white rounded-xl font-semibold hover:bg-[#006e61] transition text-xs flex items-center justify-center gap-2 shadow-md shadow-[#008272]/20"
              >
                {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                Confirm Advance
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}