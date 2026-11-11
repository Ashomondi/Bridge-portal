'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Bell, 
  User, 
  BarChart2, 
  FileEdit, 
  CreditCard, 
  UserX, 
  Calendar, 
  Eye, 
  LayoutGrid, 
  Users, 
  Settings as SettingsIcon 
} from 'lucide-react';

interface NotificationItem {
  id: string;
  title: string;
  subtitle: string;
  timestamp: string;
  unread: boolean;
  section: 'Today' | 'Earlier this week' | 'Earlier';
  type: 'results' | 'correction' | 'fee' | 'attendance' | 'exam' | 'term';
  note?: {
    author: string;
    text: string;
    actionLabel?: string;
  };
}

export default function NotificationsPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<'all' | 'unread'>('all');
  const [activeNav, setActiveNav] = useState('alerts');

  // Sample data mimicking your notifications mockup
  const [notifications, setNotifications] = useState<NotificationItem[]>([
    {
      id: '1',
      title: 'Term 1 results are ready',
      subtitle: 'Amani Wekesa • Grade 8 East',
      timestamp: 'Today, 8:10 am',
      unread: true,
      section: 'Today',
      type: 'results',
    },
    {
      id: '2',
      title: 'A result was corrected',
      subtitle: 'Integrated Science • Amani Wekesa',
      timestamp: 'Yesterday, 4:30 pm',
      unread: true,
      section: 'Today',
      type: 'correction',
      note: {
        author: "Teacher's note",
        text: 'Practical assessment score recalculated by Ms. Wanjiru',
        actionLabel: "Read the teacher's note",
      },
    },
    {
      id: '3',
      title: 'Your Term 2 fee change',
      subtitle: 'Transport charge removed after a route change.',
      timestamp: 'Tuesday, 11:02 am',
      unread: true,
      section: 'Today',
      type: 'fee',
      note: {
        author: 'Note from the school',
        text: 'Transport charge removed after a route change.',
      },
    },
    {
      id: '4',
      title: 'Amani was marked absent',
      subtitle: 'Morning register • Form 2 class teacher',
      timestamp: 'Wed 29 Oct, 9:05 am',
      unread: false,
      section: 'Earlier this week',
      type: 'attendance',
    },
    {
      id: '5',
      title: 'Term 2 exams begin soon',
      subtitle: 'Updated by the dean',
      timestamp: 'Mon 27 Oct, 2:15 pm',
      unread: false,
      section: 'Earlier this week',
      type: 'exam',
    },
    {
      id: '6',
      title: 'Term 2 start date moved',
      subtitle: 'Fee due dates moved with it.',
      timestamp: 'Fri 24 Oct, 10:00 am',
      unread: false,
      section: 'Earlier',
      type: 'term',
      note: {
        author: 'Note from the dean',
        text: 'School opening delayed by a week.',
      },
    },
  ]);

  const unreadCount = notifications.filter((n) => n.unread).length;

  const markAllAsRead = () => {
    setNotifications(notifications.map((n) => ({ ...n, unread: false })));
  };

  const filteredNotifications = activeTab === 'unread' 
    ? notifications.filter((n) => n.unread) 
    : notifications;

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white flex flex-col items-center justify-between p-4 text-slate-800">
      
      {/* Top Demo Banner */}
      <div className="text-xs font-medium text-slate-600 bg-white/80 backdrop-blur px-4 py-1.5 rounded-full border border-slate-200/60 shadow-sm mt-2">
        Demo environment • sample data only
      </div>

      {/* Main Container Card */}
      <div className="w-full max-w-md bg-white/70 backdrop-blur-md border border-slate-200/90 rounded-[32px] p-6 shadow-2xl space-y-5 my-4">
        
        {/* Top Header Bar inside card */}
        <div className="flex items-center justify-between pb-3 border-b border-dashed border-slate-300">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-[#008272] flex items-center justify-center text-white font-bold shadow-md shadow-[#008272]/20 text-sm">
              I
            </div>
            <span className="font-bold text-slate-900 tracking-tight">Notifications</span>
          </div>
          <div className="flex items-center gap-3">
            <button className="p-2 text-slate-600 hover:bg-slate-100 rounded-full transition relative">
              <Bell className="w-4 h-4 text-[#008272]" />
              {unreadCount > 0 && (
                <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-rose-500"></span>
              )}
            </button>
            <div className="w-8 h-8 rounded-full bg-slate-900 text-white flex items-center justify-center text-xs font-semibold">
              <User className="w-4 h-4" />
            </div>
          </div>
        </div>

        {/* Header & Mark all as read */}
        <div className="flex items-center justify-between pt-1">
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">Notifications</h1>
          <button 
            onClick={markAllAsRead}
            className="text-xs font-semibold text-[#008272] hover:underline"
          >
            Mark all as read
          </button>
        </div>

        {/* Filter Tabs (All / Unread) */}
        <div className="flex bg-[#f2f8f7] p-1 rounded-2xl border border-slate-200 text-xs font-medium">
          <button
            onClick={() => setActiveTab('all')}
            className={`flex-1 py-2 rounded-xl transition ${
              activeTab === 'all' 
                ? 'bg-white text-slate-900 font-semibold shadow-sm border border-slate-200/60' 
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            All
          </button>
          <button
            onClick={() => setActiveTab('unread')}
            className={`flex-1 py-2 rounded-xl transition flex items-center justify-center gap-1.5 ${
              activeTab === 'unread' 
                ? 'bg-white text-slate-900 font-semibold shadow-sm border border-slate-200/60' 
                : 'text-slate-600 hover:text-slate-900'
            }`}
          >
            Unread 
            <span className={`px-1.5 py-0.2 rounded-full text-[10px] ${activeTab === 'unread' ? 'bg-[#008272] text-white' : 'bg-slate-200 text-slate-700'}`}>
              {unreadCount}
            </span>
          </button>
        </div>

        {/* Notifications List Content Grouped by Sections */}
        <div className="space-y-4">
          
          {/* Today Section */}
          <div className="space-y-2.5">
            <p className="text-[10px] font-bold tracking-wider text-slate-400 uppercase px-1">Today</p>
            
            {filteredNotifications.filter(n => n.section === 'Today').map((item) => (
              <div key={item.id} className="bg-white border border-slate-200 rounded-2xl p-3.5 shadow-sm space-y-2.5 relative">
                <div className="flex items-start gap-3">
                  {/* Unread dot */}
                  {item.unread && (
                    <span className="w-2 h-2 rounded-full bg-purple-600 shrink-0 mt-2"></span>
                  )}
                  
                  {/* Icon Box */}
                  <div className={`p-2.5 rounded-xl shrink-0 ${
                    item.type === 'results' ? 'bg-purple-100 text-purple-700' :
                    item.type === 'correction' ? 'bg-purple-100 text-purple-700' : 'bg-purple-50 text-purple-600'
                  }`}>
                    {item.type === 'results' && <BarChart2 className="w-4 h-4" />}
                    {item.type === 'correction' && <FileEdit className="w-4 h-4" />}
                    {item.type === 'fee' && <CreditCard className="w-4 h-4" />}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-1">
                      <h2 className="font-bold text-slate-900 text-sm truncate">{item.title}</h2>
                      <span className="text-[10px] text-slate-400 shrink-0">{item.timestamp}</span>
                    </div>
                    <p className="text-xs text-slate-500 truncate">{item.subtitle}</p>
                  </div>
                </div>

                {/* Optional nested note card */}
                {item.note && (
                  <div className="bg-[#f5f3f9] border border-[#e8e2f3] rounded-xl p-3 space-y-2 text-xs text-slate-700 ml-4">
                    <p><span className="font-semibold text-slate-900">{item.note.author}:</span> {item.note.text}</p>
                    {item.note.actionLabel && (
                      <button className="inline-flex items-center gap-1.5 text-xs font-semibold text-purple-700 hover:underline pt-1">
                        <Eye className="w-3.5 h-3.5" /> {item.note.actionLabel}
                      </button>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Earlier This Week Section */}
          <div className="space-y-2.5 pt-2">
            <p className="text-[10px] font-bold tracking-wider text-slate-400 uppercase px-1">Earlier This Week</p>
            
            {filteredNotifications.filter(n => n.section === 'Earlier this week').map((item) => (
              <div key={item.id} className="bg-white/80 border border-slate-200/80 rounded-2xl p-3.5 shadow-sm flex items-start gap-3">
                <div className="p-2.5 rounded-xl shrink-0 bg-slate-100 text-slate-600 mt-0.5">
                  {item.type === 'attendance' && <UserX className="w-4 h-4 text-slate-500" />}
                  {item.type === 'exam' && <Calendar className="w-4 h-4 text-slate-500" />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-1">
                    <h2 className="font-semibold text-slate-800 text-sm truncate">{item.title}</h2>
                    <span className="text-[10px] text-slate-400 shrink-0">{item.timestamp}</span>
                  </div>
                  <p className="text-xs text-slate-500 truncate">{item.subtitle}</p>
                </div>
              </div>
            ))}
          </div>

          {/* Earlier Section */}
          <div className="space-y-2.5 pt-2">
            <p className="text-[10px] font-bold tracking-wider text-slate-400 uppercase px-1">Earlier</p>
            
            {filteredNotifications.filter(n => n.section === 'Earlier').map((item) => (
              <div key={item.id} className="bg-white/80 border border-slate-200/80 rounded-2xl p-3.5 shadow-sm space-y-2.5">
                <div className="flex items-start gap-3">
                  <div className="p-2.5 rounded-xl shrink-0 bg-slate-100 text-slate-600 mt-0.5">
                    <Calendar className="w-4 h-4 text-slate-500" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-1">
                      <h2 className="font-semibold text-slate-800 text-sm truncate">{item.title}</h2>
                      <span className="text-[10px] text-slate-400 shrink-0">{item.timestamp}</span>
                    </div>
                    <p className="text-xs text-slate-500 truncate">{item.subtitle}</p>
                  </div>
                </div>

                {item.note && (
                  <div className="bg-[#eaf4f2] border border-[#d0e8e4] rounded-xl p-3 text-xs text-slate-700 ml-4">
                    <p><span className="font-semibold text-slate-900">{item.note.author}:</span> {item.note.text}</p>
                  </div>
                )}
              </div>
            ))}
          </div>

        </div>

      </div>

      {/* Bottom Floating Navigation Bar */}
      <div className="w-full max-w-md bg-white border border-slate-200/90 rounded-2xl p-2 shadow-lg flex items-center justify-around text-xs font-medium text-slate-500 mb-2">
        <button 
          onClick={() => setActiveNav('home')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'home' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <LayoutGrid className="w-5 h-5" />
          <span>Home</span>
        </button>
        <button 
          onClick={() => { setActiveNav('roles'); router.push('/select-role'); }} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'roles' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <Users className="w-5 h-5" />
          <span>Roles</span>
        </button>
        <button 
          onClick={() => setActiveNav('alerts')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'alerts' ? 'text-[#008272] font-semibold' : 'hover:text-slate-800'}`}
        >
          <Bell className="w-5 h-5" />
          <span>Alerts</span>
        </button>
        <button 
          onClick={() => setActiveNav('settings')} 
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'settings' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <SettingsIcon className="w-5 h-5" />
          <span>Settings</span>
        </button>
      </div>

    </div>
  );
}