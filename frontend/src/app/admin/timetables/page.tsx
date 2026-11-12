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
  Printer, 
  Download, 
  Sliders, 
  Check 
} from 'lucide-react';

interface LessonSlot {
  id: string;
  period: string; // e.g., 'P1', 'P2'
  time: string; // e.g., '8:00 - 8:40'
  day: 'Monday' | 'Tuesday' | 'Wednesday' | 'Thursday' | 'Friday';
  subjectName: string;
  subjectCode: string;
  educator: string;
  room: string;
  category: 'primary' | 'optional' | 'unassigned';
}

interface TimetableData {
  academicTerm: string;
  totalAllocatedPeriods: number;
  termLoadPercentage: number;
  selectedClass: string;
  viewMode: 'class' | 'teacher';
  hasDraftChanges: boolean;
  publishedDate: string;
  lessons: LessonSlot[];
}

export default function AdminTimetablesPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('timetables');

  // Dynamic state replacing hardcoded mocks
  const [timetableData, setTimetableData] = useState<TimetableData | null>(null);
  const [selectedClass, setSelectedClass] = useState('Grade 8 East');
  const [viewMode, setViewMode] = useState<'class' | 'teacher'>('class');
  
  // Selected lesson in the Inspector Panel for editing
  const [selectedLesson, setSelectedLesson] = useState<LessonSlot | null>(null);
  const [assignedEducator, setAssignedEducator] = useState('');
  const [roomAllocation, setRoomAllocation] = useState('');

  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch timetable data from backend API
  useEffect(() => {
    const fetchTimetable = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const res = await fetch(`/api/v1/admin/timetables?class=${encodeURIComponent(selectedClass)}&mode=${viewMode}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize timetable matrix from institutional registry.');
        }

        const data = await res.json();
        setTimetableData(data);

        // Default selected lesson for inspector if available
        if (data.lessons && data.lessons.length > 0 && !selectedLesson) {
          setSelectedLesson(data.lessons[0]);
          setAssignedEducator(data.lessons[0].educator);
          setRoomAllocation(data.lessons[0].room);
        }
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading timetables.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchTimetable();
  }, [selectedClass, viewMode]);

  // Handle saving lesson edits from the Inspector Panel
  const handleSaveLesson = async () => {
    if (!selectedLesson) return;
    setIsSubmitting(true);
    try {
      const res = await fetch(`/api/v1/admin/timetables/lessons/${selectedLesson.id}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          educator: assignedEducator,
          room: roomAllocation,
        }),
      });

      if (!res.ok) throw new Error('Failed to update timetable slot.');

      const updatedLesson = await res.json();
      
      // Update local state
      if (timetableData) {
        setTimetableData({
          ...timetableData,
          hasDraftChanges: true,
          lessons: timetableData.lessons.map(l => l.id === updatedLesson.id ? updatedLesson : l),
        });
      }

      alert('Lesson assignment successfully updated.');
    } catch (err: any) {
      alert(err.message || 'Could not save changes.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle publishing changes
  const handlePublishChanges = async () => {
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/admin/timetables/publish', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) throw new Error('Failed to publish timetable changes.');

      if (timetableData) {
        setTimetableData({
          ...timetableData,
          hasDraftChanges: false,
          publishedDate: 'Just now',
        });
      }

      alert('Timetable successfully published to staff and student portals.');
    } catch (err: any) {
      alert(err.message || 'Publishing failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const daysOfWeek = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'];
  const periods = ['P1', 'P2', 'P3', 'P4', 'P5', 'P6', 'P7'];

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
            <button onClick={() => setActiveNav('timetables')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <Clock className="w-4 h-4 text-[#008272]" /> Timetables
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
          
          {/* Top Academic Administration Bar */}
          <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <div>
                <span className="block text-[10px] font-bold text-slate-400 uppercase tracking-wider">Academic Administration</span>
                <h2 className="text-xl font-bold text-slate-900">Term 3 · 2024</h2>
              </div>

              {/* View Mode Toggle: By Class vs By Teacher */}
              <div className="flex bg-slate-100 p-1 rounded-2xl border border-slate-200 text-xs font-semibold">
                <button 
                  onClick={() => setViewMode('class')}
                  className={`px-3 py-1.5 rounded-xl transition ${viewMode === 'class' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600'}`}
                >
                  By class
                </button>
                <button 
                  onClick={() => setViewMode('teacher')}
                  className={`px-3 py-1.5 rounded-xl transition ${viewMode === 'teacher' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600'}`}
                >
                  By teacher
                </button>
              </div>

              {/* Class Selector Dropdown */}
              <select 
                value={selectedClass}
                onChange={(e) => setSelectedClass(e.target.value)}
                className="bg-slate-50 border border-slate-200 rounded-xl px-3 py-2 text-xs font-semibold text-slate-900 focus:outline-none"
              >
                <option value="Grade 8 East">Grade 8 East</option>
                <option value="Grade 8 West">Grade 8 West</option>
                <option value="Grade 9 East">Grade 9 East</option>
              </select>
            </div>

            <div className="flex items-center gap-3">
              <span className="text-xs font-medium text-emerald-700 bg-emerald-50 px-3 py-1 rounded-full border border-emerald-200">
                🟢 Published {timetableData?.publishedDate || '24 Oct'}
              </span>
              {timetableData?.hasDraftChanges && (
                <span className="text-xs font-medium text-amber-800 bg-amber-50 px-3 py-1 rounded-full border border-amber-200">
                  ⚠️ Draft changes
                </span>
              )}
              <button 
                onClick={handlePublishChanges}
                disabled={isSubmitting}
                className="px-4 py-2 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition shadow-md shadow-[#008272]/20 flex items-center gap-2"
              >
                {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                Publish changes
              </button>
            </div>
          </div>

          {/* Error Banner if API fails */}
          {errorMessage && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-center gap-3 text-xs">
              <AlertCircle className="w-5 h-5 text-rose-600 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {/* Main Content Layout: Timetable Grid (Left) + Inspector Panel (Right) */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Left Column: Weekly Distribution Grid */}
            <div className="lg:col-span-8 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-6">
              
              <div className="flex items-center justify-between pb-3 border-b border-slate-100">
                <div>
                  <h3 className="font-bold text-slate-900 text-base">Weekly Distribution · {selectedClass}</h3>
                  <p className="text-xs text-slate-500">CBC Curriculum Matrix · {timetableData?.totalAllocatedPeriods || 35} Allocated Periods</p>
                </div>
                <div className="flex items-center gap-2">
                  <button className="p-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-600 transition"><Printer className="w-4 h-4" /></button>
                  <button className="p-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-600 transition"><Download className="w-4 h-4" /></button>
                </div>
              </div>

              {isLoading ? (
                <div className="py-48 flex flex-col items-center justify-center gap-3 text-slate-400">
                  <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
                  <p className="text-xs font-medium">Loading institutional timetable grid...</p>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse text-xs">
                    <thead>
                      <tr className="border-b border-slate-200 text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                        <th className="py-3 px-2 w-20">Period</th>
                        {daysOfWeek.map(day => (
                          <th key={day} className="py-3 px-2">{day}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {periods.map((period, pIdx) => {
                        const times = [
                          '8:00 - 8:40', '8:40 - 9:20', '9:20 - 10:00', 
                          '10:20 - 11:00', '11:00 - 11:40', '11:40 - 12:20', '12:20 - 1:00'
                        ];

                        return (
                          <React.Fragment key={period}>
                            <tr className="align-top">
                              <td className="py-3 px-2 font-bold text-slate-900">
                                {period}
                                <span className="block text-[10px] text-slate-400 font-normal">{times[pIdx]}</span>
                              </td>

                              {daysOfWeek.map(day => {
                                const lesson = timetableData?.lessons?.find(l => l.period === period && l.day === day);
                                const isSelected = selectedLesson?.id === lesson?.id;

                                return (
                                  <td key={day} className="py-2 px-1.5">
                                    {lesson ? (
                                      <div 
                                        onClick={() => {
                                          setSelectedLesson(lesson);
                                          setAssignedEducator(lesson.educator);
                                          setRoomAllocation(lesson.room);
                                        }}
                                        className={`p-2.5 rounded-xl border cursor-pointer transition ${
                                          isSelected ? 'border-2 border-[#008272] bg-[#eaf4f3]' : 
                                          lesson.category === 'primary' ? 'bg-[#eaf4f3]/60 border-slate-200 hover:bg-[#eaf4f3]' : 'bg-purple-50/60 border-purple-200 hover:bg-purple-50'
                                        }`}
                                      >
                                        <div className="flex items-center justify-between font-bold text-slate-900 text-[11px]">
                                          <span>{lesson.subjectName}</span>
                                          <span className="text-[9px] px-1 py-0.5 rounded bg-white text-[#008272]">{lesson.subjectCode}</span>
                                        </div>
                                        <p className="text-[10px] text-slate-600 mt-1 truncate">{lesson.educator}</p>
                                      </div>
                                    ) : (
                                      <div className="p-3 rounded-xl border border-dashed border-slate-200 text-center text-[10px] text-slate-400 hover:bg-slate-50 cursor-pointer">
                                        + Assign
                                      </div>
                                    )}
                                  </td>
                                );
                              })}
                            </tr>

                            {/* Morning Break after P3 */}
                            {pIdx === 2 && (
                              <tr>
                                <td colSpan={6} className="py-2 px-3 text-center bg-slate-50 text-[11px] font-bold text-slate-500 uppercase tracking-wider border-y border-slate-100">
                                  ☕ MORNING BREAK • 10:00 – 10:20 (20 MIN)
                                </td>
                              </tr>
                            )}
                          </React.Fragment>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Legend & Term Load Footer */}
              <div className="flex flex-wrap items-center justify-between pt-4 border-t border-slate-100 text-xs text-slate-500">
                <div className="flex items-center gap-4">
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-[#008272]"></span> Primary Core Subject</span>
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-purple-500"></span> Complementary / CBC Optional</span>
                  <span className="flex items-center gap-1.5"><span className="w-2.5 h-2.5 rounded-full bg-slate-300"></span> Unassigned Slot</span>
                </div>
                <strong className="text-slate-900 font-mono">Term load: {timetableData?.totalAllocatedPeriods || 35}/40 periods filled ({timetableData?.termLoadPercentage || 87.5}%)</strong>
              </div>

            </div>

            {/* Right Column: Inspector Panel */}
            <div className="lg:col-span-4 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-6">
              
              <div className="flex items-center justify-between pb-3 border-b border-slate-100">
                <div className="flex items-center gap-2">
                  <Sliders className="w-4 h-4 text-[#008272]" />
                  <span className="font-bold text-slate-900 text-sm">INSPECTOR PANEL</span>
                </div>
                <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded-md text-[10px] font-mono font-bold">Period P1</span>
              </div>

              {selectedLesson ? (
                <div className="space-y-5 text-xs">
                  <div>
                    <h3 className="text-lg font-bold text-slate-900">Edit lesson: {selectedLesson.day} {selectedLesson.time}</h3>
                    <p className="text-slate-500 text-[11px]">Assign accredited staff and primary learning studio</p>
                  </div>

                  {/* Subject Curriculum */}
                  <div className="space-y-1">
                    <div className="flex items-center justify-between">
                      <label className="font-bold text-slate-700">Subject Curriculum</label>
                      <span className="text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded font-semibold">CBC Compliant</span>
                    </div>
                    <div className="bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-semibold text-slate-900 flex items-center justify-between">
                      <span>{selectedLesson.subjectName}</span>
                      <BookOpen className="w-4 h-4 text-slate-400" />
                    </div>
                  </div>

                  {/* Assigned Educator Dropdown */}
                  <div className="space-y-1">
                    <label className="font-bold text-slate-700">Assigned Educator <span className="text-[10px] font-normal text-slate-400">Filtered from L8 roster</span></label>
                    <select 
                      value={assignedEducator}
                      onChange={(e) => setAssignedEducator(e.target.value)}
                      className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-semibold text-slate-900 focus:outline-none"
                    >
                      <option value={assignedEducator}>{assignedEducator}</option>
                      <option value="Mr. Mutua (Math & Science)">Mr. Mutua (Math & Science)</option>
                      <option value="Ms. Wanjiru (Languages)">Ms. Wanjiru (Languages)</option>
                      <option value="Mr. Omondi (Humanities)">Mr. Omondi (Humanities)</option>
                    </select>
                  </div>

                  {/* Facility / Room Dropdown */}
                  <div className="space-y-1">
                    <label className="font-bold text-slate-700">Facility / Room</label>
                    <select 
                      value={roomAllocation}
                      onChange={(e) => setRoomAllocation(e.target.value)}
                      className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 font-semibold text-slate-900 focus:outline-none"
                    >
                      <option value="Room 8E">Room 8E</option>
                      <option value="Lab 2">Lab 2</option>
                      <option value="Main Hall">Main Hall</option>
                      <option value="Studio Art">Studio Art</option>
                    </select>
                  </div>

                  {/* Conflict Checker Notice */}
                  <div className="bg-emerald-50 border border-emerald-200 rounded-2xl p-4 space-y-1 text-emerald-900">
                    <div className="flex items-center gap-1.5 font-bold text-xs">
                      <CheckCircle2 className="w-4 h-4 text-emerald-600" /> No schedule clashes
                    </div>
                    <p className="text-[11px] text-emerald-700 leading-relaxed">
                      {assignedEducator} is free during {selectedLesson.period} on {selectedLesson.day}. {roomAllocation} is clear.
                    </p>
                  </div>

                  {/* Action Buttons */}
                  <div className="pt-2 flex items-center gap-2">
                    <button className="px-4 py-3 bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold rounded-xl transition text-xs">
                      Clear cell
                    </button>
                    <button 
                      onClick={handleSaveLesson}
                      disabled={isSubmitting}
                      className="flex-1 bg-[#008272] hover:bg-[#006e61] text-white font-semibold py-3 rounded-xl transition shadow-md shadow-[#008272]/20 flex items-center justify-center gap-2 text-xs"
                    >
                      {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                      <Check className="w-4 h-4" /> Done
                    </button>
                  </div>

                  {/* Audit Timestamp Footer */}
                  <div className="bg-slate-50 border border-slate-200 rounded-xl p-3 text-[11px] text-slate-500">
                    Last modified by Dean of Studies today at 09:42 AM.
                  </div>
                </div>
              ) : (
                <p className="text-xs text-slate-400 py-16 text-center">Select any period cell on the timetable grid to inspect or edit assignment.</p>
              )}

            </div>

          </div>

        </div>

      </main>

    </div>
  );
}