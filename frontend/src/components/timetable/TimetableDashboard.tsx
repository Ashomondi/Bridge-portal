"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import {
  Bell,
  BookOpen,
  CalendarDays,
  Check,
  ChevronRight,
  ChevronsUpDown,
  ClipboardCheck,
  Coffee,
  DoorOpen,
  GraduationCap,
  HelpCircle,
  Info,
  LayoutDashboard,
  LayoutGrid,
  LayoutList,
  Library,
  ListChecks,
  LogOut,
  MapPin,
  Menu,
  Printer,
  RefreshCw,
  Search,
  Settings,
  ShieldCheck,
  Sparkles,
  Users,
  UsersRound,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import {
  getScheduleForDay,
  getWeekSummary,
  type TimetableEntry,
  type WeekDay,
} from "./timetable-data";
import {
  ClassesNavIcon,
  DashboardNavIcon,
  MoreNavIcon,
  ResourcesNavIcon,
  TimetableNavIcon,
} from "./nav-icons";

type TimetableDashboardProps = {
  week: WeekDay[];
};

type SyncState = "idle" | "syncing" | "done";

const SIDEBAR_ITEMS: {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
}[] = [
  { label: "Dashboard", icon: LayoutDashboard, href: "#" },
  { label: "My Timetable", icon: CalendarDays, href: "/teacher", active: true },
  { label: "My Classes", icon: UsersRound, href: "#" },
  { label: "Attendance", icon: ClipboardCheck, href: "#" },
  { label: "Resources", icon: Library, href: "#" },
  { label: "Settings", icon: Settings, href: "#" },
];

const BOTTOM_NAV = [
  { label: "Home", icon: DashboardNavIcon },
  { label: "Timetable", icon: TimetableNavIcon, active: true },
  { label: "Classes", icon: ClassesNavIcon },
  { label: "Library", icon: ResourcesNavIcon },
  { label: "More", icon: MoreNavIcon },
];

const BADGE_STYLES: Record<string, string> = {
  class: "bg-slate-100 text-slate-600 ring-slate-200",
  support: "bg-[#F0EBF9] text-[#6D4FA3] ring-[#E4DBF5]",
  live: "bg-[#7C5CBF] text-white ring-[#7C5CBF]",
  planning: "bg-brand-50 text-brand-700 ring-brand-100",
  pastoral: "bg-amber-50 text-amber-700 ring-amber-200",
};

const CARD_STYLES: Record<string, string> = {
  class: "border-slate-200/80 bg-white",
  support: "border-slate-200/80 bg-white",
  live: "border-[#D7C8F2] bg-[#F7F3FD] shadow-[0_12px_34px_-24px_rgba(109,79,163,0.65)]",
  planning: "border-brand-100 bg-[#F1FAF8]",
  pastoral: "border-slate-200/80 bg-white",
};

const TIME_STYLES: Record<string, string> = {
  class: "text-slate-900",
  support: "text-slate-900",
  live: "text-[#6D4FA3]",
  planning: "text-brand-700",
  pastoral: "text-slate-900",
};

const DIVIDER_STYLES: Record<string, string> = {
  class: "bg-slate-100",
  support: "bg-slate-100",
  live: "bg-[#D7C8F2]",
  planning: "bg-brand-100",
  pastoral: "bg-slate-100",
};

export default function TimetableDashboard({ week }: TimetableDashboardProps) {
  const todayIndex = week.findIndex((day) => day.isToday);
  const initialIndex = todayIndex >= 0 ? todayIndex : Math.max(week.length - 1, 0);

  const [selectedId, setSelectedId] = useState(
    week[initialIndex]?.id ?? week[0]?.id,
  );
  const [menuOpen, setMenuOpen] = useState(false);
  const [compact, setCompact] = useState(false);
  const [syncState, setSyncState] = useState<SyncState>("idle");
  const syncTimer = useRef<number | null>(null);

  const selectedIndex = Math.max(
    0,
    week.findIndex((day) => day.id === selectedId),
  );
  const selectedDay = week[selectedIndex] ?? week[0];
  const schedule = useMemo(
    () => getScheduleForDay(selectedIndex),
    [selectedIndex],
  );
  const weekSummary = useMemo(() => getWeekSummary(), []);
  const nextEntry =
    schedule.find((entry) => entry.kind === "live") ??
    schedule.find((entry) => entry.kind === "class") ??
    schedule.find((entry) => entry.kind !== "break");

  const periods = schedule.filter((entry) => entry.kind !== "break");
  const teaching = periods.filter((entry) =>
    ["class", "support", "live"].includes(entry.kind),
  );
  const progress = Math.round(((selectedIndex + 1) / week.length) * 100);

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setMenuOpen(false);
      }
    };

    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [menuOpen]);

  useEffect(() => {
    return () => {
      if (syncTimer.current !== null) {
        window.clearTimeout(syncTimer.current);
      }
    };
  }, []);

  function handleSync() {
    if (syncState === "syncing") {
      return;
    }

    setSyncState("syncing");
    syncTimer.current = window.setTimeout(() => {
      setSyncState("done");
    }, 900);
  }

  return (
    <div className="min-h-screen bg-[#F6F8F8] text-slate-900">
      <MobileAppBar onOpenMenu={() => setMenuOpen(true)} />

      <DesktopSidebar />

      <div className="lg:pl-[264px]">
        <DesktopHeader />

        <main className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-8 print:max-w-none print:p-0">
          <PageTitleBar
            compact={compact}
            onToggleCompact={() => setCompact((current) => !current)}
          />

          <NotificationBanner />

          <section
            aria-label="Week date selector"
            className="mt-4 rounded-2xl border border-slate-200/80 bg-white p-1.5 shadow-card sm:p-2 print:hidden"
          >
            <div className="grid grid-cols-5 gap-1 sm:gap-2">
              {week.map((day) => {
                const active = day.id === selectedDay?.id;

                return (
                  <button
                    key={day.id}
                    type="button"
                    onClick={() => setSelectedId(day.id)}
                    aria-pressed={active}
                    aria-label={`${day.fullLabel}${day.isToday ? ", today" : ""}`}
                    className={cn(
                      "group relative flex min-w-0 flex-col items-center rounded-xl px-1 py-2.5 transition-all duration-200 sm:py-3.5",
                      active
                        ? "bg-brand-500 text-white shadow-lg shadow-brand-500/20"
                        : "text-slate-600 hover:bg-slate-50",
                    )}
                  >
                    <span
                      className={cn(
                        "mb-0.5 h-4 rounded-full px-1.5 text-[8px] font-semibold leading-4 tracking-wide",
                        day.isToday
                          ? active
                            ? "bg-white/20 text-white"
                            : "bg-brand-50 text-brand-700"
                          : "bg-transparent text-transparent",
                      )}
                    >
                      {day.isToday ? "Today" : "•"}
                    </span>
                    <span
                      className={cn(
                        "text-[9px] font-semibold uppercase tracking-[0.12em] sm:text-[10px]",
                        active ? "text-white/80" : "text-slate-400",
                      )}
                    >
                      {day.weekdayShort}
                    </span>
                    <span className="mt-0.5 text-lg font-bold leading-none tracking-tight sm:text-xl">
                      {day.date}
                    </span>
                    <span
                      className={cn(
                        "mt-1 text-[9px] font-medium uppercase tracking-wide",
                        active ? "text-white/70" : "text-slate-400",
                      )}
                    >
                      {day.month}
                    </span>
                    <span
                      className={cn(
                        "mt-1.5 h-0.5 w-5 rounded-full sm:mt-2",
                        active ? "bg-white/80" : "bg-transparent",
                      )}
                    />
                  </button>
                );
              })}
            </div>
          </section>

          <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-[minmax(0,1fr)_336px]">
            <div className="min-w-0">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
                <div>
                  <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-brand-600">
                    Selected day
                  </p>
                  <h2 className="mt-1 text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">
                    {selectedDay?.weekday ?? "Today"} schedule
                  </h2>
                  <p className="mt-1 text-sm text-slate-500">
                    {selectedDay?.fullLabel} · [Term label]
                  </p>
                </div>

                <div className="flex w-full items-center gap-2 rounded-2xl border border-slate-200/80 bg-white px-3 py-2.5 shadow-sm sm:w-auto">
                  <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-50 text-brand-600">
                    <BookOpen className="h-4 w-4" />
                  </span>
                  <div className="min-w-0 flex-1 sm:flex-none">
                    <p className="text-sm font-bold leading-tight text-slate-900">
                      {teaching.length} teaching
                    </p>
                    <p className="text-[10px] font-medium text-slate-400">
                      periods today
                    </p>
                  </div>
                  <span className="ml-1 shrink-0 rounded-full bg-brand-500 px-2.5 py-1 text-[10px] font-semibold text-white">
                    {periods.length} periods
                  </span>
                </div>
              </div>

              <div
                className={cn(
                  "mt-4 space-y-3",
                  compact ? "space-y-2" : "space-y-3",
                )}
              >
                {schedule.map((entry) =>
                  entry.kind === "break" ? (
                    <BreakCard key={entry.id} entry={entry} />
                  ) : (
                    <ScheduleCard
                      key={entry.id}
                      entry={entry}
                      compact={compact}
                    />
                  ),
                )}
              </div>

              <SyncCallout state={syncState} onSync={handleSync} />
            </div>

            <aside className="hidden space-y-4 lg:block print:hidden">
              <WeekOverview summary={weekSummary} progress={progress} />
              <NextUp
                entry={nextEntry}
                day={selectedDay}
                isToday={Boolean(selectedDay?.isToday)}
              />
              <UpcomingCard />
            </aside>
          </div>
        </main>
      </div>

      <BottomNav />

      {menuOpen && <MobileDrawer onClose={() => setMenuOpen(false)} />}
    </div>
  );
}

function MobileAppBar({ onOpenMenu }: { onOpenMenu: () => void }) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <button
          type="button"
          onClick={onOpenMenu}
          className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open navigation menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="min-w-0 flex-1 text-center">
          <p className="truncate text-sm font-bold tracking-tight text-slate-900">
            [Organization Name]
          </p>
          <p className="truncate text-[10px] font-medium uppercase tracking-[0.16em] text-slate-400">
            Academic workspace
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          <button
            type="button"
            className="relative grid h-10 w-10 place-items-center rounded-xl text-slate-500 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Notifications, 1 unread"
          >
            <Bell className="h-5 w-5" />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
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
    <header className="sticky top-0 z-30 hidden border-b border-slate-200/60 bg-[#F6F8F8]/85 backdrop-blur-xl lg:block print:hidden">
      <div className="mx-auto flex h-16 w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="text-slate-700">Teacher portal</span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>My Timetable</span>
        </div>

        <div className="flex items-center gap-2.5">
          <label className="group relative hidden xl:block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <input
              type="search"
              placeholder="Search timetable..."
              className="h-10 w-56 rounded-xl border border-slate-200/80 bg-white pl-9 pr-3 text-sm text-slate-700 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-brand-300 focus:ring-4 focus:ring-brand-500/10"
            />
          </label>

          <button
            type="button"
            className="relative grid h-10 w-10 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Notifications, 1 unread"
          >
            <Bell className="h-[18px] w-[18px]" />
            <span className="absolute right-2.5 top-2.5 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>

          <div className="flex items-center gap-2.5 rounded-xl border border-slate-200/80 bg-white py-1.5 pl-1.5 pr-3 shadow-sm">
            <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-[10px] font-bold text-white">
              [AB]
            </span>
            <div className="min-w-0">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                [User Name]
              </p>
              <p className="text-[10px] text-slate-400">[Role]</p>
            </div>
            <ChevronsUpDown className="h-3.5 w-3.5 text-slate-400" />
          </div>
        </div>
      </div>
    </header>
  );
}

function DesktopSidebar() {
  return (
    <aside className="fixed inset-y-0 left-0 z-40 hidden w-[264px] flex-col border-r border-slate-200/70 bg-white px-5 py-6 lg:flex print:hidden">
      <div className="flex items-center gap-3 px-1">
        <span className="grid h-11 w-11 place-items-center rounded-2xl bg-brand-500 text-white shadow-lg shadow-brand-500/20">
          <GraduationCap className="h-[22px] w-[22px]" />
        </span>
        <div>
          <p className="text-[15px] font-bold tracking-tight text-slate-900">
            Ikonex <span className="text-brand-600">Academy</span>
          </p>
          <p className="text-[9px] font-semibold uppercase tracking-[0.2em] text-slate-400">
            School portal
          </p>
        </div>
      </div>

      <button
        type="button"
        className="mt-7 flex w-full items-center gap-3 rounded-2xl border border-slate-200/80 bg-slate-50/80 p-3 text-left transition hover:border-brand-200 hover:bg-brand-50/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/30"
      >
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-brand-600 shadow-sm">
          <CalendarDays className="h-4 w-4" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            Organization
          </span>
          <span className="block truncate text-xs font-semibold text-slate-700">
            [Organization Name]
          </span>
        </span>
        <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 text-slate-400" />
      </button>

      <nav className="mt-6 space-y-1" aria-label="Primary">
        {SIDEBAR_ITEMS.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href={item.href}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-[13px] font-medium transition-all",
                item.active
                  ? "bg-brand-500 text-white shadow-lg shadow-brand-500/15"
                  : "text-slate-500 hover:bg-slate-50 hover:text-slate-800",
              )}
            >
              <Icon
                className={cn(
                  "h-[18px] w-[18px]",
                  item.active
                    ? "text-white"
                    : "text-slate-400 group-hover:text-brand-600",
                )}
              />
              <span className="flex-1">{item.label}</span>
              {item.active && (
                <span className="h-1.5 w-1.5 rounded-full bg-white/80" />
              )}
            </a>
          );
        })}
      </nav>

      <div className="mt-auto space-y-4">
        <div className="rounded-2xl border border-brand-100 bg-brand-50/60 p-4">
          <div className="flex items-center gap-2 text-brand-700">
            <Sparkles className="h-4 w-4" />
            <p className="text-xs font-bold">Timetable tips</p>
          </div>
          <p className="mt-2 text-[11px] leading-relaxed text-slate-500">
            Use the layout control to switch between timeline and compact
            views.
          </p>
          <button
            type="button"
            className="mt-3 inline-flex items-center gap-1 text-[11px] font-semibold text-brand-700 hover:text-brand-800"
          >
            <HelpCircle className="h-3.5 w-3.5" />
            View guide
          </button>
        </div>

        <div className="flex items-center gap-2.5 border-t border-slate-100 pt-4">
          <span className="grid h-9 w-9 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
            [AB]
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-xs font-semibold text-slate-800">
              [User Name]
            </p>
            <p className="text-[10px] text-slate-400">[Role]</p>
          </div>
          <button
            type="button"
            className="grid h-8 w-8 place-items-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-700"
            aria-label="Sign out"
          >
            <LogOut className="h-4 w-4" />
          </button>
        </div>
      </div>
    </aside>
  );
}

function PageTitleBar({
  compact,
  onToggleCompact,
}: {
  compact: boolean;
  onToggleCompact: () => void;
}) {
  return (
    <div className="flex items-start justify-between gap-3 sm:items-end">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(0,122,120,0.10)]" />
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            Teacher portal
          </p>
        </div>
        <h1 className="mt-2 text-[26px] font-bold tracking-tight text-slate-900 sm:text-3xl">
          My Timetable
        </h1>
        <p className="mt-1.5 max-w-xl text-sm leading-relaxed text-slate-500">
          A clear view of your week, period by period.
        </p>
      </div>

      <div className="flex shrink-0 items-center gap-2 print:hidden">
        <button
          type="button"
          onClick={() => window.print()}
          className="inline-flex h-10 w-10 items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-xs font-semibold text-slate-600 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/30 sm:w-auto sm:px-3"
          aria-label="Print timetable"
        >
          <Printer className="h-4 w-4" />
          <span className="hidden sm:inline">Print</span>
        </button>
        <button
          type="button"
          onClick={onToggleCompact}
          className="inline-flex h-10 w-10 items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-xs font-semibold text-slate-600 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/30 sm:w-auto sm:px-3"
          aria-label={compact ? "Switch to expanded view" : "Switch to compact view"}
        >
          {compact ? (
            <LayoutGrid className="h-4 w-4" />
          ) : (
            <LayoutList className="h-4 w-4" />
          )}
          <span className="hidden sm:inline">
            {compact ? "Expanded" : "Compact"}
          </span>
        </button>
      </div>
    </div>
  );
}

function NotificationBanner() {
  return (
    <section
      aria-label="Timetable notices"
      className="mt-5 rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-3.5 sm:p-4 print:border-slate-200 print:bg-white"
    >
      <div className="flex gap-3">
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
          <Info className="h-4 w-4" />
        </span>

        <div className="min-w-0 flex-1">
          <div className="flex flex-col gap-1.5 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between sm:gap-3">
            <p className="text-[13px] font-semibold text-slate-800">
              Timetable updated{" "}
              <span className="font-normal text-slate-500">
                [Timestamp note]
              </span>
            </p>
            <span className="inline-flex w-fit items-center gap-1.5 rounded-full bg-white/70 px-2.5 py-1 text-[10px] font-semibold text-brand-700 ring-1 ring-white">
              <ShieldCheck className="h-3.5 w-3.5" />
              [Verification message]
            </span>
          </div>

          <div className="mt-2.5 flex flex-col gap-2.5 rounded-xl border border-white/80 bg-white/70 p-2.5 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-center gap-2.5">
              <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-[#F0EBF9] text-[#6D4FA3]">
                <CalendarDays className="h-4 w-4" />
              </span>
              <div className="min-w-0">
                <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
                  Event reminder
                </p>
                <p className="truncate text-xs font-semibold text-slate-800">
                  [Event Title]
                </p>
              </div>
            </div>
            <span className="w-fit shrink-0 rounded-full bg-white px-2.5 py-1 text-[10px] font-semibold text-[#6D4FA3] shadow-sm ring-1 ring-[#E4DBF5]">
              [Date]
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}

function ScheduleCard({
  entry,
  compact,
}: {
  entry: TimetableEntry;
  compact: boolean;
}) {
  const live = entry.kind === "live";
  const details =
    entry.kind === "planning"
      ? [
          { icon: DoorOpen, label: entry.location ?? "[Desk / Room]" },
          { icon: ListChecks, label: entry.detail ?? "[Task checklist item]" },
        ]
      : entry.kind === "support"
        ? [
            { icon: Library, label: entry.location ?? "[Library]" },
            { icon: Users, label: entry.detail ?? "[Group focus]" },
          ]
        : [
            { icon: MapPin, label: entry.location ?? "[Location]" },
            { icon: Users, label: entry.detail ?? "[Student Count]" },
          ];

  return (
    <article
      className={cn(
        "group flex overflow-hidden rounded-2xl border shadow-card transition-all duration-200 hover:-translate-y-px hover:shadow-card-hover print:break-inside-avoid print:shadow-none",
        CARD_STYLES[entry.kind],
      )}
    >
      <div
        className={cn(
          "flex w-[92px] shrink-0 flex-col justify-between px-3 sm:w-[128px] sm:px-4",
          compact ? "py-3.5 sm:py-4" : "py-4 sm:py-5",
        )}
      >
        <div>
          <p className="text-[9px] font-semibold uppercase tracking-[0.14em] text-slate-400">
            Start
          </p>
          <p
            className={cn(
              "mt-1 text-[17px] font-bold leading-none tracking-tight sm:text-xl",
              TIME_STYLES[entry.kind],
            )}
          >
            {entry.start}
          </p>
          <p className="mt-1.5 text-[11px] font-medium text-slate-400">
            to {entry.end}
          </p>
          {live ? (
            <span className="mt-2 inline-flex items-center gap-1 text-[9px] font-bold uppercase tracking-[0.14em] text-[#7C5CBF]">
              <span className="relative flex h-1.5 w-1.5">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#7C5CBF] opacity-70" />
                <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-[#7C5CBF]" />
              </span>
              Live
            </span>
          ) : (
            <p className="mt-2 text-[9px] font-semibold uppercase tracking-[0.14em] text-slate-400">
              {entry.duration}
            </p>
          )}
        </div>
        <p className="hidden text-[9px] font-medium text-slate-400 sm:block">
          Period
        </p>
      </div>

      <div
        className={cn("my-3 w-px shrink-0 sm:my-4", DIVIDER_STYLES[entry.kind])}
      />

      <div
        className={cn(
          "min-w-0 flex-1 px-3 py-3.5 sm:px-5 sm:py-4",
          compact && "sm:py-3.5",
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="text-[9px] font-semibold uppercase tracking-[0.14em] text-slate-400">
              {entry.kind === "planning" ? "Free period" : "Lesson"}
            </p>
            <h3 className="mt-1 truncate text-sm font-bold tracking-tight text-slate-900 sm:text-[15px]">
              {entry.title}
            </h3>
          </div>
          <span
            className={cn(
              "inline-flex shrink-0 items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-semibold ring-1",
              BADGE_STYLES[entry.kind],
            )}
          >
            {live && <span className="h-1.5 w-1.5 rounded-full bg-white" />}
            {entry.badge}
          </span>
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1.5">
          {details.map((detail) => {
            const Icon = detail.icon;
            return (
              <span
                key={detail.label}
                className="inline-flex items-center gap-1.5 text-[11px] font-medium text-slate-500"
              >
                <Icon className="h-3.5 w-3.5 text-slate-400" />
                {detail.label}
              </span>
            );
          })}
        </div>

        {entry.tag && !compact && (
          <span className="mt-2.5 inline-flex items-center gap-1 rounded-full bg-white/80 px-2 py-1 text-[9px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
            <Sparkles className="h-3 w-3 text-brand-500" />
            {entry.tag}
          </span>
        )}
      </div>
    </article>
  );
}

function BreakCard({ entry }: { entry: TimetableEntry }) {
  return (
    <article className="flex items-center justify-between gap-3 rounded-2xl border border-[#F0E0B6] bg-[#FFF6E5] px-3.5 py-3.5 shadow-card sm:px-5 sm:py-4 print:break-inside-avoid print:shadow-none">
      <div className="flex min-w-0 items-center gap-3">
        <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-[#FCE9BC] text-[#9A6A10]">
          <Coffee className="h-[18px] w-[18px]" />
        </span>
        <div className="min-w-0">
          <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-[#A47A28]">
            Break
          </p>
          <h3 className="truncate text-sm font-bold text-[#5F4715]">
            {entry.title}
          </h3>
        </div>
      </div>
      <div className="flex shrink-0 flex-col items-end gap-1">
        <p className="text-[11px] font-semibold text-[#6B521C] sm:text-xs">
          {entry.start} – {entry.end}
        </p>
        <span className="rounded-full bg-white/70 px-2 py-0.5 text-[9px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
          {entry.duration}
        </span>
      </div>
    </article>
  );
}

function WeekOverview({
  summary,
  progress,
}: {
  summary: ReturnType<typeof getWeekSummary>;
  progress: number;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-card">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-[9px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            Overview
          </p>
          <h2 className="mt-1 text-base font-bold tracking-tight text-slate-900">
            Week at a glance
          </h2>
        </div>
        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700">
          This week
        </span>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-2">
        <MiniStat label="Teaching" value={summary.teaching} icon={BookOpen} />
        <MiniStat label="Periods" value={summary.periods} icon={CalendarDays} />
        <MiniStat label="Planning" value={summary.planning} icon={ListChecks} />
        <MiniStat label="Support" value={summary.support} icon={Library} />
      </div>

      <div className="mt-5">
        <div className="flex items-center justify-between text-[10px] font-semibold">
          <span className="text-slate-500">Week progress</span>
          <span className="text-brand-700">{progress}%</span>
        </div>
        <div className="mt-2 h-2 overflow-hidden rounded-full bg-slate-100">
          <div
            className="h-full rounded-full bg-brand-500 transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>
    </section>
  );
}

function MiniStat({
  label,
  value,
  icon: Icon,
}: {
  label: string;
  value: number;
  icon: LucideIcon;
}) {
  return (
    <div className="rounded-xl border border-slate-100 bg-slate-50/70 p-3">
      <Icon className="h-3.5 w-3.5 text-brand-500" />
      <p className="mt-2 text-lg font-bold leading-none tracking-tight text-slate-900">
        {value}
      </p>
      <p className="mt-1 text-[10px] font-medium text-slate-400">{label}</p>
    </div>
  );
}

function NextUp({
  entry,
  day,
  isToday,
}: {
  entry: TimetableEntry | undefined;
  day: WeekDay | undefined;
  isToday: boolean;
}) {
  return (
    <section className="overflow-hidden rounded-2xl border border-[#D7C8F2] bg-[#F7F3FD] p-5">
      <div className="flex items-center justify-between gap-2">
        <p className="text-[9px] font-semibold uppercase tracking-[0.18em] text-[#8A6DBC]">
          Next up
        </p>
        <span className="inline-flex items-center gap-1 rounded-full bg-white px-2 py-0.5 text-[9px] font-semibold text-[#6D4FA3] ring-1 ring-[#E4DBF5]">
          <span className="h-1.5 w-1.5 rounded-full bg-[#7C5CBF]" />
          {isToday ? "Now" : "Next"}
        </span>
      </div>
      <h2 className="mt-2 text-base font-bold tracking-tight text-slate-900">
        {entry?.title ?? "[Current Lesson Title]"}
      </h2>
      <p className="mt-1 text-xs text-slate-500">
        {day?.fullLabel} · {entry?.start ?? "[Start Time]"} –{" "}
        {entry?.end ?? "[End Time]"}
      </p>
      <div className="mt-3 flex items-center gap-3 rounded-xl bg-white/75 p-2.5">
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-[#F0EBF9] text-[#6D4FA3]">
          <MapPin className="h-3.5 w-3.5" />
        </span>
        <div className="min-w-0">
          <p className="truncate text-xs font-semibold text-slate-700">
            {entry?.location ?? "[Location]"}
          </p>
          <p className="text-[10px] text-slate-400">
            {entry?.detail ?? "[Student Count]"}
          </p>
        </div>
      </div>
    </section>
  );
}

function UpcomingCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-card">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Upcoming
        </h2>
        <span className="text-[10px] font-semibold text-brand-600">View all</span>
      </div>
      <div className="mt-3 space-y-2">
        <UpcomingRow title="[Event Title]" date="[Date]" />
        <UpcomingRow title="[Deadline Title]" date="[Date]" />
      </div>
    </section>
  );
}

function UpcomingRow({ title, date }: { title: string; date: string }) {
  return (
    <div className="flex items-center gap-2.5 rounded-xl border border-slate-100 bg-slate-50/60 p-2.5">
      <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white text-brand-600 shadow-sm">
        <CalendarDays className="h-3.5 w-3.5" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="truncate text-xs font-semibold text-slate-700">{title}</p>
        <p className="text-[10px] text-slate-400">{date}</p>
      </div>
      <ChevronRight className="h-3.5 w-3.5 shrink-0 text-slate-300" />
    </div>
  );
}

function SyncCallout({
  state,
  onSync,
}: {
  state: SyncState;
  onSync: () => void;
}) {
  return (
    <section className="mt-4 flex flex-col gap-3 rounded-2xl border border-brand-100 bg-[#E8F5F3] p-4 sm:flex-row sm:items-center sm:justify-between print:hidden">
      <div className="flex gap-3">
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-brand-600 shadow-sm">
          <RefreshCw
            className={cn("h-4 w-4", state === "syncing" && "animate-spin")}
          />
        </span>
        <div className="min-w-0">
          <p className="text-xs font-bold text-brand-800">
            {state === "done" ? "Timetable synced" : "Timetable sync"}
          </p>
          <p className="mt-0.5 max-w-xl text-[11px] leading-relaxed text-slate-500">
            [System sync disclaimer] Changes made by the academic office sync
            automatically. Last verified [Timestamp].
          </p>
        </div>
      </div>
      <button
        type="button"
        onClick={onSync}
        disabled={state === "syncing"}
        className="inline-flex h-9 shrink-0 items-center justify-center gap-1.5 rounded-xl bg-white px-3 text-[11px] font-semibold text-brand-700 shadow-sm ring-1 ring-brand-100 transition hover:bg-brand-50 disabled:cursor-wait disabled:opacity-70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/30"
      >
        {state === "done" ? (
          <Check className="h-3.5 w-3.5" />
        ) : (
          <RefreshCw
            className={cn("h-3.5 w-3.5", state === "syncing" && "animate-spin")}
          />
        )}
        {state === "done"
          ? "Synced"
          : state === "syncing"
            ? "Syncing..."
            : "Refresh"}
      </button>
    </section>
  );
}

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
              href={item.active ? "/teacher" : "#"}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "relative flex flex-col items-center gap-1 rounded-xl py-2 text-[9px] font-semibold transition",
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

function MobileDrawer({ onClose }: { onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 lg:hidden print:hidden">
      <button
        type="button"
        onClick={onClose}
        className="absolute inset-0 cursor-default bg-slate-900/40 backdrop-blur-sm animate-fade-in"
        aria-label="Close navigation menu"
      />
      <div className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white p-5 shadow-2xl animate-drawer-in">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <span className="grid h-10 w-10 place-items-center rounded-xl bg-brand-500 text-white shadow-lg shadow-brand-500/20">
              <GraduationCap className="h-5 w-5" />
            </span>
            <div>
              <p className="text-sm font-bold tracking-tight text-slate-900">
                Ikonex <span className="text-brand-600">Academy</span>
              </p>
              <p className="text-[9px] font-semibold uppercase tracking-[0.18em] text-slate-400">
                School portal
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="grid h-9 w-9 place-items-center rounded-xl bg-slate-50 text-slate-500 transition hover:bg-slate-100"
            aria-label="Close navigation menu"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-6 rounded-2xl border border-slate-200/80 bg-slate-50/80 p-3">
          <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            Organization
          </p>
          <div className="mt-1 flex items-center justify-between gap-2">
            <p className="truncate text-xs font-semibold text-slate-700">
              [Organization Name]
            </p>
            <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 text-slate-400" />
          </div>
        </div>

        <nav className="mt-5 space-y-1" aria-label="Mobile primary">
          {SIDEBAR_ITEMS.slice(0, 5).map((item) => {
            const Icon = item.icon;

            return (
              <a
                key={item.label}
                href={item.href}
                aria-current={item.active ? "page" : undefined}
                onClick={onClose}
                className={cn(
                  "flex items-center gap-3 rounded-xl px-3 py-2.5 text-[13px] font-medium transition",
                  item.active
                    ? "bg-brand-500 text-white shadow-lg shadow-brand-500/15"
                    : "text-slate-500 hover:bg-slate-50 hover:text-slate-800",
                )}
              >
                <Icon className="h-[18px] w-[18px]" />
                {item.label}
              </a>
            );
          })}
        </nav>

        <div className="mt-auto rounded-2xl border border-brand-100 bg-brand-50/60 p-4">
          <p className="flex items-center gap-2 text-xs font-bold text-brand-700">
            <Sparkles className="h-4 w-4" />
            Timetable tips
          </p>
          <p className="mt-1.5 text-[11px] leading-relaxed text-slate-500">
            Tap a date to view that day, or use the compact layout for a faster
            scan.
          </p>
        </div>
      </div>
    </div>
  );
}
