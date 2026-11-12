'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import {
  ArrowLeft,
  Bell,
  Check,
  CheckCircle2,
  CircleDashed,
  Clock3,
  FileText,
  Info,
  MapPin,
  Phone,
  RefreshCw,
  ShieldCheck,
  UserRound,
} from 'lucide-react';

const statusSteps = [
  { label: 'Waiting for verification', state: 'pending', detail: 'Step 2 of 3' },
  { label: 'Submitted online', state: 'done', detail: 'Submitted today' },
  { label: 'Verified in person', state: 'done', detail: 'School office confirmation' },
  { label: 'Access granted', state: 'done', detail: 'SMS confirmation and portal access' },
];

const nextSteps = [
  'Visit the school office for the first-time verification appointment.',
  'Bring a valid ID and proof of residence or a signed address confirmation if required.',
  'Complete the in-person review and receive your confirmation message once approved.',
];

export default function ParentRegistrationStatusPage() {
  const router = useRouter();

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#edf4f3] via-[#f4f7f7] to-white px-4 py-6 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center gap-5">
        <div className="w-full max-w-5xl overflow-hidden rounded-[32px] border border-slate-200 bg-white shadow-[0_25px_65px_rgba(15,23,42,0.08)]">
          <div className="flex flex-col lg:flex-row">
            <div className="w-full bg-[#f5f8f7] px-4 py-4 sm:px-6 lg:w-[36%] lg:border-r lg:border-slate-200 lg:px-7 lg:py-7">
              <div className="flex items-center justify-between gap-3 rounded-2xl bg-white/80 px-3 py-2 shadow-sm ring-1 ring-slate-200/80">
                <button
                  type="button"
                  onClick={() => router.push('/login')}
                  className="flex h-9 w-9 items-center justify-center rounded-full border border-slate-200 bg-slate-50 text-slate-700 transition hover:bg-slate-100"
                  aria-label="Go back"
                >
                  <ArrowLeft className="h-4 w-4" />
                </button>

                <div className="flex-1 text-left">
                  <p className="text-sm font-bold text-slate-900">Parent Registration Status</p>
                  <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">Your school</p>
                </div>

                <button
                  type="button"
                  className="flex h-9 w-9 items-center justify-center rounded-full bg-[#0d6e66] text-white shadow-sm shadow-emerald-900/10"
                  aria-label="Notifications"
                >
                  <Bell className="h-4 w-4" />
                </button>
              </div>

              <div className="mt-6 rounded-[26px] border border-slate-200 bg-white p-4 shadow-sm">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-500">
                    <ShieldCheck className="h-3.5 w-3.5 text-[#0d6e66]" />
                    Access code
                  </div>
                  <span className="rounded-full bg-[#eaf5f4] px-2.5 py-1 text-[10px] font-bold tracking-[0.18em] text-[#0d6e66]">
                    REG-XXXX
                  </span>
                </div>

                <div className="mt-5 space-y-4">
                  <div>
                    <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">Status</p>
                    <h2 className="mt-1 text-[28px] font-bold leading-tight tracking-[-0.04em] text-slate-900">
                      We&apos;ve got your registration
                    </h2>
                  </div>

                  <p className="text-sm leading-6 text-slate-600">
                    School portal access request in progress
                  </p>
                </div>
              </div>

              <div className="mt-6 rounded-[22px] border border-slate-200 bg-white p-4 shadow-sm">
                <div className="mb-4 flex items-center justify-between">
                  <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-500">Progress</p>
                  <span className="rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-600">Step 2 of 3</span>
                </div>

                <div className="space-y-4">
                  {statusSteps.map((step, index) => {
                    const isDone = step.state === 'done';
                    const isCurrent = step.state === 'pending';

                    return (
                      <div key={step.label} className="flex items-start gap-3">
                        <div className="relative flex flex-col items-center">
                          <div
                            className={`flex h-7 w-7 items-center justify-center rounded-full border text-[10px] font-bold ${
                              isDone
                                ? 'border-[#0d6e66] bg-[#0d6e66] text-white'
                                : isCurrent
                                  ? 'border-[#0d6e66] bg-[#eaf5f4] text-[#0d6e66]'
                                  : 'border-slate-200 bg-slate-50 text-slate-400'
                            }`}
                          >
                            {isDone ? <Check className="h-3.5 w-3.5" /> : isCurrent ? <Clock3 className="h-3.5 w-3.5" /> : <CircleDashed className="h-3.5 w-3.5" />}
                          </div>
                          {index < statusSteps.length - 1 && (
                            <div className="mt-2 h-8 w-px bg-slate-200" />
                          )}
                        </div>

                        <div className="min-w-0 flex-1 pt-0.5">
                          <div className="flex items-center justify-between gap-3">
                            <p
                              className={`text-sm font-semibold ${
                                isDone ? 'text-slate-900' : isCurrent ? 'text-[#0d6e66]' : 'text-slate-500'
                              }`}
                            >
                              {step.label}
                            </p>
                            {isCurrent && (
                              <span className="rounded-full bg-[#eaf5f4] px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.16em] text-[#0d6e66]">
                                Active
                              </span>
                            )}
                          </div>
                          <p className="mt-1 text-[11px] text-slate-500">{step.detail}</p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>

            <div className="w-full flex-1 px-4 py-4 sm:px-6 lg:px-7 lg:py-7">
              <div className="space-y-5">
                <div className="rounded-[24px] border border-slate-200 bg-[#f8faf9] p-4 shadow-sm">
                  <div className="flex items-start gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[#dff3f0] text-[#0d6e66]">
                      <Info className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">Status note</p>
                      <p className="mt-1 text-sm leading-6 text-slate-600">
                        Your registration has been received and is awaiting the next verification step from the school office.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="rounded-[26px] border border-slate-200 bg-white p-4 shadow-sm">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 text-sm font-bold text-slate-900">
                      <FileText className="h-4 w-4 text-slate-600" />
                      Submitted details
                    </div>
                    <span className="rounded-full bg-emerald-50 px-2 py-1 text-[10px] font-semibold text-emerald-700">Verified</span>
                  </div>

                  <div className="mt-4 space-y-3">
                    <div className="flex items-center justify-between gap-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
                      <div className="flex items-center gap-3">
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[#eaf5f4] text-[#0d6e66]">
                          <UserRound className="h-4 w-4" />
                        </div>
                        <div>
                          <p className="text-[11px] uppercase tracking-[0.14em] text-slate-500">Parent</p>
                          <p className="text-sm font-semibold text-slate-900">Parent name</p>
                        </div>
                      </div>
                      <span className="text-xs font-medium text-slate-500">+254 7xx xxx xxx</span>
                    </div>

                    <div className="flex items-center justify-between gap-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
                      <div className="flex items-center gap-3">
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[#eaf5f4] text-[#0d6e66]">
                          <UserRound className="h-4 w-4" />
                        </div>
                        <div>
                          <p className="text-[11px] uppercase tracking-[0.14em] text-slate-500">Student</p>
                          <p className="text-sm font-semibold text-slate-900">Student name</p>
                        </div>
                      </div>
                      <span className="text-xs font-medium text-slate-500">Grade level</span>
                    </div>

                    <div className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
                      <div className="flex items-center gap-2">
                        <div className="flex h-7 w-7 items-center justify-center rounded-full bg-[#eaf5f4] text-[#0d6e66]">
                          <MapPin className="h-3.5 w-3.5" />
                        </div>
                        <div>
                          <p className="text-[11px] uppercase tracking-[0.14em] text-slate-500">Address</p>
                          <p className="text-sm font-medium text-slate-700">Residential area / city</p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="rounded-[26px] border border-slate-200 bg-white p-4 shadow-sm">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 text-sm font-bold text-slate-900">
                      <CheckCircle2 className="h-4 w-4 text-[#0d6e66]" />
                      What happens next
                    </div>
                    <span className="text-[10px] font-medium uppercase tracking-[0.18em] text-slate-500">Checklist</span>
                  </div>

                  <ol className="mt-4 space-y-3">
                    {nextSteps.map((text, index) => (
                      <li key={text} className="flex gap-3 rounded-xl border border-slate-200 bg-[#f8faf9] p-3">
                        <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[#0d6e66] text-[10px] font-bold text-white">
                          {index + 1}
                        </span>
                        <p className="text-sm leading-6 text-slate-600">{text}</p>
                      </li>
                    ))}
                  </ol>
                </div>

                <div className="rounded-[24px] border border-slate-200 bg-white p-4 shadow-sm">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">School contact</p>
                      <h3 className="mt-1 text-lg font-bold text-slate-900">School reception desk</h3>
                    </div>
                    <div className="rounded-full bg-[#eaf5f4] p-2 text-[#0d6e66]">
                      <Phone className="h-4 w-4" />
                    </div>
                  </div>

                  <div className="mt-4 space-y-3">
                    <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
                      <Phone className="h-4 w-4 text-slate-500" />
                      <span className="text-sm text-slate-700">School phone number</span>
                    </div>
                    <div className="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
                      <MapPin className="h-4 w-4 text-slate-500" />
                      <span className="text-sm text-slate-700">School office hours</span>
                    </div>
                  </div>
                </div>

                <div className="flex flex-col gap-3 sm:flex-row">
                  <button
                    type="button"
                    className="inline-flex flex-1 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-semibold text-slate-700 transition hover:bg-slate-100"
                  >
                    <RefreshCw className="h-4 w-4" />
                    Refresh status
                  </button>
                  <button
                    type="button"
                    className="inline-flex flex-1 items-center justify-center gap-2 rounded-2xl bg-[#0d6e66] px-4 py-3 text-sm font-semibold text-white shadow-[0_12px_28px_rgba(13,110,102,0.25)] transition hover:bg-[#0b5d58]"
                  >
                    Contact school
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
