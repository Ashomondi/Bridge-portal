"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  AlertTriangle,
  ArrowRight,
  Bell,
  BookOpen,
  CalendarDays,
  Check,
  CheckCircle2,
  ChevronRight,
  ClipboardCheck,
  ClipboardList,
  Clock3,
  Coffee,
  GraduationCap,
  LayoutDashboard,
  Library,
  Megaphone,
  Menu,
  MoreHorizontal,
  Pencil,
  SlidersHorizontal,
  Sparkles,
  UsersRound,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import {
  ANNOUNCEMENT_TONE_BY_KIND,
  TODO_TONE_BY_KIND,
  type Announcement,
  type NextLesson,
  type TeacherOverviewData,
  type TimetableRow,
  type TodoItem,
} from "./teacher-overview-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href?: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  {
    label: "Overview",
    icon: LayoutDashboard,
    href: "/teacher/overview",
    active: true,
  },
  { label: "Timetable", icon: CalendarDays, href: "/teacher" },
  { label: "My class", icon: UsersRound, href: "/teacher/class" },
  { label: "Attendance", icon: ClipboardCheck },
  { label: "Marks", icon: ClipboardList },
  { label: "Resources", icon: Library },
];

const BOTTOM_NAV: {
  label: string;
  icon: LucideIcon;
  href?: string;
  active?: boolean;
  opensDrawer?: boolean;
}[] = [
  {
    label: "Overview",
    icon: LayoutDashboard,
    href: "/teacher/overview",
    active: true,
  },
  { label: "Timetable", icon: CalendarDays, href: "/teacher" },
  { label: "Class", icon: UsersRound, href: "/teacher/class" },
  { label: "Tasks", icon: CheckCircle2 },
  { label: "More", icon: MoreHorizontal, opensDrawer: true },
];

const TODO_ICON: Record<TodoItem["kind"], LucideIcon> = {
  warning: AlertTriangle,
  task: BookOpen,
  progress: ClipboardList,
  done: CheckCircle2,
};

const ANNOUNCEMENT_ICON: Record<Announcement["kind"], LucideIcon> = {
  check: Check,
  edit: Pencil,
  calendar: CalendarDays,
};

export default function TeacherOverviewDashboard({
  overview,
}: {
  overview: TeacherOverviewData | null;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const organizationName =
    overview?.portal?.organizationName?.trim() || "Staff workspace";
  const dateLabel = overview?.portal?.dateLabel?.trim() || null;
  const academicTerm = overview?.portal?.academicTerm?.trim() || null;
  const weekLabel = overview?.portal?.weekLabel?.trim() || null;
  const teacherName = overview?.teacher?.name?.trim() || null;
  const greeting =
    overview?.teacher?.greeting?.trim() ||
    (teacherName ? `Good morning, ${teacherName}` : "Teacher Overview");
  const session = overview?.session ?? null;
  const userFullName = session?.fullName?.trim() || "Teacher";
  const userRole = session?.role?.trim() || "Staff";
  const userSubject = session?.subject?.trim() || "No subject assigned";
  const userInitials = session?.initials?.trim() || initialsFromName(userFullName);
  const nextLesson = overview?.nextLesson ?? null;
  const todo = overview?.todo ?? { pendingCount: null, items: [] };
  const timetable = overview?.timetable ?? { dayLabel: null, rows: [] };
  const announcements = overview?.announcements ?? [];

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
    const handleBreakpointChange = (event: MediaQueryListEvent) => {
      if (event.matches) {
        setMenuOpen(false);
      }
    };

    desktopQuery.addEventListener("change", handleBreakpointChange);

    return () => {
      desktopQuery.removeEventListener("change", handleBreakpointChange);
    };
  }, []);

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#overview-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to overview
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          userFullName={userFullName}
          userRole={userRole}
          userSubject={userSubject}
          userInitials={userInitials}
        />
      </aside>

      <MobileAppBar
        organizationName={organizationName}
        userInitials={userInitials}
        menuOpen={menuOpen}
        onOpenMenu={() => setMenuOpen(true)}
      />

      <div className="lg:pl-[260px]">
        <DesktopHeader
          organizationName={organizationName}
          userFullName={userFullName}
          userRole={userRole}
          userInitials={userInitials}
        />

        <main
          id="overview-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <Greeting
            greeting={greeting}
            dateLabel={dateLabel}
            academicTerm={academicTerm}
            weekLabel={weekLabel}
          />

          <div className="mt-5">
            <NextLessonCard lesson={nextLesson} />
          </div>

          <div className="mt-6 grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(360px,1fr)]">
            <div className="order-1 min-w-0 xl:col-start-1 xl:row-start-1">
              <TodoSection todo={todo} />
            </div>

            <div className="order-2 min-w-0 xl:col-start-2 xl:row-span-2 xl:row-start-1 xl:sticky xl:top-[84px] xl:self-start">
              <TimetableWidget timetable={timetable} />
            </div>

            <div className="order-3 min-w-0 xl:col-start-1 xl:row-start-2">
              <AnnouncementsCard announcements={announcements} />
            </div>
          </div>
        </main>
      </div>

      <BottomNav onOpenMenu={() => setMenuOpen(true)} />

      {menuOpen && (
        <MobileDrawer
          userFullName={userFullName}
          userRole={userRole}
          userSubject={userSubject}
          userInitials={userInitials}
          onClose={() => setMenuOpen(false)}
        />
      )}
    </div>
  );
}

function SidebarContent({
  userFullName,
  userRole,
  userSubject,
  userInitials,
  onNavigate,
}: {
  userFullName: string;
  userRole: string;
  userSubject: string;
  userInitials: string;
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
            Bridge <span className="text-brand-600">Portal</span>
          </p>
          <div className="mt-0.5 flex items-center gap-1.5">
            <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-600">
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
                    : "text-slate-600 group-hover:text-brand-600",
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
              onClick={onNavigate}
              aria-current={item.active ? "page" : undefined}
              className={className}
            >
              {content}
            </Link>
          );
        })}
      </nav>

      <div className="mt-auto space-y-4 pt-6">
        <div className="rounded-2xl border border-brand-100 bg-brand-50/60 p-4">
          <div className="flex items-center gap-2 text-brand-700">
            <Sparkles className="h-4 w-4" />
            <p className="text-xs font-bold">Quick tip</p>
          </div>
          <p className="mt-2 text-[11px] leading-relaxed text-slate-500">
            Use the timetable widget to jump straight into your next lesson.
          </p>
        </div>

        <div className="flex items-center gap-2.5 border-t border-slate-100 pt-4">
          <span className="grid h-9 w-9 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
            {userInitials}
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-xs font-semibold text-slate-800">
              {userFullName}
            </p>
            <p className="text-[10px] text-slate-600">{userRole}</p>
            <p className="truncate text-[9px] font-medium text-brand-600">
              {userSubject}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function MobileAppBar({
  organizationName,
  userInitials,
  menuOpen,
  onOpenMenu,
}: {
  organizationName: string;
  userInitials: string;
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
          aria-controls="overview-navigation-drawer"
          className="grid h-11 w-11 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
          aria-label="Open navigation menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="min-w-0 flex-1 text-center">
          <p className="truncate text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-600">
            {organizationName}
          </p>
          <p className="truncate text-sm font-bold tracking-tight text-slate-900">
            Teacher Portal
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          <button
            type="button"
            disabled
            aria-disabled="true"
            aria-pressed={false}
            title="Notifications are unavailable until the notifications service is connected."
            className="relative grid h-11 w-11 place-items-center rounded-xl text-slate-500 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="Notifications unavailable"
          >
            <Bell className="h-5 w-5" />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
            <span className="sr-only">Unread notifications</span>
          </button>
          <button
            type="button"
            disabled
            aria-disabled="true"
            title="Profile menu is unavailable until authentication is connected."
            className="grid h-11 w-11 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="User profile unavailable"
          >
            {userInitials}
          </button>
        </div>
      </div>
    </header>
  );
}

function DesktopHeader({
  organizationName,
  userFullName,
  userRole,
  userInitials,
}: {
  organizationName: string;
  userFullName: string;
  userRole: string;
  userInitials: string;
}) {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-600">
          <span className="max-w-[220px] truncate text-slate-700">
            {organizationName}
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Overview</span>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            type="button"
            disabled
            aria-disabled="true"
            aria-pressed={false}
            title="Notifications are unavailable until the notifications service is connected."
            className="relative grid h-11 w-11 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="Notifications unavailable"
          >
            <Bell className="h-[18px] w-[18px]" />
            <span className="absolute right-2.5 top-2.5 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
            <span className="sr-only">Unread notifications</span>
          </button>

          <button
            type="button"
            disabled
            aria-disabled="true"
            aria-expanded={false}
            title="Profile menu is unavailable until authentication is connected."
            className="flex min-h-11 items-center gap-2.5 rounded-xl border border-slate-200/80 bg-white py-1.5 pl-1.5 pr-3 shadow-sm transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
          >
            <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-[10px] font-bold text-white">
              {userInitials}
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                {userFullName}
              </p>
              <p className="text-[10px] text-slate-600">{userRole}</p>
            </div>
            <ChevronRight className="h-3.5 w-3.5 text-slate-500" />
          </button>
        </div>
      </div>
    </header>
  );
}

function Greeting({
  greeting,
  dateLabel,
  academicTerm,
  weekLabel,
}: {
  greeting: string;
  dateLabel: string | null;
  academicTerm: string | null;
  weekLabel: string | null;
}) {
  const metadata = [dateLabel, academicTerm, weekLabel].filter(
    (value): value is string => Boolean(value),
  );

  return (
    <div className="flex items-start justify-between gap-4">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(0,122,120,0.10)]" />
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-600">
            Teacher overview
          </p>
        </div>
        <h1 className="mt-2 text-[26px] font-bold leading-tight tracking-tight text-slate-900 sm:text-3xl">
          {greeting}
        </h1>
        {metadata.length > 0 ? (
          <div className="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-slate-500">
            {metadata.map((value, index) => (
              <span key={`${value}-${index}`} className="flex items-center gap-2">
                {index > 0 && <span className="text-slate-300">•</span>}
                {index === 0 && (
                  <CalendarDays className="h-3.5 w-3.5 text-slate-600" />
                )}
                {value}
              </span>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-xs text-slate-500">
            Date and academic period will appear here when available.
          </p>
        )}
      </div>

      <button
        type="button"
        disabled
        aria-disabled="true"
        title="Overview view settings are unavailable until the preference service is connected."
        className="grid h-11 w-11 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
        aria-label="Overview view settings unavailable"
      >
        <SlidersHorizontal className="h-4 w-4" />
      </button>
    </div>
  );
}

function NextLessonCard({ lesson }: { lesson: NextLesson | null }) {
  const now = useCurrentTime();
  const countdown = lesson ? getCountdownLabel(lesson, now) : null;
  const safePlanHref = sanitizeHref(lesson?.planHref);

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-card sm:p-6">
      <div className="flex items-center justify-between gap-3">
        <p className="text-[9px] font-semibold uppercase tracking-[0.2em] text-slate-600">
          Next lesson
        </p>
        {lesson && countdown && (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700 ring-1 ring-brand-100">
            <span className="relative flex h-1.5 w-1.5">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-400 opacity-70 motion-reduce:animate-none" />
              <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-brand-500" />
            </span>
            {countdown}
          </span>
        )}
      </div>

      {lesson ? (
        <>
          <h2 className="mt-3 text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">
            {lesson.title}
          </h2>

          {(lesson.group || lesson.startTime) && (
            <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[11px] font-medium text-slate-500">
              {lesson.group && (
                <span className="inline-flex items-center gap-1.5">
                  <UsersRound className="h-3.5 w-3.5 text-slate-600" />
                  {lesson.group}
                </span>
              )}
              {lesson.startTime && (
                <span className="inline-flex items-center gap-1.5">
                  <Clock3 className="h-3.5 w-3.5 text-slate-600" />
                  {lesson.startTime}
                </span>
              )}
            </div>
          )}

          {(lesson.unitTopic || safePlanHref) && (
            <div className="mt-4 flex items-center justify-between gap-3 rounded-xl bg-[#E8F5F3] px-3.5 py-3">
              <div className="min-w-0">
                <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-brand-600">
                  Unit / topic
                </p>
                <p className="mt-0.5 truncate text-xs font-semibold text-slate-800">
                  {lesson.unitTopic || "Not provided"}
                </p>
              </div>
              {safePlanHref && (
                <a
                  href={safePlanHref}
                  rel="noreferrer noopener"
                  className="inline-flex shrink-0 items-center gap-1 rounded-lg px-2 py-1 text-[10px] font-bold text-brand-700 transition hover:bg-white/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
                >
                  Plan
                  <ArrowRight className="h-3.5 w-3.5" />
                </a>
              )}
            </div>
          )}
        </>
      ) : (
        <EmptyState
          icon={Clock3}
          title="No upcoming lesson"
          body="Your next lesson will appear here when the timetable is available."
        />
      )}
    </section>
  );
}

function TodoSection({
  todo,
}: {
  todo: TeacherOverviewData["todo"];
}) {
  const pendingCount =
    todo.items.length > 0
      ? todo.items.filter((item) => item.kind !== "done").length
      : todo.pendingCount;

  return (
    <section>
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          To do
        </h2>
        <span className="rounded-full bg-white px-2.5 py-1 text-[10px] font-semibold text-slate-600 shadow-sm ring-1 ring-slate-200/80">
          {pendingCount === null ? "Pending" : `${pendingCount} pending`}
        </span>
      </div>

      {todo.items.length > 0 ? (
        <div className="mt-3 space-y-2.5">
          {todo.items.map((item) => (
            <TodoCard key={item.id} item={item} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={CheckCircle2}
          title="No tasks available"
          body="Pending tasks will appear here when they are assigned."
          className="mt-3"
        />
      )}
    </section>
  );
}

function TodoCard({ item }: { item: TodoItem }) {
  const Icon = TODO_ICON[item.kind];
  const tone = TODO_TONE_BY_KIND[item.kind];

  return (
    <div
      aria-disabled="true"
      title="Task details are unavailable until the tasks service is connected."
      className="flex items-center gap-3 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card"
    >
      <span
        className={cn(
          "grid h-10 w-10 shrink-0 place-items-center rounded-xl",
          tone === "gold" && "bg-[#FEF3D6] text-[#9A6A10]",
          tone === "lavender" && "bg-[#F0EBF9] text-[#6D4FA3]",
          tone === "slate" && "bg-slate-100 text-slate-600",
          tone === "mint" && "bg-[#E2F0ED] text-brand-700",
        )}
      >
        <Icon className="h-4 w-4" />
      </span>

      <div className="min-w-0 flex-1">
        <p className="truncate text-[12px] font-bold text-slate-800">
          {item.title}
        </p>
        {item.meta && (
          <p className="mt-0.5 truncate text-[10px] text-slate-600">
            {item.meta}
          </p>
        )}
        <span
          className={cn(
            "mt-1.5 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[9px] font-semibold ring-1",
            tone === "gold" &&
              "bg-[#FEF3D6] text-[#8A6A25] ring-[#EBD7A5]",
            tone === "lavender" &&
              "bg-[#F0EBF9] text-[#6D4FA3] ring-[#E4DBF5]",
            tone === "slate" && "bg-slate-100 text-slate-600 ring-slate-200",
            tone === "mint" && "bg-[#E2F0ED] text-brand-700 ring-brand-100",
          )}
        >
          {tone === "gold" && <AlertTriangle className="h-2.5 w-2.5" />}
          {tone === "mint" && <CheckCircle2 className="h-2.5 w-2.5" />}
          {item.status}
        </span>
      </div>

      <ChevronRight className="h-4 w-4 shrink-0 text-slate-600" />
    </div>
  );
}

function TimetableWidget({
  timetable,
}: {
  timetable: TeacherOverviewData["timetable"];
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card sm:p-4">
      <div className="flex items-center justify-between gap-3 px-1 pb-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Today&apos;s timetable
        </h2>
        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700 ring-1 ring-brand-100">
          {timetable.dayLabel?.trim() || "—"}
        </span>
      </div>

      {timetable.rows.length > 0 ? (
        <div className="space-y-2">
          {timetable.rows.map((row) => (
            <TimetableRowCard key={row.id} row={row} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={CalendarDays}
          title="No periods available"
          body="Today's timetable will appear here when it is published."
        />
      )}
    </section>
  );
}

function TimetableRowCard({ row }: { row: TimetableRow }) {
  if (row.kind === "break") {
    return (
      <div className="flex items-center justify-between gap-3 rounded-xl bg-slate-50 px-3 py-2.5">
        <span className="inline-flex items-center gap-1.5 text-[10px] font-medium text-slate-500">
          <Coffee className="h-3.5 w-3.5 text-slate-600" />
          {row.label?.trim() || "Break"}
        </span>
        <span className="text-[10px] font-semibold text-slate-600">
          {formatTimeRange(row.startTime, row.endTime)}
        </span>
      </div>
    );
  }

  const active = row.kind === "active";
  const completed = row.kind === "completed";
  const StatusIcon = active ? ArrowRight : completed ? Check : Clock3;

  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-xl p-3",
        active
          ? "border border-[#E4DBF5] bg-[#F0EBF9]"
          : "bg-[#E8F5F3]",
        row.kind === "upcoming" && "bg-[#E8F5F3]/70",
      )}
    >
      <div className="w-[58px] shrink-0">
        <p
          className={cn(
            "text-[9px] font-semibold",
            active ? "text-[#6D4FA3]" : "text-slate-500",
          )}
        >
          {row.startTime || "—"}
        </p>
        <p
          className={cn(
            "text-[10px] font-medium",
            active ? "text-[#8A6DBC]" : "text-slate-600",
          )}
        >
          {row.endTime || "—"}
        </p>
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <p className="truncate text-[11px] font-bold text-slate-800">
            {row.subject?.trim() || "Subject not provided"}
          </p>
          {active && (
            <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-white px-1.5 py-0.5 text-[8px] font-bold text-[#6D4FA3]">
              <span className="h-1 w-1 rounded-full bg-[#7C5CBF]" />
              Now
            </span>
          )}
        </div>
        <p className="truncate text-[10px] text-slate-500">
          {row.location?.trim() || "Location not provided"}
        </p>
      </div>
      <span
        className={cn(
          "grid h-6 w-6 shrink-0 place-items-center rounded-full bg-white/80",
          active ? "text-[#6D4FA3]" : completed ? "text-brand-600" : "text-slate-600",
        )}
      >
        <StatusIcon className="h-3.5 w-3.5" />
      </span>
    </div>
  );
}

function AnnouncementsCard({
  announcements,
}: {
  announcements: Announcement[];
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Staff announcements
        </h2>
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <Megaphone className="h-4 w-4" />
        </span>
      </div>

      {announcements.length > 0 ? (
        <div className="mt-3 divide-y divide-slate-100">
          {announcements.map((item) => {
            const Icon = ANNOUNCEMENT_ICON[item.kind];
            const tone = ANNOUNCEMENT_TONE_BY_KIND[item.kind];

            return (
              <div
                key={item.id}
                className="flex gap-3 py-3 first:pt-0 last:pb-0"
              >
                <span
                  className={cn(
                    "grid h-8 w-8 shrink-0 place-items-center rounded-full",
                    tone === "mint" && "bg-[#E2F0ED] text-brand-700",
                    tone === "lavender" &&
                      "bg-[#F0EBF9] text-[#6D4FA3]",
                    tone === "slate" && "bg-slate-100 text-slate-500",
                  )}
                >
                  <Icon className="h-3.5 w-3.5" />
                </span>
                <div className="min-w-0">
                  <p className="text-[11px] font-bold text-slate-800">
                    {item.title}
                  </p>
                  <p className="mt-0.5 text-[10px] leading-relaxed text-slate-500">
                    {item.description}
                  </p>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <EmptyState
          icon={Megaphone}
          title="No announcements available"
          body="Staff updates will appear here when they are published."
          className="mt-3"
        />
      )}

      <span
        aria-disabled="true"
        title="Notification history is unavailable until the notifications service is connected."
        className="mt-3 flex cursor-default items-center justify-center gap-1.5 border-t border-slate-100 pt-3 text-[10px] font-semibold text-slate-500 opacity-70"
      >
        See all notifications unavailable
        <ArrowRight className="h-3.5 w-3.5" />
      </span>
    </section>
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
        "rounded-xl border border-dashed border-slate-200 bg-slate-50/60 px-4 py-7 text-center",
        className,
      )}
    >
      <span className="mx-auto grid h-10 w-10 place-items-center rounded-xl bg-white text-slate-600 shadow-sm">
        <Icon className="h-4 w-4" />
      </span>
      <p className="mt-2.5 text-xs font-semibold text-slate-700">{title}</p>
      <p className="mx-auto mt-1 max-w-xs text-[10px] leading-relaxed text-slate-500">
        {body}
      </p>
    </div>
  );
}

function BottomNav({ onOpenMenu }: { onOpenMenu: () => void }) {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-5 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
        {BOTTOM_NAV.map((item) => {
          const Icon = item.icon;
          const className = cn(
            "relative flex min-h-11 flex-col items-center justify-center gap-1 rounded-xl py-2 text-[9px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
            item.active
              ? "text-brand-600"
              : "text-slate-500 hover:text-slate-700",
            !item.href && !item.opensDrawer && "cursor-default opacity-60",
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

          if (item.opensDrawer) {
            return (
              <button
                key={item.label}
                type="button"
                onClick={onOpenMenu}
                className={className}
              >
                {content}
              </button>
            );
          }

          if (item.href) {
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
          }

          return (
            <span
              key={item.label}
              aria-disabled="true"
              className={className}
            >
              {content}
            </span>
          );
        })}
      </div>
    </nav>
  );
}

function MobileDrawer({
  userFullName,
  userRole,
  userSubject,
  userInitials,
  onClose,
}: {
  userFullName: string;
  userRole: string;
  userSubject: string;
  userInitials: string;
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
        id="overview-navigation-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="overview-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="overview-menu-title"
            className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-600"
          >
            Navigation
          </h2>
          <button
            type="button"
            onClick={onClose}
            data-autofocus
            className="grid h-9 w-9 place-items-center rounded-xl bg-slate-50 text-slate-500 transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600"
            aria-label="Close navigation menu"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="min-h-0 flex-1">
          <SidebarContent
            userFullName={userFullName}
            userRole={userRole}
            userSubject={userSubject}
            userInitials={userInitials}
            onNavigate={onClose}
          />
        </div>
      </div>
    </div>
  );
}

/** Returns a client clock that refreshes the countdown without hydration drift. */
function useCurrentTime() {
  const [now, setNow] = useState<number | null>(null);

  useEffect(() => {
    setNow(Date.now());
    const timer = window.setInterval(() => setNow(Date.now()), 30_000);

    return () => {
      window.clearInterval(timer);
    };
  }, []);

  return now;
}

/** Builds a guarded countdown from timestamps, falling back to supplied minutes. */
function getCountdownLabel(lesson: NextLesson, now: number | null) {
  if (lesson.startAt) {
    const start = Date.parse(lesson.startAt);

    if (!Number.isNaN(start)) {
      if (now === null) {
        return null;
      }

      const difference = start - now;

      if (difference > 60_000) {
        return `in ${Math.ceil(difference / 60_000)} min`;
      }

      if (difference > -60_000) {
        return "Starting now";
      }

      return `Started ${Math.max(1, Math.floor(-difference / 60_000))} min ago`;
    }
  }

  if (lesson.countdownMinutes === null) {
    return null;
  }

  if (lesson.countdownMinutes > 0) {
    return `in ${lesson.countdownMinutes} min`;
  }

  if (lesson.countdownMinutes === 0) {
    return "Starting now";
  }

  return "Lesson in progress";
}

/** Allows only same-origin relative paths and HTTP(S) destinations. */
function sanitizeHref(value: string | null | undefined) {
  if (!value) {
    return null;
  }

  const href = value.trim();

  if (/^\/(?!\/)/.test(href) || /^https?:\/\//i.test(href)) {
    return href;
  }

  return null;
}

function formatTimeRange(start: string | null, end: string | null) {
  if (!start && !end) {
    return "—";
  }

  return `${start || "—"} – ${end || "—"}`;
}

function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "T";
}
