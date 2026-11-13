"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import {
  AlertTriangle,
  ArrowLeft,
  Award,
  CalendarCheck,
  CalendarDays,
  Check,
  ChevronLeft,
  ChevronRight,
  Clock3,
  GraduationCap,
  Home,
  Info,
  LogOut,
  ShieldCheck,
  Sparkles,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import {
  buildCalendarMonth,
  formatDayCount,
  formatMonthLabel,
  formatRecordedDays,
  formatTermAttendance,
  type AbsenceRecord,
  type AttendanceStatus,
  type StudentAttendanceData,
} from "./student-attendance-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href?: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "/parent" },
  { label: "Results", icon: Award },
  {
    label: "Attendance",
    icon: CalendarCheck,
    href: "/parent/attendance",
    active: true,
  },
  { label: "Fees", icon: Wallet },
  { label: "AI assistant", icon: Sparkles },
];

const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

const STATUS_STYLES: Record<AttendanceStatus, string> = {
  present: "bg-[#F0EBF9] text-[#6D4FA3]",
  absent: "bg-[#FEF3D6] text-[#8A6A25]",
  "not-marked": "bg-transparent text-slate-500",
  "no-school": "bg-[#F1F5F4] text-slate-400",
};

export default function StudentAttendanceDashboard({
  data,
  initialMonth,
}: {
  data: StudentAttendanceData | null;
  initialMonth: { year: number; month: number };
}) {
  const students = data?.students ?? [];
  const session = data?.session ?? null;
  const guardianName = session?.fullName?.trim() || "Guardian";
  const relationship = session?.relationship?.trim() || "Parent";
  const guardianInitials =
    session?.initials?.trim() || initialsFromName(guardianName);
  const portalName = data?.portal?.name?.trim() || "Parent portal";
  const termName = data?.termLabel?.trim() || null;
  const summary = data?.summary ?? {
    presentDays: null,
    absentDays: null,
    recordedDays: null,
    liveSync: false,
  };
  const absenceRecords = data?.absenceRecords ?? [];
  const complianceNote =
    data?.complianceNote?.trim() ||
    "Attendance records are retained and accessed under school policy and data-protection rules.";

  const [selectedStudentId, setSelectedStudentId] = useState(
    data?.selectedStudentId ?? students[0]?.id ?? "",
  );
  const [viewMonth, setViewMonth] = useState(initialMonth);

  const cells = useMemo(
    () =>
      buildCalendarMonth(
        viewMonth.year,
        viewMonth.month,
        data?.statusByDate ?? {},
      ),
    [viewMonth, data?.statusByDate],
  );

  function changeMonth(offset: number) {
    setViewMonth((current) => {
      const next = new Date(current.year, current.month + offset, 1);
      return { year: next.getFullYear(), month: next.getMonth() };
    });
  }

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#attendance-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to attendance
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          guardianName={guardianName}
          relationship={relationship}
          guardianInitials={guardianInitials}
        />
      </aside>

      <MobileAppBar guardianInitials={guardianInitials} />

      <div className="lg:pl-[260px]">
        <DesktopHeader
          portalName={portalName}
          guardianName={guardianName}
          relationship={relationship}
          guardianInitials={guardianInitials}
        />

        <main
          id="attendance-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
            <h1 className="min-w-0 text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">
              {termName ? `${termName} Roll-Call` : "Roll-Call"}
            </h1>
            <span
              className={cn(
                "inline-flex w-fit items-center gap-1.5 rounded-full px-2.5 py-1 text-[10px] font-semibold ring-1",
                data?.termAttendancePercent !== null &&
                  data?.termAttendancePercent !== undefined
                  ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
                  : "bg-slate-100 text-slate-500 ring-slate-200",
              )}
            >
              <CalendarCheck className="h-3.5 w-3.5" />
              {formatTermAttendance(data?.termAttendancePercent ?? null)}
            </span>
          </div>

          <StudentFilter
            students={students}
            selectedStudentId={selectedStudentId}
            onSelect={setSelectedStudentId}
          />

          <div className="mt-4 grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(360px,1fr)]">
            <div className="min-w-0 space-y-4">
              <SummaryCard summary={summary} />
              <CalendarCard
                cells={cells}
                monthLabel={formatMonthLabel(viewMonth.year, viewMonth.month)}
                onPrevious={() => changeMonth(-1)}
                onNext={() => changeMonth(1)}
              />
            </div>

            <div className="min-w-0 space-y-4 xl:sticky xl:top-[84px] xl:self-start">
              <AbsenceRecordsCard records={absenceRecords} />
              <ComplianceCard note={complianceNote} />
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

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
            Bridge <span className="text-brand-600">Portal</span>
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
            "group flex min-h-11 items-center gap-3 rounded-xl px-3 py-2.5 text-[12px] font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
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
              {item.active && (
                <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
              )}
            </>
          );

          if (!item.href) {
            return (
              <span
                key={item.label}
                aria-disabled="true"
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
            className="mt-3 inline-flex h-10 w-full items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-[11px] font-semibold text-slate-600 transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
          >
            <LogOut className="h-3.5 w-3.5" />
            Sign out
          </button>
        </div>
      </div>
    </div>
  );
}

function MobileAppBar({ guardianInitials }: { guardianInitials: string }) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <div className="flex min-w-0 items-center gap-2">
          <Link
            href="/parent"
            className="grid h-11 w-11 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="Back to student profile"
          >
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <p className="truncate text-lg font-bold tracking-tight text-slate-900">
            Student Profile
          </p>
        </div>

        <button
          type="button"
          disabled
          aria-disabled="true"
          title="Profile menu is unavailable until authentication is connected."
          className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
          aria-label="User profile unavailable"
        >
          {guardianInitials}
        </button>
      </div>
    </header>
  );
}

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
          <span>Student Profile</span>
        </div>

        <button
          type="button"
          disabled
          aria-disabled="true"
          aria-expanded={false}
          aria-label="Open guardian profile unavailable"
          title="Profile menu is unavailable until authentication is connected."
          className="flex min-h-11 items-center gap-2.5 rounded-xl border border-slate-200/80 bg-white py-1.5 pl-1.5 pr-3 shadow-sm transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
        >
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-[10px] font-bold text-white">
            {guardianInitials}
          </span>
          <div className="min-w-0 text-left">
            <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
              {guardianName}
            </p>
            <p className="text-[10px] text-slate-600">{relationship}</p>
          </div>
          <ChevronRight className="h-3.5 w-3.5 text-slate-500" />
        </button>
      </div>
    </header>
  );
}

function StudentFilter({
  students,
  selectedStudentId,
  onSelect,
}: {
  students: StudentAttendanceData["students"];
  selectedStudentId: string;
  onSelect: (id: string) => void;
}) {
  return (
    <div
      className="mt-4 -mx-4 overflow-x-auto px-4 pb-1 sm:mx-0 sm:px-0"
      role="tablist"
      aria-label="Student records"
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
                  "inline-flex min-h-11 items-center gap-2 rounded-full border px-3 py-2 text-[11px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
                  active
                    ? "border-brand-600 bg-[#E8F5F3] text-brand-700 shadow-sm"
                    : "border-slate-200/80 bg-white text-slate-500 hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700",
                )}
              >
                <span
                  className={cn(
                    "h-1.5 w-1.5 rounded-full",
                    active ? "bg-brand-600" : "bg-slate-300",
                  )}
                />
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

function SummaryCard({
  summary,
}: {
  summary: StudentAttendanceData["summary"];
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card sm:p-4">
      <div className="grid grid-cols-2 gap-2.5">
        <div className="rounded-xl bg-[#F0EBF9] p-3.5">
          <div className="flex items-center gap-1.5 text-[#6D4FA3]">
            <span className="grid h-6 w-6 place-items-center rounded-full bg-white/80">
              <Check className="h-3.5 w-3.5" />
            </span>
            <span className="text-[9px] font-semibold uppercase tracking-wide">
              Present
            </span>
          </div>
          <p className="mt-2 text-lg font-bold tracking-tight text-[#4F3585]">
            {formatDayCount(summary.presentDays)}
          </p>
        </div>

        <div className="rounded-xl bg-[#FEF3D6] p-3.5">
          <div className="flex items-center gap-1.5 text-[#8A6A25]">
            <span className="grid h-6 w-6 place-items-center rounded-full bg-white/80">
              <AlertTriangle className="h-3.5 w-3.5" />
            </span>
            <span className="text-[9px] font-semibold uppercase tracking-wide">
              Absent
            </span>
          </div>
          <p className="mt-2 text-lg font-bold tracking-tight text-[#6B521C]">
            {formatDayCount(summary.absentDays)}
          </p>
        </div>
      </div>

      <div className="mt-3 flex items-center justify-between gap-3 border-t border-slate-100 pt-3">
        <span className="inline-flex items-center gap-1.5 text-[9px] font-medium text-slate-500">
          <CalendarDays className="h-3.5 w-3.5 text-slate-400" />
          {formatRecordedDays(summary.recordedDays)}
        </span>
        <span
          className={cn(
            "inline-flex items-center gap-1.5 text-[9px] font-semibold",
            summary.liveSync ? "text-brand-700" : "text-slate-400",
          )}
        >
          <span
            className={cn(
              "h-1.5 w-1.5 rounded-full",
              summary.liveSync ? "bg-brand-500" : "bg-slate-300",
            )}
          />
          {summary.liveSync ? "Live sync" : "Not synced"}
        </span>
      </div>
    </section>
  );
}

function CalendarCard({
  cells,
  monthLabel,
  onPrevious,
  onNext,
}: {
  cells: ReturnType<typeof buildCalendarMonth>;
  monthLabel: string;
  onPrevious: () => void;
  onNext: () => void;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card sm:p-4">
      <div className="flex items-center justify-between gap-3 px-1 pb-3">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
            <CalendarDays className="h-4 w-4" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            {monthLabel}
          </h2>
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={onPrevious}
            className="grid h-11 w-11 place-items-center rounded-full border border-slate-200/80 bg-white text-slate-500 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="Previous month"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            type="button"
            onClick={onNext}
            className="grid h-11 w-11 place-items-center rounded-full border border-slate-200/80 bg-white text-slate-500 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="Next month"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-7 gap-1">
        {WEEKDAYS.map((day) => (
          <span
            key={day}
            className="py-1 text-center text-[8px] font-semibold uppercase tracking-wide text-slate-400"
          >
            {day}
          </span>
        ))}

        {cells.map((cell) => (
          <CalendarDay key={cell.key} cell={cell} />
        ))}
      </div>

      <div className="mt-3 rounded-xl bg-[#E2F0ED] px-3 py-2.5">
        <div className="grid grid-cols-2 gap-x-4 gap-y-1.5 text-[9px] font-medium">
          <span className="inline-flex items-center gap-1.5 text-[#6D4FA3]">
            <Check className="h-3 w-3" />
            Present
          </span>
          <span className="inline-flex items-center gap-1.5 text-[#8A6A25]">
            <AlertTriangle className="h-3 w-3" />
            Absent
          </span>
          <span className="inline-flex items-center gap-1.5 text-slate-500">
            <span className="h-1 w-1 rounded-full bg-slate-400" />
            Not marked
          </span>
          <span className="inline-flex items-center gap-1.5 text-slate-400">
            <span className="h-px w-2 bg-slate-400" />
            No school
          </span>
        </div>
      </div>
    </section>
  );
}

function CalendarDay({
  cell,
}: {
  cell: ReturnType<typeof buildCalendarMonth>[number];
}) {
  const status = cell.status;

  if (!cell.currentMonth || status === null) {
    return (
      <div className="flex min-h-9 flex-col items-center justify-center rounded-lg text-[10px] font-medium text-slate-500">
        {cell.day}
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex min-h-9 flex-col items-center justify-center rounded-lg",
        status === "present" && "bg-[#F0EBF9] text-[#6D4FA3]",
        status === "absent" && "bg-[#FEF3D6] text-[#8A6A25]",
        status === "not-marked" && "bg-transparent text-slate-500",
        status === "no-school" && "bg-[#F1F5F4] text-slate-400",
      )}
    >
      <span className="text-[10px] font-semibold">{cell.day}</span>
      {status === "present" && <Check className="mt-0.5 h-2.5 w-2.5" />}
      {status === "absent" && (
        <AlertTriangle className="mt-0.5 h-2.5 w-2.5" />
      )}
      {status === "not-marked" && (
        <span className="mt-0.5 h-1 w-1 rounded-full bg-slate-400" />
      )}
      {status === "no-school" && (
        <span className="mt-0.5 h-px w-2 bg-slate-400" />
      )}
    </div>
  );
}

function AbsenceRecordsCard({ records }: { records: AbsenceRecord[] }) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-slate-50 text-slate-500">
            <Clock3 className="h-4 w-4" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Absence records this term
          </h2>
        </div>
        <span className="rounded-full bg-[#FEF3D6] px-2.5 py-1 text-[9px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
          {records.length === 1 ? "1 Record" : `${records.length} Records`}
        </span>
      </div>

      {records.length > 0 ? (
        <div className="mt-3 space-y-2">
          {records.map((record) => (
            <div key={record.id} className="rounded-xl bg-[#F8FAF9] p-3 ring-1 ring-slate-100">
              <div className="flex items-start justify-between gap-2">
                <p className="text-[11px] font-bold text-slate-800">
                  {record.date}
                </p>
                <span className="rounded-full bg-[#FEF3D6] px-2 py-0.5 text-[8px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
                  {record.status}
                </span>
              </div>
              <p className="mt-1 text-[9px] text-slate-500">
                Marked absent by {record.markedBy}
              </p>
              <p className="mt-1.5 inline-flex items-center gap-1 text-[9px] text-slate-400">
                <Clock3 className="h-3 w-3" />
                {record.timestamp}
              </p>
            </div>
          ))}
        </div>
      ) : (
        <div className="mt-3 rounded-xl border border-dashed border-slate-200 bg-slate-50/60 px-4 py-7 text-center">
          <CalendarDays className="mx-auto h-5 w-5 text-slate-400" />
          <p className="mt-2 text-[11px] font-semibold text-slate-700">
            No absence records for this term
          </p>
          <p className="mx-auto mt-1 max-w-xs text-[9px] leading-relaxed text-slate-500">
            Logged absences will appear here when attendance is recorded.
          </p>
        </div>
      )}
    </section>
  );
}

function ComplianceCard({ note }: { note: string }) {
  return (
    <section className="flex gap-3 rounded-2xl border border-brand-100 bg-[#E2F0ED] p-4">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-brand-700 shadow-sm">
        <ShieldCheck className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-xs font-bold text-brand-800">
          Attendance data protection
        </p>
        <p className="mt-1 text-[10px] leading-relaxed text-slate-500">
          {note}
        </p>
        <p className="mt-1 inline-flex items-center gap-1 text-[9px] font-medium text-brand-700">
          <Info className="h-3 w-3" />
          Access is limited to authorized guardians and staff.
        </p>
      </div>
    </section>
  );
}

function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "G";
}
