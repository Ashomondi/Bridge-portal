"use client";

import { useState } from "react";
import {
  ArrowLeft,
  Award,
  BookOpen,
  CalendarCheck,
  CalendarDays,
  CalendarRange,
  ChevronRight,
  Clock3,
  Coffee,
  GraduationCap,
  Home,
  Info,
  LogOut,
  ShieldCheck,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import type {
  KeyDate,
  ParentTimetableData,
  TimetablePeriod,
  TimetableStudent,
  WeekDay,
} from "./timetable-events-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "/parent" },
  {
    label: "Timetable",
    icon: CalendarDays,
    href: "/parent/timetable",
    active: true,
  },
  { label: "Results", icon: Award, href: "/parent/results" },
  { label: "Attendance", icon: CalendarCheck, href: "/parent/attendance" },
  { label: "Fees", icon: Wallet, href: "#" },
];

export default function ParentTimetableDashboard({
  data,
  week,
}: {
  data: ParentTimetableData | null;
  week: WeekDay[];
}) {
  const students = data?.students ?? [];
  const session = data?.session ?? null;
  const guardianName = session?.fullName?.trim() || "Guardian";
  const relationship = session?.relationship?.trim() || "Parent";
  const guardianInitials =
    session?.initials?.trim() || initialsFromName(guardianName);
  const portalName = data?.portal?.name?.trim() || "Parent portal";
  const periods = data?.periods ?? [];
  const keyDates = data?.keyDates ?? [];
  const notice = data?.notice ?? null;
  const initialDay = week.find((day) => day.isToday) ?? week[0];

  const [selectedStudentId, setSelectedStudentId] = useState(
    data?.selectedStudentId ?? students[0]?.id ?? "",
  );
  const [selectedDay, setSelectedDay] = useState(initialDay?.id ?? "");
  const [view, setView] = useState<"timetable" | "calendar">("timetable");

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#timetable-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to timetable
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
          id="timetable-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            Timetable &amp; Events
          </h1>

          <StudentFilter
            students={students}
            selectedStudentId={selectedStudentId}
            onSelect={setSelectedStudentId}
          />

          <ViewToggle view={view} onChange={setView} />
          <OfficialNotice notice={notice} />
          <WeekSelector
            week={week}
            weekLabel={data?.weekLabel?.trim() || null}
            termLabel={data?.termLabel?.trim() || null}
            selectedDay={selectedDay}
            onSelect={setSelectedDay}
          />

          <div className="mt-4 grid grid-cols-1 gap-5 xl:grid-cols-[minmax(0,1.3fr)_minmax(340px,1fr)]">
            <div className="min-w-0">
              {view === "timetable" ? (
                <PeriodsSection periods={periods} streamLabel={null} />
              ) : (
                <CalendarEmpty />
              )}
            </div>
            <div className="min-w-0 xl:sticky xl:top-[84px] xl:self-start">
              <KeyDatesCard dates={keyDates} />
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

function MobileAppBar({ guardianInitials }: { guardianInitials: string }) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <div className="flex min-w-0 items-center gap-2">
          <a
            href="/parent"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Back to parent home"
          >
            <ArrowLeft className="h-5 w-5" />
          </a>
          <p className="truncate text-lg font-bold tracking-tight text-slate-900">
            Timetable
          </p>
        </div>

        <button
          type="button"
          className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open user profile"
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
          <span>Timetable &amp; Events</span>
        </div>

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
    </header>
  );
}

function StudentFilter({
  students,
  selectedStudentId,
  onSelect,
}: {
  students: TimetableStudent[];
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
                  "inline-flex items-center gap-2 rounded-full border px-3 py-2 text-[11px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
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

function ViewToggle({
  view,
  onChange,
}: {
  view: "timetable" | "calendar";
  onChange: (view: "timetable" | "calendar") => void;
}) {
  return (
    <div className="mt-3 rounded-full bg-[#E8F5F3] p-1">
      <div className="grid grid-cols-2 gap-1" role="tablist" aria-label="View">
        <button
          type="button"
          role="tab"
          aria-selected={view === "timetable"}
          onClick={() => onChange("timetable")}
          className={cn(
            "inline-flex items-center justify-center gap-1.5 rounded-full px-3 py-2 text-[10px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
            view === "timetable"
              ? "bg-white text-brand-700 shadow-sm"
              : "text-slate-400",
          )}
        >
          <CalendarDays className="h-3.5 w-3.5" />
          Timetable
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={view === "calendar"}
          onClick={() => onChange("calendar")}
          className={cn(
            "inline-flex items-center justify-center gap-1.5 rounded-full px-3 py-2 text-[10px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
            view === "calendar"
              ? "bg-white text-brand-700 shadow-sm"
              : "text-slate-400",
          )}
        >
          <CalendarRange className="h-3.5 w-3.5" />
          Calendar
        </button>
      </div>
    </div>
  );
}

function OfficialNotice({
  notice,
}: {
  notice: ParentTimetableData["notice"];
}) {
  return (
    <section className="mt-3 flex gap-3 rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
        <ShieldCheck className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-[8px] font-semibold uppercase tracking-[0.16em] text-[#8A6DBC]">
          {notice?.authority?.trim() || "Academic office"}
        </p>
        <p className="mt-1 text-[11px] leading-relaxed text-slate-600">
          {notice?.note?.trim() || "No academic notice available."}
        </p>
      </div>
    </section>
  );
}

function WeekSelector({
  week,
  weekLabel,
  termLabel,
  selectedDay,
  onSelect,
}: {
  week: WeekDay[];
  weekLabel: string | null;
  termLabel: string | null;
  selectedDay: string;
  onSelect: (id: string) => void;
}) {
  return (
    <section className="mt-3">
      <div className="flex items-center justify-between gap-3 px-1">
        <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
          {weekLabel || "Week schedule"}
        </p>
        <p className="text-[9px] font-medium text-slate-400">
          {termLabel || "No active term"}
        </p>
      </div>

      {week.length > 0 ? (
        <div className="mt-2 rounded-2xl bg-[#E8F5F3] p-2">
          <div className="grid grid-cols-5 gap-1.5">
            {week.map((day) => {
              const active = day.id === selectedDay;

              return (
                <button
                  key={day.id}
                  type="button"
                  onClick={() => onSelect(day.id)}
                  aria-pressed={active}
                  aria-label={`${day.label} ${day.date}${day.isToday ? ", today" : ""}`}
                  className={cn(
                    "relative flex flex-col items-center rounded-xl px-1 py-2 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                    active
                      ? "bg-brand-500 text-white shadow-lg shadow-brand-500/20"
                      : "bg-white/70 text-slate-500 hover:bg-white hover:text-brand-700",
                  )}
                >
                  <span
                    className={cn(
                      "mb-0.5 h-4 rounded-full px-1.5 text-[7px] font-semibold leading-4",
                      day.isToday
                        ? active
                          ? "bg-white/20 text-white"
                          : "bg-brand-50 text-brand-700"
                        : "text-transparent",
                    )}
                  >
                    Today
                  </span>
                  <span
                    className={cn(
                      "text-[8px] font-semibold uppercase tracking-wide",
                      active ? "text-white/70" : "text-slate-400",
                    )}
                  >
                    {day.label}
                  </span>
                  <span className="mt-0.5 text-sm font-bold leading-none">
                    {day.date}
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      ) : (
        <div className="mt-2 rounded-2xl border border-dashed border-slate-300 bg-white p-6 text-center text-[10px] text-slate-500">
          No teaching week available.
        </div>
      )}
    </section>
  );
}

function PeriodsSection({
  periods,
  streamLabel,
}: {
  periods: TimetablePeriod[];
  streamLabel: string | null;
}) {
  return (
    <section>
      <div className="flex items-center justify-between gap-3 px-1">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Today&apos;s Periods
        </h2>
        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
          {streamLabel || "No stream assigned"}
        </span>
      </div>

      {periods.length > 0 ? (
        <div className="mt-3 space-y-2.5">
          {periods.map((period) => (
            <PeriodCard key={period.id} period={period} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={CalendarDays}
          title="No periods scheduled"
          body="The day's timetable will appear here when it is published."
          className="mt-3"
        />
      )}
    </section>
  );
}

function PeriodCard({ period }: { period: TimetablePeriod }) {
  if (period.kind === "break") {
    return (
      <div className="flex items-center justify-between gap-3 rounded-2xl border border-slate-200/80 bg-[#F1F5F4] px-3.5 py-3">
        <span className="inline-flex items-center gap-2 text-[10px] font-semibold text-slate-600">
          <Coffee className="h-4 w-4 text-slate-400" />
          {period.label}
        </span>
        <span className="text-[9px] font-medium text-slate-500">
          {formatTimeRange(period.start, period.end)}
          {period.duration && (
            <span className="ml-2 rounded-full bg-white px-1.5 py-0.5 text-[8px] font-semibold text-slate-400">
              {period.duration}
            </span>
          )}
        </span>
      </div>
    );
  }

  const active = period.kind === "active";
  const Icon = active ? Clock3 : period.kind === "upcoming" ? Clock3 : BookOpen;

  return (
    <article
      className={cn(
        "flex items-center gap-3 rounded-2xl border p-3.5 shadow-card",
        active
          ? "border-[#D7C8F2] bg-[#F7F3FD] shadow-[0_12px_34px_-24px_rgba(109,79,163,0.65)]"
          : "border-slate-200/80 bg-white",
      )}
    >
      <span
        className={cn(
          "grid h-10 w-10 shrink-0 place-items-center rounded-xl",
          active
            ? "bg-[#F0EBF9] text-[#6D4FA3]"
            : "bg-[#E8F5F3] text-brand-600",
        )}
      >
        <Icon className="h-4 w-4" />
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <p className="truncate text-[11px] font-bold text-slate-800">
            {period.subject}
          </p>
          {active && (
            <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-[#7C5CBF] px-1.5 py-0.5 text-[7px] font-bold text-white">
              Now
            </span>
          )}
        </div>
        <p className="mt-0.5 truncate text-[9px] text-slate-400">
          {[period.teacher, period.room].filter(Boolean).join(" · ") ||
            "Teacher and room not provided"}
        </p>
      </div>

      <div className="shrink-0 text-right">
        <p
          className={cn(
            "text-[10px] font-semibold",
            active ? "text-[#6D4FA3]" : "text-slate-600",
          )}
        >
          {formatTimeRange(period.start, period.end)}
        </p>
        <p
          className={cn(
            "mt-0.5 text-[8px] font-medium",
            active ? "text-[#8A6DBC]" : "text-slate-400",
          )}
        >
          {period.label}
        </p>
      </div>
    </article>
  );
}

function CalendarEmpty() {
  return (
    <section>
      <div className="flex items-center justify-between gap-3 px-1">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Calendar view
        </h2>
        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
          No stream assigned
        </span>
      </div>
      <div className="mt-3 rounded-2xl border border-dashed border-slate-300 bg-white p-8 text-center">
        <CalendarRange className="mx-auto h-5 w-5 text-slate-400" />
        <p className="mt-2 text-[11px] font-semibold text-slate-700">
          Calendar view structure
        </p>
        <p className="mx-auto mt-1 max-w-xs text-[9px] leading-relaxed text-slate-500">
          The monthly calendar will appear here when the view is connected.
        </p>
      </div>
    </section>
  );
}

function KeyDatesCard({ dates }: { dates: KeyDate[] }) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Key Term Dates
        </h2>
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <CalendarDays className="h-4 w-4" />
        </span>
      </div>

      {dates.length > 0 ? (
        <div className="mt-3 space-y-2">
          {dates.map((date) => (
            <KeyDateItem key={date.id} date={date} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={CalendarDays}
          title="No key dates available"
          body="Term dates and events will appear here when they are published."
          className="mt-3"
        />
      )}
    </section>
  );
}

function KeyDateItem({ date }: { date: KeyDate }) {
  if (date.kind === "notice") {
    return (
      <article className="flex gap-3 rounded-xl bg-[#E8F5F3] p-3">
        <span className="grid h-11 w-10 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
          <span className="text-[7px] font-semibold uppercase text-slate-400">
            {date.month}
          </span>
          <span className="text-sm font-bold leading-none text-[#6D4FA3]">
            {date.day}
          </span>
        </span>
        <div className="min-w-0">
          <p className="inline-flex items-center gap-1 text-[7px] font-semibold uppercase tracking-[0.12em] text-brand-700">
            <Info className="h-2.5 w-2.5" />
            {date.badge || "Notice"}
          </p>
          <p className="mt-0.5 truncate text-[10px] font-bold text-slate-800">
            {date.title}
          </p>
          <p className="mt-0.5 text-[8px] leading-relaxed text-slate-500">
            {date.meta || "Details not provided"}
          </p>
        </div>
      </article>
    );
  }

  return (
    <article className="flex items-center gap-3 rounded-xl py-2">
      <span className="grid h-11 w-10 shrink-0 place-items-center rounded-xl bg-[#F0EBF9] text-[#6D4FA3]">
        <span className="text-[7px] font-semibold uppercase text-slate-400">
          {date.month}
        </span>
        <span className="text-sm font-bold leading-none text-[#6D4FA3]">
          {date.day}
        </span>
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-[10px] font-bold text-slate-800">
          {date.title}
        </p>
        <p className="mt-0.5 truncate text-[8px] text-slate-400">
          {date.meta || "Details not provided"}
        </p>
      </div>
      {date.badge && (
        <span className="shrink-0 rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-500">
          {date.badge}
        </span>
      )}
    </article>
  );
}

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
        "rounded-2xl border border-dashed border-slate-200 bg-white/70 px-4 py-8 text-center",
        className,
      )}
    >
      <span className="mx-auto grid h-10 w-10 place-items-center rounded-xl bg-slate-50 text-slate-400">
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

function formatTimeRange(start: string | null, end: string | null) {
  if (!start && !end) {
    return "Time not provided";
  }

  return `${start || "—"} – ${end || "—"}`;
}

function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "G";
}
