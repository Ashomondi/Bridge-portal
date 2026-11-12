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
  ChevronLeft, 
  ChevronRight, 
  MapPin, 
  ShieldCheck, 
  LogOut,
  Loader2,
  AlertCircle,
  X
} from 'lucide-react';

interface CalendarEvent {
  id: string;
  title: string;
  date: string; // YYYY-MM-DD
  time: string;
  location: string;
  category: 'Parent communications' | 'Events & milestones' | 'Academic Term';
  targetAudience: string;
}

interface TermDate {
  id: string;
  termName: string;
  startDate: string;
  endDate: string;
  status: 'Active' | 'Upcoming' | 'Completed';
}

export default function AdminCalendarPage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('calendar');
  const [viewTab, setViewTab] = useState<'events' | 'term-dates'>('events');

  // Dynamic state replacing hardcoded mocks
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [termDates, setTermDates] = useState<TermDate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Current calendar view month/year state
  const [currentDate, setCurrentDate] = useState(new Date(2026, 4, 15)); // May 2026 default
  
  // Modal state for adding a new event dynamically
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newEventTitle, setNewEventTitle] = useState('');
  const [newEventDate, setNewEventDate] = useState('');
  const [newEventTime, setNewEventTime] = useState('');
  const [newEventLocation, setNewEventLocation] = useState('');
  const [newEventCategory, setNewEventCategory] = useState<'Parent communications' | 'Events & milestones' | 'Academic Term'>('Events & milestones');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch calendar events and term dates from backend API
  useEffect(() => {
    const fetchCalendarData = async () => {
      setIsLoading(true);
      setErrorMessage(null);
      try {
        const year = currentDate.getFullYear();
        const month = currentDate.getMonth() + 1;

        const res = await fetch(`/api/v1/academic/calendar?year=${year}&month=${month}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
            'Content-Type': 'application/json',
          },
        });

        if (!res.ok) {
          throw new Error('Failed to synchronize calendar data from institutional registry.');
        }

        const data = await res.json();
        setEvents(data.events || []);
        setTermDates(data.termDates || []);
      } catch (err: any) {
        setErrorMessage(err.message || 'An unexpected error occurred while loading schedule data.');
      } finally {
        setIsLoading(false);
      }
    };

    fetchCalendarData();
  }, [currentDate]);

  // Handle creating a new event dynamically via POST API
  const handleCreateEvent = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const res = await fetch('/api/v1/admin/events', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title: newEventTitle,
          date: newEventDate,
          time: newEventTime,
          location: newEventLocation,
          category: newEventCategory,
        }),
      });

      if (!res.ok) {
        throw new Error('Failed to record new event on institutional server.');
      }

      const createdEvent = await res.json();
      setEvents((prev) => [...prev, createdEvent]);
      setIsModalOpen(false);

      // Reset form fields
      setNewEventTitle('');
      setNewEventDate('');
      setNewEventTime('');
      setNewEventLocation('');
    } catch (err: any) {
      alert(err.message || 'Could not save event.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Month navigation helpers
  const handlePrevMonth = () => {
    setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() - 1, 1));
  };

  const handleNextMonth = () => {
    setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 1));
  };

  const monthName = currentDate.toLocaleString('default', { month: 'long', year: 'numeric' });

  // Generate days of the month dynamically
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();
  const firstDayOfMonth = new Date(year, month, 1).getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();

  // Adjust starting index for Monday-start grid
  const startingDayIndex = firstDayOfMonth === 0 ? 6 : firstDayOfMonth - 1;

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
            <button onClick={() => setActiveNav('calendar')} className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl font-semibold bg-[#e0f2f1] text-[#008272] transition">
              <CalendarIcon className="w-4 h-4 text-[#008272]" /> Calendar
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
            Live Institutional Environment
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

        <div className="p-6 md:p-10 space-y-8 max-w-7xl w-full mx-auto">
          
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 pb-2 border-b border-slate-200">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                <span>Lakeview Secondary</span> • <span>Academic Governance</span> • <span className="text-[#008272]">Active Term Cycle</span>
              </div>
              <h2 className="text-3xl font-extrabold text-slate-900 tracking-tight">Calendar</h2>
              <p className="text-sm text-slate-600">
                Manage curriculum schedules, institutional assemblies, and synchronized term boundaries.
              </p>
            </div>

            {/* Events / Term Dates Switcher */}
            <div className="bg-slate-100 p-1 rounded-2xl border border-slate-200 flex items-center text-xs font-semibold">
              <button 
                onClick={() => setViewTab('events')}
                className={`flex items-center gap-1.5 px-4 py-2 rounded-xl transition ${
                  viewTab === 'events' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                <CalendarIcon className="w-3.5 h-3.5 text-[#008272]" /> Events
              </button>
              <button 
                onClick={() => setViewTab('term-dates')}
                className={`flex items-center gap-1.5 px-4 py-2 rounded-xl transition ${
                  viewTab === 'term-dates' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'
                }`}
              >
                <Clock className="w-3.5 h-3.5 text-slate-500" /> Term dates
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

          {/* Conditional View: Events / Calendar Grid OR Term Dates List */}
          {viewTab === 'events' ? (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
              
              {/* Left Column: Dynamic Calendar Grid */}
              <div className="lg:col-span-7 bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-6">
                
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <h3 className="text-xl font-bold text-slate-900 capitalize">{monthName}</h3>
                    <span className="px-2.5 py-1 bg-purple-50 text-purple-700 font-semibold text-xs rounded-full border border-purple-100">
                      Live Synchronized
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="flex bg-slate-100 border border-slate-200 rounded-xl p-0.5">
                      <button onClick={handlePrevMonth} className="p-1.5 hover:bg-white rounded-lg text-slate-600 transition"><ChevronLeft className="w-4 h-4" /></button>
                      <button onClick={() => setCurrentDate(new Date())} className="px-3 py-1 bg-white shadow-sm rounded-lg text-xs font-semibold text-slate-900 transition">Today</button>
                      <button onClick={handleNextMonth} className="p-1.5 hover:bg-white rounded-lg text-slate-600 transition"><ChevronRight className="w-4 h-4" /></button>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-7 text-center text-xs font-bold text-slate-400 uppercase tracking-wider py-1 border-b border-slate-100">
                  <span>Mon</span><span>Tue</span><span>Wed</span><span>Thu</span><span>Fri</span><span>Sat</span><span>Sun</span>
                </div>

                {isLoading ? (
                  <div className="py-24 flex flex-col items-center justify-center gap-3 text-slate-400">
                    <Loader2 className="w-8 h-8 animate-spin text-[#008272]" />
                    <p className="text-xs font-medium">Loading calendar schedule...</p>
                  </div>
                ) : (
                  <div className="grid grid-cols-7 gap-1.5 text-xs">
                    
                    {/* Padding cells for prior month offset */}
                    {Array.from({ length: startingDayIndex }).map((_, i) => (
                      <div key={`empty-${i}`} className="h-20 p-2 rounded-2xl bg-slate-50/30 text-slate-300 border border-transparent"></div>
                    ))}

                    {/* Active Month Days */}
                    {Array.from({ length: daysInMonth }).map((_, i) => {
                      const dayNumber = i + 1;
                      const formattedDayString = `${year}-${String(month + 1).padStart(2, '0')}-${String(dayNumber).padStart(2, '0')}`;
                      
                      // Find events matching this day
                      const dayEvents = events.filter((ev) => ev.date === formattedDayString);
                      const isToday = new Date().toISOString().split('T')[0] === formattedDayString;

                      return (
                        <div 
                          key={`day-${dayNumber}`} 
                          className={`h-20 p-2 rounded-2xl border flex flex-col justify-between transition ${
                            isToday 
                              ? 'border-2 border-[#008272] bg-white shadow-sm font-bold text-slate-900' 
                              : 'bg-[#eaf4f3]/40 border-slate-200/80 text-slate-800'
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <span className={isToday ? 'text-[#008272]' : 'font-semibold'}>{dayNumber}</span>
                            {isToday && <span className="w-1.5 h-1.5 rounded-full bg-[#008272]"></span>}
                          </div>

                          <div className="space-y-1 overflow-hidden">
                            {dayEvents.map((ev) => (
                              <div 
                                key={ev.id} 
                                title={`${ev.title} at ${ev.time}`}
                                className={`text-[9px] font-bold px-1.5 py-0.5 rounded truncate ${
                                  ev.category === 'Parent communications' ? 'bg-purple-100 text-purple-700' :
                                  ev.category === 'Academic Term' ? 'bg-[#008272] text-white' :
                                  'bg-emerald-50 text-[#008272]'
                                }`}
                              >
                                {ev.title}
                              </div>
                            ))}
                          </div>
                        </div>
                      );
                    })}

                  </div>
                )}

                <div className="flex flex-wrap items-center justify-between pt-4 border-t border-slate-100 text-xs text-slate-500">
                  <div className="flex items-center gap-2"><span className="w-2.5 h-2.5 rounded-full bg-[#008272]"></span> Current day marker</div>
                  <div className="flex items-center gap-2"><span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span> Events & milestones</div>
                  <div className="flex items-center gap-2"><span className="w-2.5 h-2.5 rounded-full bg-purple-600"></span> Parent communications</div>
                </div>

              </div>

              {/* Right Column: Dynamic Coming Up & Dean Notice */}
              <div className="lg:col-span-5 space-y-6">
                
                <div className="bg-white border border-slate-200 rounded-[28px] p-6 shadow-sm space-y-5">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-bold text-slate-900 text-base">Coming up</h3>
                      <p className="text-xs text-slate-500">Upcoming curriculum engagements</p>
                    </div>
                    <button 
                      onClick={() => setIsModalOpen(true)}
                      className="px-4 py-2 bg-[#008272] hover:bg-[#006e61] text-white font-semibold rounded-xl text-xs transition flex items-center gap-1.5 shadow-md shadow-[#008272]/20"
                    >
                      <Plus className="w-3.5 h-3.5" /> Add event
                    </button>
                  </div>

                  <div className="space-y-3">
                    {events.length === 0 && !isLoading ? (
                      <p className="text-xs text-slate-400 text-center py-8">No upcoming events scheduled for this period.</p>
                    ) : (
                      events.map((item) => {
                        const eventDateObj = new Date(item.date);
                        const monthStr = eventDateObj.toLocaleString('default', { month: 'short' }).toUpperCase();
                        const dayNum = eventDateObj.getDate();

                        return (
                          <div key={item.id} className="bg-[#f8faf9] border border-slate-200 rounded-2xl p-4 flex items-center justify-between gap-3">
                            <div className="flex items-center gap-3.5">
                              <div className="bg-white border border-slate-200 rounded-xl p-2.5 text-center min-w-[50px] shadow-sm">
                                <span className="block text-[10px] font-bold text-[#008272] uppercase">{monthStr}</span>
                                <span className="block text-lg font-extrabold text-slate-900 leading-none">{dayNum}</span>
                              </div>
                              <div className="space-y-1">
                                <div className="flex items-center gap-2">
                                  <h4 className="font-bold text-slate-900 text-sm">{item.title}</h4>
                                  <span className={`px-2 py-0.5 rounded-md text-[10px] font-semibold ${
                                    item.category === 'Parent communications' ? 'bg-purple-100 text-purple-700' : 'bg-emerald-100 text-[#008272]'
                                  }`}>
                                    {item.targetAudience || item.category}
                                  </span>
                                </div>
                                <div className="flex items-center gap-3 text-[11px] text-slate-500">
                                  <span className="flex items-center gap-1"><Clock className="w-3 h-3 text-slate-400" /> {item.time}</span>
                                  <span className="flex items-center gap-1"><MapPin className="w-3 h-3 text-slate-400" /> {item.location}</span>
                                </div>
                              </div>
                            </div>
                          </div>
                        );
                      })
                    )}
                  </div>
                </div>

                <div className="bg-[#f2f8f7] border border-slate-200/80 rounded-[28px] p-6 space-y-2 shadow-sm">
                  <div className="flex items-center gap-2 text-slate-900 font-bold text-sm">
                    <ShieldCheck className="w-4 h-4 text-[#008272]" /> Dean Synchronized Notice
                  </div>
                  <p className="text-xs text-slate-600 leading-relaxed">
                    Changes to milestone assessments automatically propagate to teacher gradebooks and parent mobile summaries.
                  </p>
                </div>

              </div>

            </div>
          ) : (
            /* Term Dates Tab View */
            <div className="bg-white border border-slate-200 rounded-[28px] p-6 md:p-8 shadow-sm space-y-6">
              <h3 className="text-xl font-bold text-slate-900">Institutional Term Dates Registry</h3>
              <div className="divide-y divide-slate-100">
                {termDates.length === 0 ? (
                  <p className="text-xs text-slate-400 py-8 text-center">No academic term records found.</p>
                ) : (
                  termDates.map((term) => (
                    <div key={term.id} className="py-4 flex items-center justify-between">
                      <div>
                        <h4 className="font-bold text-slate-900 text-base">{term.termName}</h4>
                        <p className="text-xs text-slate-500">Starts: {term.startDate} • Ends: {term.endDate}</p>
                      </div>
                      <span className={`px-3 py-1 rounded-full text-xs font-semibold ${
                        term.status === 'Active' ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-600'
                      }`}>
                        {term.status}
                      </span>
                    </div>
                  ))
                )}
              </div>
            </div>
          )}

        </div>

      </main>

      {/* Add Event Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-[28px] max-w-md w-full p-6 shadow-2xl space-y-6">
            <div className="flex items-center justify-between pb-3 border-b border-slate-100">
              <h3 className="font-bold text-slate-900 text-lg">Add New Calendar Event</h3>
              <button onClick={() => setIsModalOpen(false)} className="p-1 hover:bg-slate-100 rounded-full text-slate-500">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateEvent} className="space-y-4 text-xs">
              <div>
                <label className="block font-semibold text-slate-700 mb-1">Event Title</label>
                <input 
                  type="text" 
                  required
                  value={newEventTitle} 
                  onChange={(e) => setNewEventTitle(e.target.value)}
                  placeholder="e.g. Science Fair & Exhibition" 
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none focus:border-[#008272]"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-semibold text-slate-700 mb-1">Date</label>
                  <input 
                    type="date" 
                    required
                    value={newEventDate} 
                    onChange={(e) => setNewEventDate(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none focus:border-[#008272]"
                  />
                </div>
                <div>
                  <label className="block font-semibold text-slate-700 mb-1">Time</label>
                  <input 
                    type="text" 
                    required
                    value={newEventTime} 
                    onChange={(e) => setNewEventTime(e.target.value)}
                    placeholder="8:00 am" 
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none focus:border-[#008272]"
                  />
                </div>
              </div>

              <div>
                <label className="block font-semibold text-slate-700 mb-1">Location / Venue</label>
                <input 
                  type="text" 
                  required
                  value={newEventLocation} 
                  onChange={(e) => setNewEventLocation(e.target.value)}
                  placeholder="Main Hall" 
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none focus:border-[#008272]"
                />
              </div>

              <div>
                <label className="block font-semibold text-slate-700 mb-1">Category</label>
                <select 
                  value={newEventCategory}
                  onChange={(e) => setNewEventCategory(e.target.value as any)}
                  className="w-full bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5 text-slate-900 focus:outline-none focus:border-[#008272]"
                >
                  <option value="Events & milestones">Events & milestones</option>
                  <option value="Parent communications">Parent communications</option>
                  <option value="Academic Term">Academic Term</option>
                </select>
              </div>

              <div className="pt-4 flex items-center justify-end gap-3">
                <button 
                  type="button" 
                  onClick={() => setIsModalOpen(false)}
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
                  Save Event
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
}