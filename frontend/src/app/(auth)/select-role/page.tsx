'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import {
  Bell,
  User,
  ArrowRight,
  Users,
  BookOpen,
  Lock,
  HelpCircle,
  LayoutGrid,
  Shield,
  Settings as SettingsIcon,
} from 'lucide-react';
import { isMembership, Membership, roleRoute } from '@/lib/auth';

function roleLabel(role: string): string {
  const lower = role.toLowerCase();
  if (lower.includes('parent')) return 'Parent';
  if (lower.includes('teacher')) return 'Teacher';
  return role.replace(/_/g, ' ').replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function navClass(active: boolean): string {
  return `flex flex-col items-center gap-1 py-1 px-4 rounded-xl transition ${
    active ? 'text-[#008272] font-semibold' : 'hover:text-slate-800'
  }`;
}

export default function SelectRolePage() {
  const router = useRouter();
  const pathname = usePathname();
  const [memberships, setMemberships] = useState<Membership[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    try {
      const stored = sessionStorage.getItem('auth_memberships');
      const current = sessionStorage.getItem('auth_membership');
      const parsed = stored ? JSON.parse(stored) : null;
      const fallback = current ? [JSON.parse(current)] : null;
      const list: unknown[] = Array.isArray(parsed) ? parsed : Array.isArray(fallback) ? fallback : [];

      if (list.length === 0) {
        router.replace('/login');
        return;
      }

      const valid = list.filter(isMembership);
      if (valid.length === 0) {
        setError('Unable to read your school memberships. Please sign in again.');
        return;
      }

      setMemberships(valid);
    } catch {
      setError('Unable to read your school memberships. Please sign in again.');
    } finally {
      setLoading(false);
    }
  }, [router]);

  const handleSelectRole = (membership: Membership) => {
    const route = roleRoute(membership.role);
    if (!route) return;

    sessionStorage.setItem('auth_membership', JSON.stringify(membership));
    router.push(route);
  };

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
            <Link
              href="/notifications"
              aria-label="Open notifications"
              className="p-2 text-slate-600 hover:bg-slate-100 rounded-full transition"
            >
              <Bell className="w-4 h-4" />
            </Link>
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
          ) : memberships.length === 0 ? (
            <div className="rounded-2xl border border-slate-200 bg-slate-50 p-5 text-sm text-slate-600">
              No active school memberships were returned by the backend. Contact the school registrar to get access.
            </div>
          ) : (
            memberships.map((membership) => {
              const route = roleRoute(membership.role);
              const label = roleLabel(membership.role);
              const disabled = !route;

              return (
                <button
                  key={membership.id}
                  type="button"
                  disabled={disabled}
                  aria-disabled={disabled}
                  aria-label={`${membership.school_name}, ${label}`}
                  onClick={() => handleSelectRole(membership)}
                  className={`w-full bg-white border rounded-2xl p-4 text-left transition shadow-sm relative group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#008272] focus-visible:ring-offset-2 ${
                    disabled
                      ? 'border-slate-200/80 opacity-90 cursor-not-allowed bg-slate-50/50'
                      : 'border-slate-200 hover:border-[#008272] hover:shadow-md cursor-pointer'
                  }`}
                >
                  <div className="flex items-center justify-between mb-1">
                    <h2 className="font-bold text-slate-900 text-base">{membership.school_name}</h2>
                    {disabled ? (
                      <Lock className="w-4 h-4 text-slate-400" />
                    ) : (
                      <ArrowRight className="w-4 h-4 text-slate-400 group-hover:text-[#008272] transition" />
                    )}
                  </div>

                  <div className="flex items-center gap-2 mb-2.5">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-700">
                      {label}
                    </span>
                  </div>

                  <div className="flex items-center gap-2 text-xs text-slate-500 pt-1 border-t border-slate-100">
                    {label === 'Parent' && <Users className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                    {label === 'Teacher' && <BookOpen className="w-3.5 h-3.5 text-slate-400 shrink-0" />}
                    {label !== 'Parent' && label !== 'Teacher' && (
                      <Shield className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    )}
                    <span className="break-words">
                      {disabled
                        ? 'No portal is configured for this role'
                        : label === 'Parent'
                          ? 'Guardian account'
                          : 'School membership'}
                    </span>
                  </div>
                </button>
              );
            })
          )}
        </div>

        <div className="pt-2 border-t border-dashed border-slate-300 text-center">
          <button type="button" className="inline-flex items-center gap-1.5 text-xs text-slate-500 hover:text-[#008272] transition font-medium">
            <HelpCircle className="w-3.5 h-3.5 text-slate-400" />
            Missing an institutional profile? Contact school registrar
          </button>
        </div>
      </div>

      <nav aria-label="Primary" className="w-full max-w-md bg-white border border-slate-200/90 rounded-2xl p-2 shadow-lg flex items-center justify-around text-xs font-medium text-slate-500 mb-2">
        <Link href="/" aria-current={pathname === '/' ? 'page' : undefined} className={navClass(pathname === '/')}>
          <LayoutGrid className="w-5 h-5" />
          <span>Home</span>
        </Link>
        <Link href="/select-role" aria-current={pathname === '/select-role' ? 'page' : undefined} className={navClass(pathname === '/select-role')}>
          <Users className="w-5 h-5" />
          <span>Roles</span>
        </Link>
        <Link href="/settings" aria-current={pathname === '/settings' ? 'page' : undefined} className={navClass(pathname === '/settings')}>
          <SettingsIcon className="w-5 h-5" />
          <span>Settings</span>
        </Link>
        <Link href="/notifications" aria-current={pathname === '/notifications' ? 'page' : undefined} className={navClass(pathname === '/notifications')}>
          <Bell className="w-5 h-5" />
          <span>Alerts</span>
        </Link>
      </nav>
    </div>
  );
}
