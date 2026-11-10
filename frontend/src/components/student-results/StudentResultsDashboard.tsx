"use client";

import { useState } from "react";
import {
  Award,
  BarChart3,
  Bell,
  CalendarCheck,
  ChevronRight,
  FileText,
  GraduationCap,
  Home,
  Info,
  LogOut,
  MessageCircle,
  Quote,
  ShieldCheck,
  Sparkles,
  TrendingUp,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "#" },
  { label: "Results", icon: Award, href: "/parent/results", active: true },
  { label: "Attendance", icon: CalendarCheck, href: "#" },
  { label: "Fees", icon: Wallet, href: "#" },
  { label: "AI assistant", icon: Sparkles, href: "#" },
];

const BOTTOM_NAV = [
  { label: "Home", icon: Home, href: "#" },
  { label: "Results", icon: BarChart3, href: "/parent/results", active: true },
  { label: "Fees", icon: Wallet, href: "#" },
  { label: "Assistant", icon: MessageCircle, href: "#" },
];

const STUDENTS = [
  { id: "student-1", label: "[Student Name]", initials: "[AB]" },
  { id: "student-2", label: "[Student]", initials: "[CD]" },
  { id: "student-3", label: "[Student]", initials: "[EF]" },
];

const SUBJECT_ROWS = [
  {
    id: "subject-row-1",
    subject: "[Subject Name]",
    level: "[Performance Level]",
    corrected: false,
  },
  {
    id: "subject-row-2",
    subject: "[Subject Name]",
    level: "[Performance Level]",
    corrected: true,
  },
  {
    id: "subject-row-3",
    subject: "[Subject Name]",
    level: "[Performance Level]",
    corrected: false,
  },
  {
    id: "subject-row-4",
    subject: "[Subject Name]",
    level: "[Performance Level]",
    corrected: false,
  },
  {
    id: "subject-row-5",
    subject: "[Subject Name]",
    level: "[Performance Level]",
    corrected: false,
  },
];

const LEVEL_CARDS = [
  {
    id: "level-1",
    level: "Level 1",
    title: "Below expectation",
    description: "[Description]",
    className: "bg-slate-100 text-slate-600",
  },
  {
    id: "level-2",
    level: "Level 2",
    title: "Approaching",
    description: "[Description]",
    className: "bg-[#F0EBF9] text-[#6D4FA3]",
  },
  {
    id: "level-3",
    level: "Level 3",
    title: "Meeting",
    description: "[Description]",
    className: "bg-[#D8CBF0] text-[#4F3585]",
  },
  {
    id: "level-4",
    level: "Level 4",
    title: "Exceeding",
    description: "[Description]",
    className: "bg-[#4E3A8A] text-white",
  },
];

export default function StudentResultsDashboard() {
  const [selectedStudent, setSelectedStudent] = useState(STUDENTS[0].id);

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#results-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to results
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar />

      <div className="lg:pl-[260px]">
        <DesktopHeader />

        <main
          id="results-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <StudentFilter
            selectedStudent={selectedStudent}
            onSelect={setSelectedStudent}
          />
          <TermSelector />

          <div className="mt-4 grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(360px,1fr)]">
            <div className="min-w-0 space-y-4">
              <SubjectPerformanceCard />
              <AdjustmentNotice />
              <TeacherComment />
            </div>

            <div className="min-w-0 xl:sticky xl:top-[84px] xl:self-start">
              <LevelsCard />
            </div>
          </div>
        </main>
      </div>

      <BottomNav />
    </div>
  );
}

function SidebarContent() {
  return (
    <div className="flex h-full w-full flex-col overflow-y-auto px-4 py-5">
      <div className="flex items-center gap-2.5 px-1">
        <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-brand-500 text-white shadow-lg shadow-brand-500/20">
          <GraduationCap className="h-[22px] w-[22px]" />
        </span>
        <div>
          <p className="text-[15px] font-bold tracking-tight text-slate-900">
            Ikonex <span className="text-brand-600">Academy</span>
          </p>
          <div className="mt-0.5 flex items-center gap-1.5">
            <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
              Parent portal
            </p>
            <span className="rounded-full bg-brand-50 px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wide text-brand-700 ring-1 ring-brand-100">
              Parent
            </span>
          </div>
        </div>
      </div>

      <nav className="mt-6 space-y-0.5" aria-label="Primary">
        {SIDEBAR_ITEMS.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-[12px] font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                item.active
                  ? "bg-[#E8F5F3] font-semibold text-brand-700 ring-1 ring-brand-100"
                  : "text-slate-500 hover:bg-slate-50 hover:text-slate-800",
              )}
            >
              <Icon
                className={cn(
                  "h-[17px] w-[17px] shrink-0",
                  item.active
                    ? "text-brand-600"
                    : "text-slate-400 group-hover:text-brand-600",
                )}
              />
              <span className="flex-1">{item.label}</span>
              {item.active && (
                <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
              )}
            </a>
          );
        })}
      </nav>

      <div className="mt-auto pt-6">
        <div className="rounded-2xl border border-slate-200/80 bg-slate-50/70 p-3">
          <div className="flex items-center gap-2.5">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
              [AB]
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                [Guardian Name]
              </p>
              <p className="truncate text-[10px] text-slate-400">
                [Relationship]
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                [Email / Phone]
              </p>
            </div>
          </div>
          <button
            type="button"
            className="mt-3 inline-flex h-9 w-full items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-[11px] font-semibold text-slate-600 transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <LogOut className="h-3.5 w-3.5" />
            Sign out
          </button>
        </div>
      </div>
    </div>
  );
}

function MobileAppBar() {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <div className="min-w-0 flex-1">
          <p className="truncate text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            [Organization Name]
          </p>
          <p className="truncate text-lg font-bold tracking-tight text-slate-900">
            Results
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          <button
            type="button"
            className="relative grid h-10 w-10 place-items-center rounded-xl text-slate-500 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Notifications"
          >
            <Bell className="h-5 w-5" />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Open user profile"
          >
            [AB]
          </button>
        </div>
      </div>
    </header>
  );
}

function DesktopHeader() {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="max-w-[220px] truncate text-slate-700">
            [Portal Name]
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Results</span>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            type="button"
            className="relative grid h-10 w-10 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Notifications"
          >
            <Bell className="h-[18px] w-[18px]" />
            <span className="absolute right-2.5 top-2.5 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>

          <button
            type="button"
            className="flex items-center gap-2.5 rounded-xl border border-slate-200/80 bg-white py-1.5 pl-1.5 pr-3 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-[10px] font-bold text-white">
              [AB]
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                [Guardian Name]
              </p>
              <p className="text-[10px] text-slate-400">[Relationship]</p>
            </div>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

function StudentFilter({
  selectedStudent,
  onSelect,
}: {
  selectedStudent: string;
  onSelect: (id: string) => void;
}) {
  return (
    <div
      className="-mx-4 overflow-x-auto px-4 pb-1 sm:mx-0 sm:px-0"
      role="tablist"
      aria-label="Student records"
    >
      <div className="flex min-w-max gap-2">
        {STUDENTS.map((student) => {
          const active = student.id === selectedStudent;

          return (
            <button
              key={student.id}
              type="button"
              role="tab"
              aria-selected={active}
              onClick={() => onSelect(student.id)}
              className={cn(
                "inline-flex items-center gap-2 rounded-full px-3 py-2 text-[11px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                active
                  ? "bg-brand-500 text-white shadow-md shadow-brand-500/20"
                  : "border border-slate-200/80 bg-white text-slate-500 hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700",
              )}
            >
              <span
                className={cn(
                  "grid h-6 w-6 place-items-center rounded-full text-[8px] font-bold",
                  active
                    ? "bg-white/20 text-white"
                    : "bg-slate-100 text-slate-500",
                )}
              >
                {student.initials}
              </span>
              {student.label}
            </button>
          );
        })}
      </div>
    </div>
  );
}

function TermSelector() {
  return (
    <div className="mt-3 rounded-full bg-[#E8F5F3] p-1">
      <div className="grid grid-cols-3 gap-1" role="tablist" aria-label="Terms">
        <button
          type="button"
          role="tab"
          aria-selected="true"
          className="rounded-full bg-white px-2 py-2 text-[10px] font-semibold text-brand-700 shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          Term 1
        </button>
        <button
          type="button"
          role="tab"
          aria-selected="false"
          disabled
          className="rounded-full px-2 py-2 text-[10px] font-medium text-slate-400 disabled:cursor-not-allowed"
        >
          Term 2 · Not released
        </button>
        <button
          type="button"
          role="tab"
          aria-selected="false"
          disabled
          className="rounded-full px-2 py-2 text-[10px] font-medium text-slate-400 disabled:cursor-not-allowed"
        >
          Term 3 · Not released
        </button>
      </div>
    </div>
  );
}

function SubjectPerformanceCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Subject performance levels
          </h2>
          <p className="mt-1 text-[10px] text-slate-400">
            Released [Date] · [X] learning areas
          </p>
        </div>
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <BarChart3 className="h-4 w-4" />
        </span>
      </div>

      <div className="mt-3">
        {SUBJECT_ROWS.map((row) => (
          <div
            key={row.id}
            className="flex items-center justify-between gap-3 border-b border-slate-100 py-3 last:border-0"
          >
            <div className="min-w-0">
              <div className="flex items-center gap-1.5">
                <p className="truncate text-[11px] font-bold text-slate-800">
                  {row.subject}
                </p>
                {row.corrected && (
                  <span className="shrink-0 rounded-full bg-[#FEF3D6] px-1.5 py-0.5 text-[8px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
                    Corrected
                  </span>
                )}
              </div>
              <p className="mt-0.5 flex items-center gap-1 text-[9px] font-medium text-slate-400">
                <TrendingUp className="h-3 w-3 text-brand-500" />
                {row.level}
              </p>
            </div>

            <div className="flex shrink-0 items-center gap-0.5 rounded-full bg-slate-50 p-0.5 ring-1 ring-slate-200/80">
              <TermPill label="T1" active />
              <TermPill label="T2" />
              <TermPill label="T3" />
            </div>
          </div>
        ))}
      </div>

      <div className="mt-3 rounded-xl bg-[#E2F0ED] px-3 py-2.5">
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-[9px] font-medium text-brand-700">
          <span>T1 · Term 1</span>
          <span>T2 · Term 2</span>
          <span>T3 · Term 3</span>
        </div>
      </div>
    </section>
  );
}

function TermPill({ label, active = false }: { label: string; active?: boolean }) {
  return (
    <span
      className={cn(
        "grid h-6 min-w-7 place-items-center rounded-full px-1.5 text-[8px] font-bold",
        active ? "bg-brand-500 text-white" : "text-slate-300",
      )}
    >
      {label}
    </span>
  );
}

function AdjustmentNotice() {
  return (
    <section className="rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4">
      <div className="flex gap-3">
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
          <FileText className="h-4 w-4" />
        </span>
        <div className="min-w-0 flex-1">
          <p className="text-xs font-bold text-slate-800">
            Grade adjustment recorded
          </p>
          <p className="mt-1 text-[11px] leading-relaxed text-slate-500">
            [Adjustment note description] A correction was applied to
            [Subject Name].
          </p>
          <div className="mt-3 flex items-center justify-between gap-3">
            <button
              type="button"
              className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-[#D7C8F2] bg-white px-2.5 text-[9px] font-semibold text-[#6D4FA3] transition hover:bg-[#F7F3FD] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            >
              <FileText className="h-3 w-3" />
              Read note
            </button>
            <span className="text-[9px] font-medium text-slate-400">
              [Timestamp]
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}

function TeacherComment() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Class teacher&apos;s comment
        </h2>
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-[#F0EBF9] text-[#6D4FA3]">
          <Quote className="h-4 w-4" />
        </span>
      </div>

      <div className="mt-3 rounded-xl bg-[#E8F5F3] p-4">
        <p className="text-[11px] italic leading-relaxed text-slate-600">
          “[Comment]”
        </p>
      </div>

      <div className="mt-4 flex items-center gap-2.5 border-t border-slate-100 pt-3">
        <span className="grid h-9 w-9 place-items-center rounded-full bg-slate-900 text-[9px] font-bold text-white">
          [AB]
        </span>
        <div className="min-w-0 flex-1">
          <p className="truncate text-[11px] font-bold text-slate-800">
            [Teacher Name]
          </p>
          <p className="text-[9px] text-slate-400">[Role]</p>
        </div>
        <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2 py-0.5 text-[8px] font-semibold text-brand-700 ring-1 ring-brand-100">
          <ShieldCheck className="h-3 w-3" />
          Verified
        </span>
      </div>
    </section>
  );
}

function LevelsCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          How levels work
        </h2>
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <GraduationCap className="h-4 w-4" />
        </span>
      </div>

      <div className="mt-3 grid grid-cols-2 gap-2">
        {LEVEL_CARDS.map((level) => (
          <div key={level.id} className={cn("rounded-xl p-3", level.className)}>
            <p className="text-[8px] font-semibold uppercase tracking-wide opacity-60">
              {level.level}
            </p>
            <p className="mt-1 text-[11px] font-bold">{level.title}</p>
            <p className="mt-1 text-[9px] leading-relaxed opacity-70">
              {level.description}
            </p>
          </div>
        ))}
      </div>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <Info className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          [Evaluation philosophy note] Levels describe achievement against
          learning outcomes, not comparison with other learners.
        </p>
      </div>
    </section>
  );
}

function BottomNav() {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-4 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
        {BOTTOM_NAV.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "relative flex flex-col items-center gap-1 rounded-xl py-2 text-[9px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                item.active
                  ? "text-brand-600"
                  : "text-slate-400 hover:text-slate-600",
              )}
            >
              {item.active && (
                <span className="absolute -top-1 h-0.5 w-8 rounded-full bg-brand-500" />
              )}
              <Icon className="h-5 w-5" />
              <span>{item.label}</span>
            </a>
          );
        })}
      </div>
    </nav>
  );
}
