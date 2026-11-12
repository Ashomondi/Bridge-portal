'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Bell,
  User,
  BarChart2,
  FileEdit,
  CreditCard,
  Calendar,
  Eye,
  LayoutGrid,
  Users,
  Settings as SettingsIcon,
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

  const [notifications, setNotifications] = useState<NotificationItem[]>([
    {
      id: '1',
      title: 'Term 1 results are ready',
      subtitle: 'Student results • Class update',
      timestamp: 'Today, 8:10 am',
      unread: true,
      section: 'Today',
      type: 'results',
    },
    {
      id: '2',
      title: 'A result was corrected',
      subtitle: 'Assessment update • Classroom review',
      timestamp: 'Yesterday, 4:30 pm',
      unread: true,
      section: 'Today',
      type: 'correction',
      note: {
        author: "Teacher's note",
        text: 'A score was revised after an assessment review.',
        actionLabel: 'Read the update',
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
      title: 'A student was marked absent',
      subtitle: 'Morning register • Class teacher record',
      timestamp: 'Wed 29 Oct, 9:05 am',
      unread: false,
      section: 'Earlier this week',
      type: 'attendance',
    },
    {
      id: '5',
      title: 'Term 2 exams begin soon',
      subtitle: 'Updated by the school office',
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
        author: 'Note from the school',
        text: 'School opening delayed by a week.',
      },
    },
  ]);

  const unreadCount = notifications.filter((n) => n.unread).length;

  const markAllAsRead = () => setNotifications(notifications.map((n) => ({ ...n, unread: false })));

  const filteredNotifications = activeTab === 'unread' ? notifications.filter((n) => n.unread) : notifications;

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white px-4 py-6 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center gap-6">
        <div className="grid w-full overflow-hidden rounded-[32px] border border-slate-200/90 bg-white/70 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur-md lg:grid-cols-[0.78fr_1.22fr]">
          <aside className="border-b border-slate-200 bg-[radial-gradient(circle_at_top_left,_rgba(0,130,114,0.18),_transparent_30%),linear-gradient(135deg,#0c2b35_0%,#113a3d_35%,#0d5258_100%)] p-6 text-white lg:border-b-0 lg:border-r lg:p-8">
            <div className="flex items-center justify-between gap-4">
              <div className="inline-flex items-center gap-3 rounded-2xl border border-white/15 bg-white/5 px-3 py-2 backdrop-blur-sm">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#008272] text-white shadow-md shadow-[#008272]/30">
                  <Bell className="h-5 w-5" />
                </div>
                <div>
                  <p className="text-[10px] uppercase tracking-[0.18em] text-emerald-100/80">Inbox</p>
                  <p className="text-sm font-semibold">Ikonex Alerts</p>
                </div>
              </div>
              <div className="rounded-full border border-white/15 bg-white/5 px-3 py-1.5 text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-100">
                live
              </div>
            </div>

            <div className="mt-8 space-y-5">
              <div>
                <p className="text-[11px] uppercase tracking-[0.18em] text-emerald-100/80">Summary</p>
                <h1 className="mt-2 text-3xl font-bold tracking-tight text-white">Your updates</h1>
              </div>

              <div className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm text-slate-200">Unread</span>
                  <span className="text-2xl font-bold text-white">{unreadCount}</span>
                </div>
              </div>

              <div className="space-y-3 text-sm text-slate-200/85">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-emerald-400" />
                  Academic and fee updates
                </div>
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-violet-400" />
                  Attendance and exam notices
                </div>
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-amber-400" />
                  Institutional announcements
                </div>
              </div>
            </div>
          </aside>

          <main className="bg-white p-5 sm:p-6 lg:p-8">
            <div className="mx-auto max-w-2xl">
              <div className="flex items-center justify-between gap-3 border-b border-dashed border-slate-300 pb-4">
                <div className="flex items-center gap-2">
                  <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-[#008272] text-sm font-bold text-white shadow-md shadow-[#008272]/20">
                    I
                  </div>
                  <span className="font-bold tracking-tight text-slate-900">Notifications</span>
                </div>
                <div className="flex items-center gap-3">
                  <button className="relative rounded-full p-2 text-slate-600 transition hover:bg-slate-100">
                    <Bell className="h-4 w-4 text-[#008272]" />
                    {unreadCount > 0 && <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-rose-500" />}
                  </button>
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-slate-900 text-white">
                    <User className="h-4 w-4" />
                  </div>
                </div>
              </div>

              <div className="mt-5 flex items-center justify-between gap-3">
                <h2 className="text-2xl font-bold tracking-tight text-slate-900">Notifications</h2>
                <button onClick={markAllAsRead} className="text-xs font-semibold text-[#008272] hover:underline">
                  Mark all as read
                </button>
              </div>

              <div className="mt-5 flex rounded-2xl border border-slate-200 bg-[#f2f8f7] p-1 text-xs font-medium">
                <button
                  onClick={() => setActiveTab('all')}
                  className={`flex-1 rounded-xl py-2 transition ${activeTab === 'all' ? 'border border-slate-200/60 bg-white font-semibold text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
                >
                  All
                </button>
                <button
                  onClick={() => setActiveTab('unread')}
                  className={`flex flex-1 items-center justify-center gap-1.5 rounded-xl py-2 transition ${activeTab === 'unread' ? 'border border-slate-200/60 bg-white font-semibold text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-900'}`}
                >
                  Unread
                  <span className={`rounded-full px-1.5 py-0.2 text-[10px] ${activeTab === 'unread' ? 'bg-[#008272] text-white' : 'bg-slate-200 text-slate-700'}`}>
                    {unreadCount}
                  </span>
                </button>
              </div>

              <div className="mt-5 space-y-4">
                {filteredNotifications.filter((n) => n.section === 'Today').length > 0 && (
                  <div className="space-y-2.5">
                    <p className="px-1 text-[10px] font-bold uppercase tracking-wider text-slate-400">Today</p>
                    {filteredNotifications.filter((n) => n.section === 'Today').map((item) => (
                      <div key={item.id} className="rounded-2xl border border-slate-200 bg-white p-3.5 shadow-sm">
                        <div className="flex items-start gap-3">
                          {item.unread && <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-purple-600" />}
                          <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${item.type === 'results' || item.type === 'correction' ? 'bg-purple-100 text-purple-700' : 'bg-purple-50 text-purple-600'}`}>
                            {item.type === 'results' && <BarChart2 className="h-4 w-4" />}
                            {item.type === 'correction' && <FileEdit className="h-4 w-4" />}
                            {item.type === 'fee' && <CreditCard className="h-4 w-4" />}
                          </div>

                          <div className="min-w-0 flex-1">
                            <div className="flex items-center justify-between gap-2">
                              <h3 className="truncate text-sm font-bold text-slate-900">{item.title}</h3>
                              <span className="shrink-0 text-[10px] text-slate-400">{item.timestamp}</span>
                            </div>
                            <p className="truncate text-xs text-slate-500">{item.subtitle}</p>
                          </div>
                        </div>

                        {item.note && (
                          <div className="ml-5 mt-3 rounded-xl border border-[#e8e2f3] bg-[#f5f3f9] p-3 text-xs text-slate-700">
                            <p><span className="font-semibold text-slate-900">{item.note.author}:</span> {item.note.text}</p>
                            {item.note.actionLabel && (
                              <button className="mt-2 inline-flex items-center gap-1.5 text-xs font-semibold text-purple-700 hover:underline">
                                <Eye className="h-3.5 w-3.5" /> {item.note.actionLabel}
                              </button>
                            )}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}

                {filteredNotifications.filter((n) => n.section === 'Earlier this week').length > 0 && (
                  <div className="space-y-2.5 pt-2">
                    <p className="px-1 text-[10px] font-bold uppercase tracking-wider text-slate-400">Earlier this week</p>
                    {filteredNotifications.filter((n) => n.section === 'Earlier this week').map((item) => (
                      <div key={item.id} className="rounded-2xl border border-slate-200 bg-white p-3.5 shadow-sm">
                        <div className="flex items-start gap-3">
                          <div className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-600">
                            <Calendar className="h-4 w-4" />
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center justify-between gap-2">
                              <h3 className="truncate text-sm font-bold text-slate-900">{item.title}</h3>
                              <span className="shrink-0 text-[10px] text-slate-400">{item.timestamp}</span>
                            </div>
                            <p className="truncate text-xs text-slate-500">{item.subtitle}</p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {filteredNotifications.filter((n) => n.section === 'Earlier').length > 0 && (
                  <div className="space-y-2.5 pt-2">
                    <p className="px-1 text-[10px] font-bold uppercase tracking-wider text-slate-400">Earlier</p>
                    {filteredNotifications.filter((n) => n.section === 'Earlier').map((item) => (
                      <div key={item.id} className="rounded-2xl border border-slate-200 bg-white p-3.5 shadow-sm">
                        <div className="flex items-start gap-3">
                          <div className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-600">
                            <Calendar className="h-4 w-4" />
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center justify-between gap-2">
                              <h3 className="truncate text-sm font-bold text-slate-900">{item.title}</h3>
                              <span className="shrink-0 text-[10px] text-slate-400">{item.timestamp}</span>
                            </div>
                            <p className="truncate text-xs text-slate-500">{item.subtitle}</p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </main>
        </div>

        <div className="w-full max-w-md rounded-2xl border border-slate-200/90 bg-white p-2 shadow-lg lg:max-w-lg">
          <div className="flex items-center justify-around text-xs font-medium text-slate-500">
            <button onClick={() => router.push('/')} className={`flex flex-col items-center gap-1 rounded-xl px-4 py-1 transition ${activeNav === 'home' ? 'text-[#008272]' : 'hover:text-slate-800'}`}>
              <LayoutGrid className="h-5 w-5" />
              <span>Home</span>
            </button>
            <button onClick={() => setActiveNav('alerts')} className={`flex flex-col items-center gap-1 rounded-xl px-4 py-1 transition ${activeNav === 'alerts' ? 'font-semibold text-[#008272]' : 'hover:text-slate-800'}`}>
              <Bell className="h-5 w-5" />
              <span>Alerts</span>
            </button>
            <button onClick={() => setActiveNav('settings')} className={`flex flex-col items-center gap-1 rounded-xl px-4 py-1 transition ${activeNav === 'settings' ? 'text-[#008272]' : 'hover:text-slate-800'}`}>
              <SettingsIcon className="h-5 w-5" />
              <span>Settings</span>
            </button>
            <button onClick={() => setActiveNav('members')} className={`flex flex-col items-center gap-1 rounded-xl px-4 py-1 transition ${activeNav === 'members' ? 'text-[#008272]' : 'hover:text-slate-800'}`}>
              <Users className="h-5 w-5" />
              <span>Members</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}