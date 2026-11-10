"use client";

import { useEffect, useState } from "react";
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

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
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
  { label: "Attendance", icon: ClipboardCheck, href: "#" },
  { label: "Marks", icon: ClipboardList, href: "#" },
  { label: "Resources", icon: Library, href: "#" },
];

const BOTTOM_NAV = [
  { label: "Overview", icon: LayoutDashboard, active: true },
  { label: "Timetable", icon: CalendarDays },
  { label: "Class", icon: UsersRound },
  { label: "Tasks", icon: CheckCircle2 },
  { label: "More", icon: MoreHorizontal },
];

const TODO_ITEMS = [
  {
    id: "todo-not-marked",
    tone: "gold" as const,
    icon: AlertTriangle,
    title: "[Task Title]",
    meta: "",
    status: "Not yet marked",
    statusTone: "gold" as const,
  },
  {
    id: "todo-pending",
    tone: "lavender" as const,
    icon: BookOpen,
    title: "[Task Title]",
    meta: "[Metadata]",
    status: "[X] pending",
    statusTone: "lavender" as const,
  },
  {
    id: "todo-progress",
    tone: "slate" as const,
    icon: ClipboardList,
    title: "[Task Title]",
    meta: "[Metadata]",
    status: "[X] of [Y] waiting",
    statusTone: "slate" as const,
  },
  {
    id: "todo-done",
    tone: "mint" as const,
    icon: CheckCircle2,
    title: "[Task Title]",
    meta: "[Metadata]",
    status: "[X] of [Y] done",
    statusTone: "mint" as const,
  },
];

const ANNOUNCEMENTS = [
  {
    id: "announcement-1",
    icon: Check,
    tone: "mint" as const,
    title: "[Announcement Title]",
    body: "[Announcement description]",
  },
  {
    id: "announcement-2",
    icon: Pencil,
    tone: "lavender" as const,
    title: "[Announcement Title]",
    body: "[Announcement description]",
  },
  {
    id: "announcement-3",
    icon: CalendarDays,
    tone: "slate" as const,
    title: "[Announcement Title]",
    body: "[Announcement description]",
  },
];

export default function TeacherOverviewDashboard() {
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = "";
    };
  }, [menuOpen]);

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#overview-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to overview
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar
        menuOpen={menuOpen}
        onOpenMenu={() => setMenuOpen(true)}
      />

      <div className="lg:pl-[260px]">
        <DesktopHeader />

        <main
          id="overview-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <Greeting />

          <div className="mt-5">
            <NextLessonCard />
          </div>

          <div className="mt-6 grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(360px,1fr)]">
            <div className="order-1 min-w-0 xl:col-start-1 xl:row-start-1">
              <TodoSection />
            </div>

            <div className="order-2 min-w-0 xl:col-start-2 xl:row-span-2 xl:row-start-1 xl:sticky xl:top-[84px] xl:self-start">
              <TimetableWidget />
            </div>

            <div className="order-3 min-w-0 xl:col-start-1 xl:row-start-2">
              <AnnouncementsCard />
            </div>
          </div>
        </main>
      </div>

      <BottomNav />

      {menuOpen && <MobileDrawer onClose={() => setMenuOpen(false)} />}
    </div>
  );
}

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
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
            [AB]
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-xs font-semibold text-slate-800">
              [Full Name]
            </p>
            <p className="text-[10px] text-slate-400">[Role Title]</p>
            <p className="truncate text-[9px] font-medium text-brand-600">
              [Subject Assignment]
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function MobileAppBar({
  menuOpen,
  onOpenMenu,
}: {
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
          className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open navigation menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="min-w-0 flex-1 text-center">
          <p className="truncate text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            [Organization Name]
          </p>
          <p className="truncate text-sm font-bold tracking-tight text-slate-900">
            Teacher Portal
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
            [Organization Name]
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Overview</span>
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
                [Full Name]
              </p>
              <p className="text-[10px] text-slate-400">[Role Title]</p>
            </div>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

function Greeting() {
  return (
    <div className="flex items-start justify-between gap-4">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(0,122,120,0.10)]" />
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            Teacher overview
          </p>
        </div>
        <h1 className="mt-2 text-[26px] font-bold leading-tight tracking-tight text-slate-900 sm:text-3xl">
          Good morning, [Teacher Name]
        </h1>
        <div className="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-slate-500">
          <span className="inline-flex items-center gap-1.5">
            <CalendarDays className="h-3.5 w-3.5 text-slate-400" />
            [Date]
          </span>
          <span className="text-slate-300">•</span>
          <span>[Academic Term]</span>
          <span className="text-slate-300">•</span>
          <span>[Week]</span>
        </div>
      </div>

      <button
        type="button"
        className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        aria-label="Change overview view"
      >
        <SlidersHorizontal className="h-4 w-4" />
      </button>
    </div>
  );
}

function NextLessonCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-card sm:p-6">
      <div className="flex items-center justify-between gap-3">
        <p className="text-[9px] font-semibold uppercase tracking-[0.2em] text-slate-400">
          Next lesson
        </p>
        <span className="inline-flex items-center gap-1.5 rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700 ring-1 ring-brand-100">
          <span className="relative flex h-1.5 w-1.5">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-400 opacity-70 motion-reduce:animate-none" />
            <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-brand-500" />
          </span>
          in [X] min
        </span>
      </div>

      <h2 className="mt-3 text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">
        [Lesson Title]
      </h2>

      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-[11px] font-medium text-slate-500">
        <span className="inline-flex items-center gap-1.5">
          <UsersRound className="h-3.5 w-3.5 text-slate-400" />
          [Group / Grade]
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Clock3 className="h-3.5 w-3.5 text-slate-400" />
          [Start Time]
        </span>
      </div>

      <div className="mt-4 flex items-center justify-between gap-3 rounded-xl bg-[#E8F5F3] px-3.5 py-3">
        <div className="min-w-0">
          <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-brand-600">
            Unit / topic
          </p>
          <p className="mt-0.5 truncate text-xs font-semibold text-slate-800">
            [Unit / Topic]
          </p>
        </div>
        <a
          href="#"
          className="inline-flex shrink-0 items-center gap-1 rounded-lg px-2 py-1 text-[10px] font-bold text-brand-700 transition hover:bg-white/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          Plan
          <ArrowRight className="h-3.5 w-3.5" />
        </a>
      </div>
    </section>
  );
}

function TodoSection() {
  return (
    <section>
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          To do
        </h2>
        <span className="rounded-full bg-white px-2.5 py-1 text-[10px] font-semibold text-slate-600 shadow-sm ring-1 ring-slate-200/80">
          [X] pending
        </span>
      </div>

      <div className="mt-3 space-y-2.5">
        {TODO_ITEMS.map((item) => (
          <TodoCard key={item.id} item={item} />
        ))}
      </div>
    </section>
  );
}

function TodoCard({ item }: { item: (typeof TODO_ITEMS)[number] }) {
  const Icon = item.icon;

  return (
    <a
      href="#"
      className="group flex items-center gap-3 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card transition hover:-translate-y-px hover:shadow-card-hover focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2 motion-reduce:transition-none"
    >
      <span
        className={cn(
          "grid h-10 w-10 shrink-0 place-items-center rounded-xl",
          item.tone === "gold" && "bg-[#FEF3D6] text-[#9A6A10]",
          item.tone === "lavender" && "bg-[#F0EBF9] text-[#6D4FA3]",
          item.tone === "slate" && "bg-slate-100 text-slate-600",
          item.tone === "mint" && "bg-[#E2F0ED] text-brand-700",
        )}
      >
        <Icon className="h-4 w-4" />
      </span>

      <div className="min-w-0 flex-1">
        <p className="truncate text-[12px] font-bold text-slate-800">
          {item.title}
        </p>
        {item.meta && (
          <p className="mt-0.5 truncate text-[10px] text-slate-400">
            {item.meta}
          </p>
        )}
        <span
          className={cn(
            "mt-1.5 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[9px] font-semibold ring-1",
            item.statusTone === "gold" &&
              "bg-[#FEF3D6] text-[#8A6A25] ring-[#EBD7A5]",
            item.statusTone === "lavender" &&
              "bg-[#F0EBF9] text-[#6D4FA3] ring-[#E4DBF5]",
            item.statusTone === "slate" &&
              "bg-slate-100 text-slate-600 ring-slate-200",
            item.statusTone === "mint" &&
              "bg-[#E2F0ED] text-brand-700 ring-brand-100",
          )}
        >
          {item.statusTone === "gold" && (
            <AlertTriangle className="h-2.5 w-2.5" />
          )}
          {item.statusTone === "mint" && (
            <CheckCircle2 className="h-2.5 w-2.5" />
          )}
          {item.status}
        </span>
      </div>

      <ChevronRight className="h-4 w-4 shrink-0 text-slate-300 transition group-hover:translate-x-0.5 group-hover:text-brand-500" />
    </a>
  );
}

function TimetableWidget() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card sm:p-4">
      <div className="flex items-center justify-between gap-3 px-1 pb-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Today&apos;s timetable
        </h2>
        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700 ring-1 ring-brand-100">
          [Day]
        </span>
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-3 rounded-xl bg-[#E8F5F3] p-3">
          <div className="w-[58px] shrink-0">
            <p className="text-[9px] font-semibold text-slate-500">[Start]</p>
            <p className="text-[10px] font-medium text-slate-400">[End]</p>
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[11px] font-bold text-slate-800">
              [Subject Title]
            </p>
            <p className="truncate text-[10px] text-slate-500">[Location]</p>
          </div>
          <span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-white/80 text-brand-600">
            <Check className="h-3.5 w-3.5" />
          </span>
        </div>

        <div className="flex items-center gap-3 rounded-xl border border-[#E4DBF5] bg-[#F0EBF9] p-3">
          <div className="w-[58px] shrink-0">
            <p className="text-[9px] font-semibold text-[#6D4FA3]">[Start]</p>
            <p className="text-[10px] font-medium text-[#8A6DBC]">[End]</p>
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-1.5">
              <p className="truncate text-[11px] font-bold text-slate-800">
                [Subject Title]
              </p>
              <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-white px-1.5 py-0.5 text-[8px] font-bold text-[#6D4FA3]">
                <span className="h-1 w-1 rounded-full bg-[#7C5CBF]" />
                Now
              </span>
            </div>
            <p className="truncate text-[10px] text-slate-500">[Location]</p>
          </div>
          <ArrowRight className="h-4 w-4 shrink-0 text-[#6D4FA3]" />
        </div>

        <div className="flex items-center justify-between gap-3 rounded-xl bg-slate-50 px-3 py-2.5">
          <span className="inline-flex items-center gap-1.5 text-[10px] font-medium text-slate-500">
            <Coffee className="h-3.5 w-3.5 text-slate-400" />
            [Break Label]
          </span>
          <span className="text-[10px] font-semibold text-slate-400">
            [Time]
          </span>
        </div>

        <div className="flex items-center gap-3 rounded-xl bg-[#E8F5F3]/70 p-3">
          <div className="w-[58px] shrink-0">
            <p className="text-[9px] font-semibold text-slate-500">[Start]</p>
            <p className="text-[10px] font-medium text-slate-400">[End]</p>
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[11px] font-bold text-slate-800">
              [Subject Title]
            </p>
            <p className="truncate text-[10px] text-slate-500">[Location]</p>
          </div>
          <span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-white/80 text-slate-400">
            <Clock3 className="h-3.5 w-3.5" />
          </span>
        </div>
      </div>
    </section>
  );
}

function AnnouncementsCard() {
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

      <div className="mt-3 divide-y divide-slate-100">
        {ANNOUNCEMENTS.map((item) => {
          const Icon = item.icon;

          return (
            <div key={item.id} className="flex gap-3 py-3 first:pt-0 last:pb-0">
              <span
                className={cn(
                  "grid h-8 w-8 shrink-0 place-items-center rounded-full",
                  item.tone === "mint" && "bg-[#E2F0ED] text-brand-700",
                  item.tone === "lavender" &&
                    "bg-[#F0EBF9] text-[#6D4FA3]",
                  item.tone === "slate" && "bg-slate-100 text-slate-500",
                )}
              >
                <Icon className="h-3.5 w-3.5" />
              </span>
              <div className="min-w-0">
                <p className="text-[11px] font-bold text-slate-800">
                  {item.title}
                </p>
                <p className="mt-0.5 text-[10px] leading-relaxed text-slate-500">
                  {item.body}
                </p>
              </div>
            </div>
          );
        })}
      </div>

      <a
        href="#"
        className="mt-3 flex items-center justify-center gap-1.5 border-t border-slate-100 pt-3 text-[10px] font-semibold text-brand-700 transition hover:text-brand-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
      >
        See all notifications
        <ArrowRight className="h-3.5 w-3.5" />
      </a>
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
              href={item.active ? "/teacher/overview" : "#"}
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

function MobileDrawer({ onClose }: { onClose: () => void }) {
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
          <SidebarContent onNavigate={onClose} />
        </div>
      </div>
    </div>
  );
}
