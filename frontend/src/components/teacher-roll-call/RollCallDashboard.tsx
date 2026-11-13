"use client";

import { useEffect, useMemo, useState } from "react";
import {
  AlertTriangle,
  BadgeCheck,
  Bell,
  CalendarDays,
  Check,
  CheckCheck,
  ChevronRight,
  ClipboardCheck,
  ClipboardList,
  GraduationCap,
  LayoutDashboard,
  LogOut,
  Megaphone,
  Menu,
  UserCheck,
  Users,
  UsersRound,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import {
  getStudentInitials,
  type RollCallData,
  type RollCallStatus,
  type RollCallStudent,
} from "./roll-call-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Overview", icon: LayoutDashboard, href: "/teacher/overview" },
  { label: "Timetable", icon: CalendarDays, href: "/teacher" },
  { label: "My class", icon: UsersRound, href: "/teacher/class" },
  { label: "Roll-call", icon: ClipboardCheck, href: "/teacher/roll-call", active: true },
  { label: "Marks", icon: ClipboardList, href: "#" },
  { label: "Guardians", icon: Users, href: "#" },
  { label: "Staff ID", icon: BadgeCheck, href: "#" },
];

const BOTTOM_NAV = [
  { label: "Classes", icon: UsersRound, href: "/teacher/class" },
  { label: "Attendance", icon: ClipboardCheck, href: "/teacher/roll-call", active: true },
  { label: "Marks", icon: ClipboardList, href: "#" },
  { label: "Guardians", icon: Users, href: "#" },
  { label: "Staff ID", icon: BadgeCheck, href: "#" },
];

/** Renders the teacher roll-call workspace for a resolved class roster. */
export default function RollCallDashboard({
  data,
  isPreview = true,
}: {
  data: RollCallData | null;
  isPreview?: boolean;
}) {
  const students = data?.students ?? [];
  const session = data?.session ?? null;
  const teacherName = session?.fullName?.trim() || "Teacher";
  const teacherRole = session?.role?.trim() || "Staff";
  const teacherSubject = session?.subject?.trim() || "No subject assigned";
  const teacherInitials =
    session?.initials?.trim() || initialsFromName(teacherName);
  const portalName = data?.portal?.name?.trim() || "School portal";
  const context = data?.context ?? {
    streamName: null,
    dateLabel: null,
    termLabel: null,
    registryRef: null,
  };

  const [statuses, setStatuses] = useState<Record<string, RollCallStatus>>({});
  const [menuOpen, setMenuOpen] = useState(false);
  const markedCount = Object.keys(statuses).length;
  const presentCount = useMemo(
    () => Object.values(statuses).filter((status) => status === "present").length,
    [statuses],
  );
  const absentCount = markedCount - presentCount;
  const progress = students.length
    ? Math.round((markedCount / students.length) * 100)
    : 0;

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = "";
    };
  }, [menuOpen]);

  useEffect(() => {
    const desktopQuery = window.matchMedia("(min-width: 1024px)");
    const closeOnDesktop = (event: MediaQueryListEvent) => {
      if (event.matches) {
        setMenuOpen(false);
      }
    };

    desktopQuery.addEventListener("change", closeOnDesktop);

    return () => {
      desktopQuery.removeEventListener("change", closeOnDesktop);
    };
  }, []);

  /** Sets or clears one learner's attendance mark. */
  function toggleStatus(id: string, status: RollCallStatus) {
    setStatuses((current) => {
      const next = { ...current };

      if (next[id] === status) {
        delete next[id];
      } else {
        next[id] = status;
      }

      return next;
    });
  }

  /** Marks every learner in the loaded roster as present. */
  function markAllPresent() {
    setStatuses(
      Object.fromEntries(students.map((student) => [student.id, "present"])) as Record<
        string,
        RollCallStatus
      >,
    );
  }

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#roll-call-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to roll-call
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[272px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          teacherName={teacherName}
          teacherRole={teacherRole}
          teacherSubject={teacherSubject}
          teacherInitials={teacherInitials}
        />
      </aside>

      <MobileAppBar
        portalName={portalName}
        teacherInitials={teacherInitials}
        menuOpen={menuOpen}
        onOpenMenu={() => setMenuOpen(true)}
      />

      <div className="lg:pl-[272px]">
        <DesktopHeader
          portalName={portalName}
          teacherName={teacherName}
          teacherRole={teacherRole}
          teacherInitials={teacherInitials}
        />

        <main
          id="roll-call-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1100px] px-4 pb-[260px] pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PageHeader context={context} />
          <ProgressCard
            marked={markedCount}
            total={students.length}
            present={presentCount}
            absent={absentCount}
            progress={progress}
            onMarkAllPresent={markAllPresent}
          />
          <StudentList
            students={students}
            statuses={statuses}
            onToggle={toggleStatus}
          />
          <GuardianDispatchNotice />
        </main>
      </div>

      <SubmissionCard
        marked={markedCount}
        total={students.length}
        registryRef={context.registryRef}
        termLabel={context.termLabel}
        isPreview={isPreview}
      />
      <BottomNav />

      {menuOpen && (
        <MobileDrawer
          teacherName={teacherName}
          teacherRole={teacherRole}
          teacherSubject={teacherSubject}
          teacherInitials={teacherInitials}
          onClose={() => setMenuOpen(false)}
        />
      )}
    </div>
  );
}

/** Shared navigation and identity panel for the sidebar and mobile drawer. */
function SidebarContent({
  teacherName,
  teacherRole,
  teacherSubject,
  teacherInitials,
  onNavigate,
}: {
  teacherName: string;
  teacherRole: string;
  teacherSubject: string;
  teacherInitials: string;
  onNavigate?: () => void;
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
              Teacher portal
            </p>
            <span className="rounded-full bg-brand-50 px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wide text-brand-700 ring-1 ring-brand-100">
              Staff
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
              onClick={onNavigate}
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
              {teacherInitials}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                {teacherName}
              </p>
              <p className="truncate text-[10px] text-slate-400">
                {teacherRole}
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                {teacherSubject}
              </p>
            </div>
          </div>
          <button
            type="button"
            disabled
            title="Sign out is not available in this preview."
            className="mt-3 inline-flex h-9 w-full cursor-not-allowed items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-[11px] font-semibold text-slate-400 opacity-70 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <LogOut className="h-3.5 w-3.5" />
            Sign out
          </button>
        </div>
      </div>
    </div>
  );
}

/** Mobile top bar with the menu trigger and portal identity. */
function MobileAppBar({
  portalName,
  teacherInitials,
  menuOpen,
  onOpenMenu,
}: {
  portalName: string;
  teacherInitials: string;
  menuOpen: boolean;
  onOpenMenu: () => void;
}) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <button
          type="button"
          onClick={onOpenMenu}
          aria-expanded={menuOpen}
          aria-controls="roll-call-drawer"
          className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open navigation menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="min-w-0 flex-1 text-center">
          <p className="truncate text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            {portalName}
          </p>
          <p className="truncate text-sm font-bold tracking-tight text-slate-900">
            Classes
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
            {teacherInitials}
          </button>
        </div>
      </div>
    </header>
  );
}

/** Sticky desktop header with breadcrumb and teacher identity. */
function DesktopHeader({
  portalName,
  teacherName,
  teacherRole,
  teacherInitials,
}: {
  portalName: string;
  teacherName: string;
  teacherRole: string;
  teacherInitials: string;
}) {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1100px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="max-w-[220px] truncate text-slate-700">
            {portalName}
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Classes</span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Roll-call</span>
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
              {teacherInitials}
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                {teacherName}
              </p>
              <p className="text-[10px] text-slate-400">{teacherRole}</p>
            </div>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

/** Shows the class, date and term metadata when the loader provides it. */
function PageHeader({ context }: { context: RollCallData["context"] }) {
  const metadata = [
    context.streamName,
    context.dateLabel,
    context.termLabel,
  ].filter((value): value is string => Boolean(value?.trim()));

  return (
    <div>
      <h1 className="text-2xl font-bold tracking-tight text-slate-900">
        Roll-call
      </h1>
      {metadata.length > 0 ? (
        <p className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[10px] text-slate-500">
          {metadata.map((value, index) => (
            <span key={`${value}-${index}`} className="flex items-center gap-2">
              {index > 0 && <span className="text-slate-300">•</span>}
              {value}
            </span>
          ))}
        </p>
      ) : (
        <p className="mt-1 text-[10px] text-slate-500">
          Roll-call context is not available yet.
        </p>
      )}
    </div>
  );
}

/** Summarises marked, present and absent totals with the completion bar. */
function ProgressCard({
  marked,
  total,
  present,
  absent,
  progress,
  onMarkAllPresent,
}: {
  marked: number;
  total: number;
  present: number;
  absent: number;
  progress: number;
  onMarkAllPresent: () => void;
}) {
  return (
    <section className="mt-4 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-lg font-bold tracking-tight text-slate-900">
            {marked} of {total} marked
          </p>
          <p className="mt-0.5 text-[9px] font-medium text-slate-400">
            Daily attendance entry
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3 text-[9px] font-semibold">
          <span className="inline-flex items-center gap-1.5 text-[#6D4FA3]">
            <span className="h-2 w-2 rounded-full bg-[#B9A3E0]" />
            Present {present}
          </span>
          <span className="inline-flex items-center gap-1.5 text-[#8A6A25]">
            <span className="h-2 w-2 rounded-full bg-[#E8C97A]" />
            Absent {absent}
          </span>
        </div>
      </div>

      <div className="mt-3 h-2 overflow-hidden rounded-full bg-slate-100">
        <div
          className="h-full rounded-full bg-brand-500 transition-all duration-300"
          style={{ width: `${progress}%` }}
        />
      </div>

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-[9px] leading-relaxed text-slate-400">
          Attendance records are retained and visible to authorized staff and
          guardians.
        </p>
        <button
          type="button"
          onClick={onMarkAllPresent}
          disabled={total === 0}
          className="inline-flex h-8 shrink-0 items-center justify-center gap-1.5 rounded-lg border border-brand-100 bg-[#E8F5F3] px-2.5 text-[9px] font-semibold text-brand-700 transition hover:bg-brand-50 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <CheckCheck className="h-3 w-3" />
          Mark all present
        </button>
      </div>
    </section>
  );
}

/** Lists the enrolled learners with accessible P/A attendance toggles. */
function StudentList({
  students,
  statuses,
  onToggle,
}: {
  students: RollCallStudent[];
  statuses: Record<string, RollCallStatus>;
  onToggle: (id: string, status: RollCallStatus) => void;
}) {
  return (
    <section className="mt-5">
      <div className="flex items-center justify-between gap-3 px-1">
        <h2 className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
          Enrolled students ({students.length})
        </h2>
        <p className="text-[9px] font-medium text-slate-400">
          Tap P or A to log
        </p>
      </div>

      {students.length > 0 ? (
        <div className="mt-2.5 space-y-2">
          {students.map((student) => {
            const status = statuses[student.id];

            return (
              <article
                key={student.id}
                className="flex items-center gap-3 rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card"
              >
                <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-[#F0EBF9] text-[9px] font-bold text-[#6D4FA3] ring-1 ring-white">
                  {getStudentInitials(student)}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-[11px] font-bold text-slate-800">
                    {student.name}
                  </p>
                  <p className="text-[9px] text-slate-400">
                    {student.admissionNumber}
                  </p>
                </div>

                <div
                  className="flex shrink-0 items-center gap-0.5 rounded-full bg-slate-50 p-0.5 ring-1 ring-slate-200/80"
                  role="group"
                  aria-label={`Attendance for ${student.name}`}
                >
                  <button
                    type="button"
                    aria-pressed={status === "present"}
                    onClick={() => onToggle(student.id, "present")}
                    className={cn(
                      "inline-flex h-8 items-center gap-1 rounded-full px-2.5 text-[9px] font-bold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
                      status === "present"
                        ? "bg-[#F0EBF9] text-[#6D4FA3] shadow-sm"
                        : "text-slate-400 hover:bg-white",
                    )}
                  >
                    <Check className="h-3 w-3" />
                    P
                  </button>
                  <button
                    type="button"
                    aria-pressed={status === "absent"}
                    onClick={() => onToggle(student.id, "absent")}
                    className={cn(
                      "inline-flex h-8 items-center gap-1 rounded-full px-2.5 text-[9px] font-bold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
                      status === "absent"
                        ? "bg-[#FEF3D6] text-[#8A6A25] shadow-sm"
                        : "text-slate-400 hover:bg-white",
                    )}
                  >
                    <AlertTriangle className="h-3 w-3" />
                    A
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      ) : (
        <div className="mt-2.5 rounded-2xl border border-dashed border-slate-300 bg-white px-4 py-10 text-center">
          <UsersRound className="mx-auto h-5 w-5 text-slate-400" />
          <p className="mt-2 text-[11px] font-semibold text-slate-700">
            No students available
          </p>
          <p className="mx-auto mt-1 max-w-xs text-[9px] leading-relaxed text-slate-500">
            The class roster will appear here when it is available.
          </p>
        </div>
      )}
    </section>
  );
}

/** Explains the guardian in-app notification sent after submission. */
function GuardianDispatchNotice() {
  return (
    <section className="mt-4 flex gap-3 rounded-2xl border border-brand-100 bg-[#E2F0ED] p-4">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-brand-600 shadow-sm">
        <Megaphone className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-[11px] font-bold text-brand-800">
          Guardian App Dispatch
        </p>
        <p className="mt-1 text-[9px] leading-relaxed text-slate-500">
          Guardians receive an in-app update after this roll-call is submitted.
          Health and clinical details are never shown.
        </p>
      </div>
    </section>
  );
}

/** Fixed submission summary and save action for the roll-call preview. */
function SubmissionCard({
  marked,
  total,
  registryRef,
  termLabel,
  isPreview,
}: {
  marked: number;
  total: number;
  registryRef: string | null;
  termLabel: string | null;
  isPreview: boolean;
}) {
  const ready = total > 0 && marked === total;
  const complete = marked > 0;
  const status = total === 0
    ? "No students to mark"
    : ready
      ? "Ready to commit"
      : complete
        ? "In progress"
        : "Not started";

  return (
    <div className="fixed inset-x-3 bottom-20 z-30 lg:inset-x-auto lg:bottom-6 lg:left-[296px] lg:right-6">
      <div className="mx-auto w-full max-w-[1100px] rounded-2xl border border-slate-200/80 bg-white p-3 shadow-2xl sm:p-4">
        <div className="flex items-center justify-between gap-3">
          <p className="text-[8px] font-medium text-slate-400">
            {registryRef || "No registry reference"} ·{" "}
            {termLabel || "No active term"}
          </p>
          <span
            className={cn(
              "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[8px] font-semibold",
              ready
                ? "bg-[#E2F0ED] text-brand-700"
                : complete
                  ? "bg-[#FEF3D6] text-[#8A6A25]"
                  : "bg-slate-100 text-slate-500",
            )}
          >
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                ready
                  ? "bg-brand-500"
                  : complete
                    ? "bg-[#D6A84A]"
                    : "bg-slate-400",
              )}
            />
            {status}
          </span>
        </div>

        <button
          type="button"
          disabled={isPreview || !complete}
          title={isPreview ? "Saving is not connected in this preview." : undefined}
          className="mt-2.5 inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-brand-500 text-[11px] font-bold text-white shadow-lg shadow-brand-500/20 transition hover:bg-brand-600 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <UserCheck className="h-4 w-4" />
          Save roll-call
        </button>
        {isPreview && (
          <p className="mt-2 text-center text-[8px] font-medium text-slate-400">
            Preview only — saving roll-call is not connected yet.
          </p>
        )}
      </div>
    </div>
  );
}

/** Fixed five-tab mobile navigation with Attendance marked as current. */
function BottomNav() {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-5 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
        {BOTTOM_NAV.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "relative flex flex-col items-center gap-1 rounded-xl py-2 text-[8px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
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

/** Accessible mobile drawer that traps focus and closes on Escape. */
function MobileDrawer({
  teacherName,
  teacherRole,
  teacherSubject,
  teacherInitials,
  onClose,
}: {
  teacherName: string;
  teacherRole: string;
  teacherSubject: string;
  teacherInitials: string;
  onClose: () => void;
}) {
  const drawerRef = useFocusTrap<HTMLDivElement>(true, onClose);

  return (
    <div className="fixed inset-0 z-50 lg:hidden print:hidden">
      <button
        type="button"
        tabIndex={-1}
        onClick={onClose}
        className="absolute inset-0 cursor-default bg-slate-900/40 backdrop-blur-sm animate-fade-in motion-reduce:animate-none"
        aria-label="Close navigation menu"
      />
      <div
        ref={drawerRef}
        id="roll-call-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="roll-call-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="roll-call-menu-title"
            className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400"
          >
            Navigation
          </h2>
          <button
            type="button"
            onClick={onClose}
            data-autofocus
            className="grid h-9 w-9 place-items-center rounded-xl bg-slate-50 text-slate-500 transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Close navigation menu"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="min-h-0 flex-1">
          <SidebarContent
            teacherName={teacherName}
            teacherRole={teacherRole}
            teacherSubject={teacherSubject}
            teacherInitials={teacherInitials}
            onNavigate={onClose}
          />
        </div>
      </div>
    </div>
  );
}

/** Derives up to two uppercase initials from a display name. */
function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "T";
}
