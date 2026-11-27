'use client';

import React, { useMemo, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Bell, 
  User, 
  BarChart2, 
  FileEdit, 
  CreditCard, 
  UserX, 
  Calendar, 
  LayoutGrid, 
  Users, 
  Settings as SettingsIcon 
} from 'lucide-react';

type NotificationSection = 'Today' | 'Earlier this week' | 'Earlier';

interface NotificationItem {
  id: string;
  title: string;
  subtitle: string;
  createdAt: string;
  unread: boolean;
  type: 'results' | 'correction' | 'fee' | 'attendance' | 'exam' | 'term';
  note?: {
    author: string;
    text: string;
  };
}

const hoursAgo = (hours: number) =>
  new Date(Date.now() - hours * 60 * 60 * 1000).toISOString();

const nairobiDateKey = (value: string) =>
  new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Africa/Nairobi',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(value));

const nairobiTimestamp = (value: string) =>
  new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Africa/Nairobi',
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  }).format(new Date(value));

function sectionFor(createdAt: string, todayKey: string): NotificationSection {
  if (nairobiDateKey(createdAt) === todayKey) return 'Today';
  const ageDays = (Date.now() - new Date(createdAt).getTime()) / 86_400_000;
  return ageDays < 7 ? 'Earlier this week' : 'Earlier';
}

// Sample fixtures for the demo portal; replaced by the notifications API when wired.
const SAMPLE_NOTIFICATIONS: NotificationItem[] = [
  {
    id: '1',
    title: 'Term 1 results are ready',
    subtitle: 'Amani Wekesa • Grade 8 East',
    createdAt: hoursAgo(2),
    unread: true,
    type: 'results',
  },
  {
    id: '2',
    title: 'A result was corrected',
    subtitle: 'Integrated Science • Amani Wekesa',
    createdAt: hoursAgo(26),
    unread: true,
    type: 'correction',
    note: {
      author: "Teacher's note",
      text: 'Practical assessment score recalculated by Ms. Wanjiru',
    },
  },
  {
    id: '3',
    title: 'Your Term 2 fee change',
    subtitle: 'Transport charge removed after a route change.',
    createdAt: hoursAgo(40),
    unread: true,
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
    createdAt: hoursAgo(120),
    unread: false,
    type: 'attendance',
  },
  {
    id: '5',
    title: 'Term 2 exams begin soon',
    subtitle: 'Updated by the dean',
    createdAt: hoursAgo(144),
    unread: false,
    type: 'exam',
  },
  {
    id: '6',
    title: 'Term 2 start date moved',
    subtitle: 'Fee due dates moved with it.',
    createdAt: hoursAgo(240),
    unread: false,
    type: 'term',
    note: {
      author: 'Note from the dean',
      text: 'School opening delayed by a week.',
    },
  },
];

function navClass(active: boolean): string {
  return `flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${
    active ? 'text-[#008272] font-semibold' : 'hover:text-slate-800'
  }`;
}

export default function NotificationsPage() {
  const pathname = usePathname();
  const [activeTab, setActiveTab] = useState<'all' | 'unread'>('all');
  const [notifications, setNotifications] = useState<NotificationItem[]>(SAMPLE_NOTIFICATIONS);

  const unreadCount = notifications.filter((n) => n.unread).length;

  const markAllAsRead = () => {
    setNotifications(notifications.map((n) => ({ ...n, unread: false })));
  };

  const groupedNotifications = useMemo(() => {
    const todayKey = nairobiDateKey(new Date().toISOString());
    const filtered = activeTab === 'unread'
      ? notifications.filter((n) => n.unread)
      : notifications;
    const entries = [...filtered].sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
    );

    const groups: { title: NotificationSection; items: NotificationItem[] }[] = [
      { title: 'Today', items: [] },
      { title: 'Earlier this week', items: [] },
      { title: 'Earlier', items: [] },
    ];
    for (const item of entries) {
      groups.find((group) => group.title === sectionFor(item.createdAt, todayKey))?.items.push(item);
    }
    return groups.filter((group) => group.items.length > 0);
  }, [activeTab, notifications]);

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
            <button type="button" aria-label="Notifications" className="p-2 text-slate-600 hover:bg-slate-100 rounded-full transition relative">
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
            type="button"
            onClick={markAllAsRead}
            disabled={unreadCount === 0}
            className="text-xs font-semibold text-[#008272] hover:underline disabled:cursor-not-allowed disabled:opacity-50"
          >
            Mark all as read
          </button>
        </div>

        {/* Filter Tabs (All / Unread) */}
        <div role="tablist" aria-label="Notification filters" className="flex bg-[#f2f8f7] p-1 rounded-2xl border border-slate-200 text-xs font-medium">
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'all'}
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
            type="button"
            role="tab"
            aria-selected={activeTab === 'unread'}
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
          {groupedNotifications.length === 0 ? (
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-6 text-center text-sm text-slate-600">
              {activeTab === 'unread' ? 'You have no unread notifications.' : 'No notifications yet.'}
            </div>
          ) : (
            groupedNotifications.map((group) => (
              <div key={group.title} className="space-y-2.5">
                <p className="text-[10px] font-bold tracking-wider text-slate-400 uppercase px-1">{group.title}</p>

                {group.items.map((item) => (
                  <div key={item.id} className="bg-white border border-slate-200 rounded-2xl p-3.5 shadow-sm space-y-2.5 relative">
                    <div className="flex items-start gap-3">
                      {item.unread && (
                        <span className="w-2 h-2 rounded-full bg-purple-600 shrink-0 mt-2" aria-label="Unread" />
                      )}

                      <div className="p-2.5 rounded-xl shrink-0 bg-purple-50 text-purple-600">
                        {item.type === 'results' && <BarChart2 className="w-4 h-4" />}
                        {item.type === 'correction' && <FileEdit className="w-4 h-4" />}
                        {item.type === 'fee' && <CreditCard className="w-4 h-4" />}
                        {item.type === 'attendance' && <UserX className="w-4 h-4 text-slate-500" />}
                        {(item.type === 'exam' || item.type === 'term') && <Calendar className="w-4 h-4 text-slate-500" />}
                      </div>

                      <div className="flex-1 min-w-0">
                        <div className="flex flex-wrap items-center justify-between gap-1">
                          <h2 className="font-bold text-slate-900 text-sm break-words">{item.title}</h2>
                          <span className="text-[10px] text-slate-400 shrink-0">{nairobiTimestamp(item.createdAt)}</span>
                        </div>
                        <p className="text-xs text-slate-500 break-words">{item.subtitle}</p>
                      </div>
                    </div>

                    {item.note && (
                      <div className="bg-[#f5f3f9] border border-[#e8e2f3] rounded-xl p-3 text-xs text-slate-700 ml-4">
                        <p><span className="font-semibold text-slate-900">{item.note.author}:</span> {item.note.text}</p>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            ))
          )}
        </div>

      </div>

      {/* Bottom Floating Navigation Bar */}
      <nav aria-label="Primary" className="w-full max-w-md bg-white border border-slate-200/90 rounded-2xl p-2 shadow-lg flex items-center justify-around text-xs font-medium text-slate-500 mb-2">
        <Link href="/" aria-current={pathname === '/' ? 'page' : undefined} className={navClass(pathname === '/')}>
          <LayoutGrid className="w-5 h-5" />
          <span>Home</span>
        </Link>
        <Link href="/select-role" aria-current={pathname === '/select-role' ? 'page' : undefined} className={navClass(pathname === '/select-role')}>
          <Users className="w-5 h-5" />
          <span>Roles</span>
        </Link>
        <Link href="/notifications" aria-current={pathname === '/notifications' ? 'page' : undefined} className={navClass(pathname === '/notifications')}>
          <Bell className="w-5 h-5" />
          <span>Alerts</span>
        </Link>
        <Link href="/settings" aria-current={pathname === '/settings' ? 'page' : undefined} className={navClass(pathname === '/settings')}>
          <SettingsIcon className="w-5 h-5" />
          <span>Settings</span>
        </Link>
      </nav>

    </div>
  );
}