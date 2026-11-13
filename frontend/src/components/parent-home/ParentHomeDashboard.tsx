"use client";

import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useTransition } from "react";
import {
  AlertTriangle,
  ArrowRight,
  Award,
  BarChart3,
  Bell,
  Bot,
  Bus,
  CalendarCheck,
  CalendarDays,
  Check,
  ChevronRight,
  Clock3,
  CreditCard,
  FileText,
  GraduationCap,
  Home,
  Library,
  LogOut,
  MessageCircle,
  ShieldCheck,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import {
  LEVEL_TONES,
  formatCurrency,
  getFeeProgress,
  type HomeEvent,
  type ParentHomeData,
  type ResultRow,
} from "./parent-home-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href?: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "/parent", active: true },
  { label: "Results", icon: Award, href: "/parent/results" },
  { label: "Attendance", icon: CalendarCheck, href: "/parent/attendance" },
  { label: "Fees", icon: Wallet },
  { label: "AI assistant", icon: Bot },
];

const BOTTOM_NAV: NavItem[] = [
  { label: "Home", icon: Home, href: "/parent", active: true },
  { label: "Results", icon: GraduationCap, href: "/parent/results" },
  { label: "Attendance", icon: CalendarCheck, href: "/parent/attendance" },
  { label: "Fees", icon: Wallet },
  { label: "Assistant", icon: Bot },
];

const QUICK_ACCESS = [
  { id: "timetable", label: "Timetable", icon: CalendarDays },
  { id: "messages", label: "Messages", icon: MessageCircle },
  { id: "library", label: "Library", icon: Library },
  { id: "transport", label: "Transport", icon: Bus },
];

/**
 * Renders the parent home dashboard for the selected linked student.
 *
 * Student selection is driven through the URL so the server reloads that
 * student's fee, attendance, lesson, results and event records together.
 */
export default function ParentHomeDashboard({
  data,
}: {
  data: ParentHomeData | null;
}) {
  const students = data?.students ?? [];
  const session = data?.session ?? null;
  const guardianName = session?.fullName?.trim() || "Guardian";
  const relationship = session?.relationship?.trim() || "Parent";
  const guardianInitials =
    session?.initials?.trim() || initialsFromName(guardianName);
  const portalName = data?.portal?.name?.trim() || "Parent portal";
  const fee = data?.fee ?? null;
  const results = data?.results ?? null;
  const events = data?.events ?? [];
  const selectedStudentId = data?.selectedStudentId ?? students[0]?.id ?? "";

  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();

  const selectStudent = (studentId: string) => {
    if (studentId === selectedStudentId) {
      return;
    }

    const params = new URLSearchParams(searchParams.toString());
    params.set("studentId", studentId);

    startTransition(() => {
      router.push(`${pathname}?${params.toString()}`, { scroll: false });
    });
  };

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#parent-home-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to home
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          guardianName={guardianName}
          relationship={relationship}
          guardianInitials={guardianInitials}
        />
      </aside>

      <MobileAppBar
        portalName={portalName}
        guardianInitials={guardianInitials}
      />

      <div className="lg:pl-[260px]">
        <DesktopHeader
          portalName={portalName}
          guardianName={guardianName}
          relationship={relationship}
          guardianInitials={guardianInitials}
        />

        <main
          id="parent-home-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <StudentFilter
            students={students}
            selectedStudentId={selectedStudentId}
            onSelect={selectStudent}
            isPending={isPending}
          />

          <div
            aria-busy={isPending}
            className={cn(
              "mt-4 grid grid-cols-1 gap-6 transition-opacity xl:grid-cols-[minmax(0,1.45fr)_minmax(360px,1fr)]",
              isPending && "opacity-60",
            )}
          >
            <div className="min-w-0 space-y-4">
              <FeeHero fee={fee} />
              <QuickStats
                attendance={data?.attendance ?? null}
                nextLesson={data?.nextLesson ?? null}
              />
              <ResultsSummary results={results} />
            </div>

            <div className="min-w-0 space-y-4">
              <EventsCard events={events} termLabel={results?.termLabel ?? null} />
              <QuickAccessGrid />
            </div>
          </div>
        </main>
      </div>

      <BottomNav />
    </div>
  );
}

/** Renders the persistent desktop portal navigation and guardian card. */
function SidebarContent({
  guardianName,
  relationship,
  guardianInitials,
}: {
  guardianName: string;
  relationship: string;
  guardianInitials: string;
}) {
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
          const className = cn(
            "group flex min-h-11 items-center gap-3 rounded-xl px-3 py-2.5 text-[12px] font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
            item.active
              ? "bg-[#E8F5F3] font-semibold text-brand-700 ring-1 ring-brand-100"
              : "text-slate-500 hover:bg-slate-50 hover:text-slate-800",
            !item.href && "cursor-default opacity-60",
          );
          const content = (
            <>
              <Icon
                className={cn(
                  "h-[17px] w-[17px] shrink-0",
                  item.active
                    ? "text-brand-600"
                    : "text-slate-400 group-hover:text-brand-600",
                )}
              />
              <span className="flex-1">{item.label}</span>
              {!item.href ? (
                <span className="text-[8px] font-semibold uppercase tracking-wide text-slate-400">
                  Soon
                </span>
              ) : (
                item.active && (
                  <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
                )
              )}
            </>
          );

          if (!item.href) {
            return (
              <span
                key={item.label}
                aria-disabled="true"
                title={`${item.label} is coming soon.`}
                className={className}
              >
                {content}
              </span>
            );
          }

          return (
            <Link
              key={item.label}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={className}
            >
              {content}
            </Link>
          );
        })}
      </nav>

      <div className="mt-auto pt-6">
        <div className="rounded-2xl border border-slate-200/80 bg-slate-50/70 p-3">
          <div className="flex items-center gap-2.5">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
              {guardianInitials}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                {guardianName}
              </p>
              <p className="truncate text-[10px] text-slate-400">
                {relationship}
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                Parent account
              </p>
            </div>
          </div>
          <button
            type="button"
            disabled
            aria-disabled="true"
            title="Sign out will be available once authentication is connected."
            className="mt-3 inline-flex h-10 w-full items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-[11px] font-semibold text-slate-600 transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <LogOut className="h-3.5 w-3.5" />
            Sign out
          </button>
        </div>
      </div>
    </div>
  );
}

/** Renders the compact mobile app bar. */
function MobileAppBar({
  portalName,
  guardianInitials,
}: {
  portalName: string;
  guardianInitials: string;
}) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <div className="min-w-0 flex-1">
          <p className="truncate text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            {portalName}
          </p>
          <p className="truncate text-lg font-bold tracking-tight text-slate-900">
            Home
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
            {guardianInitials}
          </button>
        </div>
      </div>
    </header>
  );
}

/** Renders the sticky desktop header with breadcrumb and guardian profile. */
function DesktopHeader({
  portalName,
  guardianName,
  relationship,
  guardianInitials,
}: {
  portalName: string;
  guardianName: string;
  relationship: string;
  guardianInitials: string;
}) {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="max-w-[220px] truncate text-slate-700">
            {portalName}
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Home</span>
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
              {guardianInitials}
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                {guardianName}
              </p>
              <p className="text-[10px] text-slate-400">{relationship}</p>
            </div>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

/** Renders the linked-student tabs that scope the dashboard to one child. */
function StudentFilter({
  students,
  selectedStudentId,
  onSelect,
  isPending,
}: {
  students: ParentHomeData["students"];
  selectedStudentId: string;
  onSelect: (id: string) => void;
  isPending: boolean;
}) {
  return (
    <div
      className="-mx-4 overflow-x-auto px-4 pb-1 sm:mx-0 sm:px-0"
      role="tablist"
      aria-label="Linked students"
      aria-busy={isPending}
    >
      {students.length > 0 ? (
        <div className="flex min-w-max gap-2">
          {students.map((student) => {
            const active = student.id === selectedStudentId;

            return (
              <button
                key={student.id}
                type="button"
                role="tab"
                aria-selected={active}
                onClick={() => onSelect(student.id)}
                className={cn(
                  "inline-flex items-center gap-2 rounded-full border px-3 py-2 text-[11px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                  active
                    ? "border-brand-600 bg-[#E8F5F3] text-brand-700 shadow-sm"
                    : "border-slate-200/80 bg-white text-slate-500 hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700",
                )}
              >
                <span
                  className={cn(
                    "grid h-6 w-6 place-items-center rounded-full text-[8px] font-bold",
                    active
                      ? "bg-brand-500 text-white"
                      : "bg-slate-100 text-slate-500",
                  )}
                >
                  {student.initials}
                </span>
                {student.name}
              </button>
            );
          })}
        </div>
      ) : (
        <span className="inline-flex items-center rounded-full border border-dashed border-slate-300 bg-white px-3 py-2 text-[11px] font-medium text-slate-500">
          No student records available
        </span>
      )}
    </div>
  );
}

/** Renders the term fee balance, payment progress and fee actions. */
function FeeHero({ fee }: { fee: ParentHomeData["fee"] }) {
  const progress = fee
    ? getFeeProgress(fee.paidAmount, fee.totalAmount)
    : null;
  const feeMeta =
    [
      fee?.dueDate ? `Due ${fee.dueDate}` : null,
      fee?.visibilityNote ?? null,
    ]
      .filter(Boolean)
      .join(" · ") || "No due date or visibility information";

  return (
    <section className="rounded-2xl bg-gradient-to-br from-[#005E5D] to-[#007A78] p-5 text-white shadow-lg shadow-brand-900/20 sm:p-6">
      <div className="flex items-center justify-between gap-3">
        <p className="text-xs font-semibold text-white/90">
          {fee ? `${fee.termLabel} fee balance` : "Fee balance"}
        </p>
        <span
          className={cn(
            "inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[9px] font-semibold ring-1",
            fee
              ? "bg-white/15 text-white ring-white/20"
              : "bg-white/10 text-white/70 ring-white/15",
          )}
        >
          {fee && <Check className="h-3 w-3" />}
          {fee ? fee.status : "No fee data"}
        </span>
      </div>

      <p className="mt-3 text-3xl font-bold leading-none tracking-tight sm:text-4xl">
        {formatCurrency(fee?.amountDue ?? null, fee?.currency ?? null)}
      </p>

      <p className="mt-3 flex items-center gap-1.5 text-[10px] text-white/70">
        <Clock3 className="h-3.5 w-3.5" />
        {feeMeta}
      </p>

      <div className="mt-5">
        <div className="h-2 overflow-hidden rounded-full bg-white/20">
          <div
            className="h-full rounded-full bg-white transition-all duration-500"
            style={{ width: `${progress ?? 0}%` }}
          />
        </div>
        <div className="mt-2 flex items-center justify-between gap-3 text-[9px] font-medium text-white/70">
          <span>
            {fee?.paidAmount !== null && fee?.paidAmount !== undefined
              ? `${formatCurrency(fee.paidAmount, fee.currency)} paid`
              : "No payment data"}
            {fee?.totalAmount !== null && fee?.totalAmount !== undefined
              ? ` of ${formatCurrency(fee.totalAmount, fee.currency)}`
              : ""}
          </span>
          <span>{progress === null ? "—" : `${progress}%`}</span>
        </div>
      </div>

      <div className="mt-5 grid grid-cols-2 gap-2.5">
        <button
          type="button"
          disabled={!fee}
          className="inline-flex h-10 items-center justify-center gap-2 rounded-xl bg-white text-[10px] font-bold text-brand-800 shadow-sm transition hover:bg-brand-50 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50"
        >
          <CreditCard className="h-3.5 w-3.5" />
          Pay with M-Pesa
        </button>
        <button
          type="button"
          disabled={!fee}
          className="inline-flex h-10 items-center justify-center gap-2 rounded-xl border border-white/40 text-[10px] font-bold text-white transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50"
        >
          <FileText className="h-3.5 w-3.5" />
          Statement
        </button>
      </div>
    </section>
  );
}

/** Renders the attendance and next-lesson summary cards. */
function QuickStats({
  attendance,
  nextLesson,
}: {
  attendance: ParentHomeData["attendance"];
  nextLesson: ParentHomeData["nextLesson"];
}) {
  return (
    <div className="grid grid-cols-2 gap-3">
      <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card">
        <div className="flex items-center justify-between gap-2">
          <p className="text-[10px] font-semibold text-slate-500">Attendance</p>
          <CalendarCheck className="h-3.5 w-3.5 text-slate-400" />
        </div>
        <span
          className={cn(
            "mt-3 inline-flex items-center gap-1 rounded-full px-2 py-1 text-[9px] font-semibold ring-1",
            attendance
              ? attendance.present
                ? "bg-[#F0EBF9] text-[#6D4FA3] ring-[#E4DBF5]"
                : "bg-[#FEF3D6] text-[#8A6A25] ring-[#EBD7A5]"
              : "bg-slate-100 text-slate-500 ring-slate-200",
          )}
        >
          {attendance ? (
            <>
              {attendance.present ? (
                <Check className="h-3 w-3" />
              ) : (
                <AlertTriangle className="h-3 w-3" />
              )}
              {attendance.status}
            </>
          ) : (
            "No attendance data"
          )}
        </span>
      </section>

      <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card">
        <div className="flex items-center justify-between gap-2">
          <p className="text-[10px] font-semibold text-slate-500">
            Next lesson
          </p>
          <Clock3 className="h-3.5 w-3.5 text-slate-400" />
        </div>
        {nextLesson ? (
          <>
            <p className="mt-2 truncate text-xs font-bold text-slate-800">
              {nextLesson.subject}
            </p>
            <p className="mt-0.5 truncate text-[9px] text-slate-400">
              {[nextLesson.time, nextLesson.room].filter(Boolean).join(" · ") ||
                "Time and room not provided"}
            </p>
          </>
        ) : (
          <>
            <p className="mt-2 text-[11px] font-semibold text-slate-600">
              No upcoming lesson
            </p>
            <p className="mt-0.5 text-[9px] leading-relaxed text-slate-400">
              Time and room will appear here.
            </p>
          </>
        )}
      </section>
    </div>
  );
}

/** Renders the released results summary, level key and correction notice. */
function ResultsSummary({
  results,
}: {
  results: ParentHomeData["results"];
}) {
  const rows = results?.rows ?? [];
  const levels = results?.levels ?? [];
  const correction = results?.correction ?? null;

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            {results?.termLabel ? `${results.termLabel} results` : "Results"}
          </h2>
          <p className="mt-1 text-[10px] text-slate-400">
            {results?.framework?.trim() || "Framework not available"}
          </p>
        </div>
        <span className="shrink-0 rounded-full bg-slate-100 px-2.5 py-1 text-[9px] font-semibold text-slate-500">
          {results?.releaseDate?.trim() || "No release date"}
        </span>
      </div>

      {rows.length > 0 ? (
        <div className="mt-3 space-y-3">
          {rows.map((row) => (
            <ResultSubjectRow key={row.id} row={row} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={BarChart3}
          title="No results available"
          body="Released subject performance levels will appear here."
          className="mt-3"
        />
      )}

      <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1.5">
        {levels.length > 0 ? (
          levels.map((level) => (
            <span
              key={level.code}
              className="inline-flex items-center gap-1 text-[8px] font-medium text-slate-500"
            >
              <span
                className={cn(
                  "h-2 w-2 rounded-full",
                  LEVEL_TONES[level.tone],
                )}
              />
              {level.code} · {level.label}
            </span>
          ))
        ) : (
          <span className="text-[8px] text-slate-400">
            Level key unavailable
          </span>
        )}
      </div>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#F0EBF9] p-3">
        <ShieldCheck className="h-3.5 w-3.5 shrink-0 text-[#6D4FA3]" />
        <div className="min-w-0 flex-1">
          <p className="text-[10px] font-semibold text-slate-700">
            {correction ? "Correction recorded" : "No corrections recorded"}
          </p>
          <p className="mt-0.5 text-[9px] leading-relaxed text-slate-500">
            {correction?.note ||
              "Correction notices will appear here when a result is adjusted."}
          </p>
        </div>
        {correction?.hasNote && (
          <button
            type="button"
            className="h-7 shrink-0 rounded-lg border border-[#D7C8F2] bg-white px-2 text-[8px] font-semibold text-[#6D4FA3]"
          >
            Read note
          </button>
        )}
      </div>

      <div className="mt-4 border-t border-slate-100 pt-3 text-center">
        <p className="text-[9px] leading-relaxed text-slate-400">
          {results?.disclaimer?.trim() ||
            "Results describe achievement against learning outcomes, not learner ranking."}
        </p>
        <a
          href="/parent/results"
          className="mt-1.5 inline-flex items-center gap-1 text-[10px] font-semibold text-brand-700 transition hover:text-brand-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          See all results
          <ArrowRight className="h-3.5 w-3.5" />
        </a>
      </div>
    </section>
  );
}

/** Renders one subject's performance level and mastery bar. */
function ResultSubjectRow({ row }: { row: ResultRow }) {
  const width = row.masteryPercent === null ? 0 : Math.min(100, Math.max(0, row.masteryPercent));

  return (
    <div>
      <div className="flex items-center justify-between gap-3">
        <p className="truncate text-[11px] font-semibold text-slate-700">
          {row.subject}
        </p>
        <div className="flex shrink-0 items-center gap-1.5">
          <span className="text-[9px] font-medium text-slate-400">
            {row.performanceLevel}
          </span>
          <span className="rounded-full bg-slate-100 px-1.5 py-0.5 text-[8px] font-bold text-slate-600">
            {row.levelCode}
          </span>
        </div>
      </div>
      <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-100">
        <div
          className="h-full rounded-full bg-brand-500"
          style={{ width: `${width}%` }}
        />
      </div>
    </div>
  );
}

/** Renders upcoming calendar events for the selected student's term. */
function EventsCard({
  events,
  termLabel,
}: {
  events: HomeEvent[];
  termLabel: string | null;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Coming up
        </h2>
        <span
          className="text-[9px] font-semibold text-slate-400"
          title="The calendar view is coming soon."
        >
          {termLabel ? `${termLabel} Calendar` : "Calendar"}
        </span>
      </div>

      {events.length > 0 ? (
        <div className="mt-3 divide-y divide-slate-100">
          {events.map((event) => (
            <div key={event.id} className="flex items-center gap-3 py-3 first:pt-0 last:pb-0">
              <span className="grid h-11 w-10 shrink-0 place-items-center rounded-xl bg-[#F0EBF9] text-[#6D4FA3]">
                <span className="text-[7px] font-semibold uppercase tracking-wide">
                  {event.month}
                </span>
                <span className="text-sm font-bold leading-none">{event.day}</span>
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-[11px] font-bold text-slate-800">
                  {event.title}
                </p>
                <p className="mt-0.5 truncate text-[9px] text-slate-400">
                  {event.meta || "Time and location not provided"}
                </p>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          icon={CalendarDays}
          title="No upcoming events"
          body="School events will appear here when they are added to the calendar."
          className="mt-3"
        />
      )}
    </section>
  );
}

/** Renders structural shortcuts that are not yet connected. */
function QuickAccessGrid() {
  return (
    <section>
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Quick access
        </h2>
        <span className="text-[9px] font-medium text-slate-400">
          Coming soon
        </span>
      </div>

      <div className="mt-3 grid grid-cols-2 gap-2.5">
        {QUICK_ACCESS.map((item) => {
          const Icon = item.icon;

          return (
            <div
              key={item.id}
              className="rounded-2xl border border-dashed border-slate-300 bg-white/70 p-4 text-center"
            >
              <span className="mx-auto grid h-9 w-9 place-items-center rounded-xl bg-slate-50 text-slate-400">
                <Icon className="h-4 w-4" />
              </span>
              <p className="mt-2 text-[10px] font-bold text-slate-700">
                {item.label}
              </p>
              <p className="mt-0.5 text-[8px] font-medium text-slate-400">
                Coming soon
              </p>
            </div>
          );
        })}
      </div>
    </section>
  );
}

/** Renders a consistent empty state for a dashboard section. */
function EmptyState({
  icon: Icon,
  title,
  body,
  className,
}: {
  icon: LucideIcon;
  title: string;
  body: string;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "rounded-xl border border-dashed border-slate-200 bg-slate-50/60 px-4 py-7 text-center",
        className,
      )}
    >
      <span className="mx-auto grid h-9 w-9 place-items-center rounded-xl bg-white text-slate-400 shadow-sm">
        <Icon className="h-4 w-4" />
      </span>
      <p className="mt-2.5 text-[11px] font-semibold text-slate-700">
        {title}
      </p>
      <p className="mx-auto mt-1 max-w-xs text-[9px] leading-relaxed text-slate-500">
        {body}
      </p>
    </div>
  );
}

/** Renders the fixed mobile navigation with real and coming-soon entries. */
function BottomNav() {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-5 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
        {BOTTOM_NAV.map((item) => {
          const Icon = item.icon;
          const className = cn(
            "relative flex min-h-11 flex-col items-center justify-center gap-1 rounded-xl py-2 text-[9px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
            item.active
              ? "text-brand-600"
              : "text-slate-400 hover:text-slate-600",
            !item.href && "cursor-default opacity-60",
          );
          const content = (
            <>
              {item.active && (
                <span className="absolute -top-1 h-0.5 w-8 rounded-full bg-brand-500" />
              )}
              <Icon className="h-5 w-5" />
              <span>{item.label}</span>
            </>
          );

          if (!item.href) {
            return (
              <span
                key={item.label}
                aria-disabled="true"
                title={`${item.label} is coming soon.`}
                className={className}
              >
                {content}
              </span>
            );
          }

          return (
            <Link
              key={item.label}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={className}
            >
              {content}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}

/** Builds up to two uppercase initials from a guardian's display name. */
function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "G";
}
