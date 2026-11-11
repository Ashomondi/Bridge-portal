'use client';

import React, { useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  Bell,
  User,
  ArrowRight,
  Users,
  BookOpen,
  Lock,
  Clock,
  HelpCircle,
  LayoutGrid,
  Shield,
  Settings as SettingsIcon,
} from 'lucide-react';
import { apiFetch } from '@/lib/api';

interface SchoolRole {
  id: string;
  schoolName: string;
  role: 'Parent' | 'Teacher' | 'Waiting for verification';
  details: string;
  subDetails: string;
  type: 'active' | 'pending';
  badgeColor?: string;
  badgeBg?: string;
}

interface MembershipPayload {
  id?: string;
  membership_id?: string;
  school_id?: string;
  school_name?: string;
  schoolName?: string;
  role?: string;
  role_name?: string;
  status?: string;
  type?: string;
  details?: string;
  description?: string;
  label?: string;
  sub_details?: string;
  subDetails?: string;
  last_used?: boolean;
  is_active?: boolean;
  active?: boolean;
  is_pending?: boolean;
  pending?: boolean;
}

function normalizeRole(raw: MembershipPayload): SchoolRole {
  const roleText = (raw.role ?? raw.role_name ?? 'Teacher').toString();
  const schoolName = raw.school_name ?? raw.schoolName ?? 'School';
  const status = raw.type ?? raw.status ?? (raw.pending || raw.is_pending ? 'pending' : 'active');
  const isPending = status === 'pending' || raw.pending || raw.is_pending || raw.is_active === false || raw.active === false;
  const role: SchoolRole['role'] =
    roleText.toLowerCase().includes('parent')
      ? 'Parent'
      : roleText.toLowerCase().includes('teacher')
        ? 'Teacher'
        : 'Waiting for verification';

  return {
    id: raw.id ?? raw.membership_id ?? raw.school_id ?? `${schoolName}-${role}`,
    schoolName,
    role,
    details:
      raw.details ??
      raw.description ??
      raw.label ??
      (role === 'Parent' ? 'guardian account' : 'school membership'),
    subDetails:
      raw.subDetails ??
      raw.sub_details ??
      (raw.last_used ? 'Last used' : ''),
    type: isPending ? 'pending' : 'active',
  };
}

export default function SelectRolePage() {
  const router = useRouter();
  const [activeNav, setActiveNav] = useState('roles');
  const [roles, setRoles] = useState<SchoolRole[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const loadRoles = async () => {
      try {
        const token = localStorage.getItem('auth_token');
        const payload = await apiFetch<{ data?: { memberships?: MembershipPayload[] }; memberships?: MembershipPayload[] }>(
          '/api/v1/auth/memberships',
          token ? { headers: { Authorization: `Bearer ${token}` } } : undefined,
        );

        const memberships = payload?.data?.memberships ?? payload?.memberships ?? [];
        const normalized = Array.isArray(memberships) ? memberships.map(normalizeRole) : [];
        setRoles(normalized);
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Unable to load your school memberships.';
        setError(message);
      } finally {
        setLoading(false);
      }
    };

    loadRoles();
  }, []);

  const handleSelectRole = (roleItem: SchoolRole) => {
    if (roleItem.type === 'pending') return;

    if (roleItem.role === 'Parent') {
      router.push('/parent');
    } else if (roleItem.role === 'Teacher') {
      router.push('/teacher');
    }
  };

  const emptyState = useMemo(
    () => (
      <div className="rounded-2xl border border-slate-200 bg-slate-50 p-5 text-sm text-slate-600">
        No active school memberships were returned by the backend. Contact the school registrar to get access.
      </div>
    ),
    [],
  );

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#e0f2f1] via-[#f4f8f7] to-white flex flex-col items-center justify-between p-4 text-slate-800">
      <div className="w-full max-w-md bg-white/70 backdrop-blur-md border border-slate-200/90 rounded-[32px] p-6 shadow-2xl space-y-6 my-4">
        <div className="flex items-center justify-between pb-3 border-b border-dashed border-slate-300">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-[#008272] flex items-center justify-center text-white font-bold shadow-md shadow-[#008272]/20 text-sm">
              I
            </div>
            <span className="font-bold text-slate-900 tracking-tight">Role Switcher</span>
          </div>
          <div className="flex items-center gap-3">
            <button className="p-2 text-slate-600 hover:bg-slate-100 rounded-full transition">
              <Bell className="w-4 h-4" />
            </button>
            <div className="w-8 h-8 rounded-full bg-slate-900 text-white flex items-center justify-center text-xs font-semibold">
              <User className="w-4 h-4" />
            </div>
          </div>
        </div>

        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">Choose where to go</h1>
          <p className="text-slate-600 text-sm leading-relaxed">
            You have more than one role. Pick one to continue. You can switch later in your account.
          </p>
        </div>

        {error ? (
          <div className="rounded-2xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">{error}</div>
        ) : null}

        <div className="space-y-3.5">
          {loading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((item) => (
                <div key={item} className="animate-pulse rounded-2xl border border-slate-200 bg-slate-50 p-4">
                  <div className="mb-3 h-4 w-24 rounded bg-slate-200" />
                  <div className="h-5 w-2/3 rounded bg-slate-200" />
                  <div className="mt-3 h-3 w-1/2 rounded bg-slate-200" />
                </div>
              ))}
            </div>
          ) : roles.length === 0 ? (
            emptyState
          ) : (
            roles.map((item) => (
              <div
                key={item.id}
                onClick={() => handleSelectRole(item)}
                className={`bg-white border rounded-2xl p-4 transition shadow-sm relative group ${
                  item.type === 'pending'
                    ? 'border-slate-200/80 opacity-90 cursor-not-allowed bg-slate-50/50'
                    : 'border-slate-200 hover:border-[#008272] hover:shadow-md cursor-pointer'
                }`}
              >
                {item.subDetails === 'Last used' && (
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="text-[11px] font-semibold text-purple-600 flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-purple-600"></span> Last used
                    </span>
                    <ArrowRight className="w-4 h-4 text-slate-400 group-hover:text-[#008272] transition" />
                  </div>
                )}

                {item.type === 'active' && item.subDetails !== 'Last used' && (
                  <div className="flex items-center justify-between mb-1">
                    <h2 className="font-bold text-slate-900 text-base">{item.schoolName}</h2>
                    <ArrowRight className="w-4 h-4 text-slate-400 group-hover:text-[#008272] transition" />
                  </div>
                )}

                {item.subDetails === 'Last used' && (
                  <h2 className="font-bold text-slate-900 text-base mb-1">{item.schoolName}</h2>
                )}

                {item.type === 'pending' && (
                  <div className="flex items-center justify-between mb-1">
                    <h2 className="font-bold text-slate-900 text-base">{item.schoolName}</h2>
                    <Lock className="w-4 h-4 text-slate-400" />
                  </div>
                )}

                <div className="flex items-center gap-2 mb-2.5">
                  {item.role === 'Parent' && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-700">
                      Parent
                    </span>
                  )}
                  {item.role === 'Teacher' && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-700">
                      Teacher
                    </span>
                  )}
                  {item.role === 'Waiting for verification' && (
                    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100/80 text-purple-700">
                      <Clock className="w-3 h-3" /> Waiting for verification
                    </span>
                  )}
                </div>

                <div className="flex items-center gap-2 text-xs text-slate-500 pt-1 border-t border-slate-100">
                  {item.role === 'Parent' && <Users className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                  {item.role === 'Teacher' && <BookOpen className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                  {item.type === 'pending' && <Clock className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                  <span className="truncate">{item.details}</span>
                </div>
              </div>
            ))
          )}
        </div>

        <div className="pt-2 border-t border-dashed border-slate-300 text-center">
          <button className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-[#008272] transition font-medium">
            <HelpCircle className="w-3.5 h-3.5 text-slate-400" />
            Missing an institutional profile? Contact school registrar
          </button>
        </div>
      </div>

      <div className="w-full max-w-md bg-white border border-slate-200/90 rounded-2xl p-2 shadow-lg flex items-center justify-around text-xs font-medium text-slate-500 mb-2">
        <button
          onClick={() => setActiveNav('home')}
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'home' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <LayoutGrid className="w-5 h-5" />
          <span>Home</span>
        </button>
        <button
          onClick={() => setActiveNav('roles')}
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'roles' ? 'text-[#008272] font-semibold' : 'hover:text-slate-800'}`}
        >
          <Users className="w-5 h-5" />
          <span>Roles</span>
        </button>
        <button
          onClick={() => setActiveNav('settings')}
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'settings' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <SettingsIcon className="w-5 h-5" />
          <span>Settings</span>
        </button>
        <button
          onClick={() => setActiveNav('shield')}
          className={`flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${activeNav === 'shield' ? 'text-[#008272]' : 'hover:text-slate-800'}`}
        >
          <Shield className="w-5 h-5" />
          <span>Security</span>
        </button>
      </div>
    </div>
  );
}