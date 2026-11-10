"use client";

import { useEffect, useState } from "react";
import {
  Award,
  BarChart3,
  Bell,
  BookOpen,
  Calculator,
  CalendarDays,
  Check,
  ChevronDown,
  ChevronRight,
  ClipboardCheck,
  ClipboardList,
  Download,
  FileText,
  FlaskConical,
  GraduationCap,
  Home,
  Info,
  Languages,
  LayoutDashboard,
  LogOut,
  Menu,
  MessageSquare,
  MoreHorizontal,
  ShieldCheck,
  TrendingUp,
  UserPlus,
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
  badge?: string;
  dot?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "#" },
  { label: "My timetable", icon: CalendarDays, href: "/teacher" },
  {
    label: "My class",
    icon: UsersRound,
    href: "/teacher/class",
    badge: "Class",
  },
  {
    label: "Class means",
    icon: BarChart3,
    href: "/teacher/class-means",
    active: true,
  },
  { label: "Roll-call", icon: ClipboardCheck, href: "#" },
  { label: "Marks", icon: ClipboardList, href: "#" },
  { label: "Results", icon: Award, href: "#" },
  { label: "Comments", icon: MessageSquare, href: "#" },
  { label: "Add students", icon: UserPlus, href: "#" },
  { label: "Notifications", icon: Bell, href: "#", dot: true },
];

const BOTTOM_NAV = [
  { label: "Overview", icon: LayoutDashboard, href: "/teacher/overview" },
  { label: "Timetable", icon: CalendarDays, href: "/teacher" },
  { label: "Class", icon: UsersRound, href: "/teacher/class" },
  { label: "Means", icon: BarChart3, href: "/teacher/class-means", active: true },
  { label: "More", icon: MoreHorizontal, href: "#" },
];

const SUBJECT_STRANDS = [
  { id: "strand-1", icon: BookOpen, points: "4,26 22,22 40,24 58,16 76,18 94,10 116,12" },
  { id: "strand-2", icon: Calculator, points: "4,22 22,24 40,18 58,20 76,12 94,15 116,8" },
  { id: "strand-3", icon: Languages, points: "4,27 22,20 40,23 58,15 76,19 94,12 116,14" },
  { id: "strand-4", icon: FlaskConical, points: "4,24 22,25 40,19 58,21 76,14 94,17 116,9" },
];

const TABLE_ROWS = Array.from({ length: 5 }, (_, index) => ({
  id: `strand-row-${index + 1}`,
  subject: "[Subject]",
  assessment: "[Assessment]",
  term1: "[Score]",
  term2: "[Score]",
  count: "[Count]",
}));

export default function ClassMeansDashboard() {
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
        href="#class-means-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to class means
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[272px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar
        menuOpen={menuOpen}
        onOpenMenu={() => setMenuOpen(true)}
      />

      <div className="lg:pl-[272px]">
        <DesktopHeader />

        <main
          id="class-means-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <PageHeader />
          <InfoBanner />
          <FilterBar />
          <StrandsSection />
          <DetailedTable />
          <PolicyBanner />
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
              {item.badge && (
                <span className="rounded-full bg-white px-1.5 py-0.5 text-[8px] font-bold text-brand-700 ring-1 ring-brand-100">
                  {item.badge}
                </span>
              )}
              {item.dot && (
                <span className="h-1.5 w-1.5 rounded-full bg-rose-500" />
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
                [Full Name]
              </p>
              <p className="truncate text-[10px] text-slate-400">
                [Role Title]
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                [Subject Assignment]
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
          aria-controls="class-means-drawer"
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
            Class Means
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
          <span className="max-w-[220px] truncate">
            [Academic Term / Year]
          </span>
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
            <ChevronDown className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

function PageHeader() {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(0,122,120,0.10)]" />
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            Academic analytics
          </p>
        </div>
        <h1 className="mt-2 text-[26px] font-bold tracking-tight text-slate-900 sm:text-3xl">
          Class means
        </h1>
      </div>

      <span className="inline-flex w-fit items-center gap-1.5 rounded-full bg-[#E2F0ED] px-3 py-1.5 text-[10px] font-semibold text-brand-700 ring-1 ring-brand-100">
        <Check className="h-3.5 w-3.5" />
        [Framework Standard Name]
      </span>
    </div>
  );
}

function InfoBanner() {
  return (
    <section className="mt-5 flex gap-3 rounded-2xl border border-brand-100 bg-[#E8F5F3] p-4">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-brand-600 shadow-sm">
        <Info className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-xs font-bold text-brand-800">Aggregate view only</p>
        <p className="mt-1 max-w-4xl text-[11px] leading-relaxed text-slate-500">
          [Access restriction explanation] Class-level means are available to
          authorized staff. Individual learner records remain protected.
        </p>
      </div>
    </section>
  );
}

function FilterBar() {
  return (
    <section
      aria-label="Class means filters"
      className="mt-4 rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card"
    >
      <div className="grid grid-cols-1 gap-2.5 md:grid-cols-3">
        <FilterSelect
          label="Target class stream"
          placeholder="[Class Stream]"
          options={["[Stream A]", "[Stream B]", "[Stream C]"]}
        />
        <FilterSelect
          label="Academic period"
          placeholder="[Academic Period]"
          options={["[Term / Period 1]", "[Term / Period 2]", "[Term / Period 3]"]}
          icon={CalendarDays}
        />
        <FilterSelect
          label="Assessment series"
          placeholder="[Assessment Series]"
          options={["[Series A]", "[Series B]", "[Series C]"]}
          icon={FileText}
        />
      </div>
    </section>
  );
}

function FilterSelect({
  label,
  placeholder,
  options,
  icon: Icon,
}: {
  label: string;
  placeholder: string;
  options: string[];
  icon?: LucideIcon;
}) {
  return (
    <label className="relative block">
      <span className="sr-only">{label}</span>
      {Icon && (
        <Icon className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
      )}
      <select
        className={cn(
          "h-11 w-full appearance-none rounded-xl border border-slate-200/80 bg-slate-50/60 pr-8 text-[11px] font-semibold text-slate-600 outline-none transition focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10",
          Icon ? "pl-9" : "pl-3",
        )}
        defaultValue=""
      >
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
    </label>
  );
}

function StrandsSection() {
  return (
    <section className="mt-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Subject strands
          </h2>
          <span className="rounded-full bg-white px-2.5 py-1 text-[10px] font-semibold text-slate-600 shadow-sm ring-1 ring-slate-200/80">
            [X] Core Learning Areas
          </span>
        </div>
        <p className="inline-flex items-center gap-1.5 text-[10px] font-medium text-slate-400">
          <TrendingUp className="h-3.5 w-3.5 text-brand-500" />
          Sorted by [Criteria]
        </p>
      </div>

      <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {SUBJECT_STRANDS.map((strand) => {
          const Icon = strand.icon;

          return (
            <article
              key={strand.id}
              className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card transition hover:-translate-y-px hover:shadow-card-hover motion-reduce:transition-none"
            >
              <div className="flex items-start justify-between gap-3">
                <p className="text-[11px] font-semibold text-slate-600">
                  [Subject]
                </p>
                <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-[#F8F9FA] text-slate-400">
                  <Icon className="h-4 w-4" />
                </span>
              </div>

              <div className="mt-3 flex items-end gap-1">
                <span className="text-3xl font-bold leading-none tracking-tight text-slate-900">
                  [Score]
                </span>
                <span className="pb-0.5 text-[11px] font-medium text-slate-400">
                  /100
                </span>
              </div>

              <div className="mt-4 flex items-center justify-between gap-2 text-[9px] font-medium text-slate-400">
                <span>T1 · T2 · T3 trend</span>
                <span>[Count] students</span>
              </div>

              <Sparkline points={strand.points} />
            </article>
          );
        })}
      </div>
    </section>
  );
}

function Sparkline({ points }: { points: string }) {
  return (
    <svg
      viewBox="0 0 120 32"
      className="mt-2 h-8 w-full"
      role="img"
      aria-label="Trend line placeholder"
    >
      <path
        d="M4 28H116"
        fill="none"
        stroke="#E9ECEF"
        strokeWidth="1"
        strokeLinecap="round"
      />
      <polyline
        points={points}
        fill="none"
        stroke="#007A78"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="116" cy={points.split(" ").at(-1)?.split(",")[1]} r="2.4" fill="#007A78" />
    </svg>
  );
}

function DetailedTable() {
  return (
    <section className="mt-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Detailed Strand Tabulation
        </h2>
        <button
          type="button"
          className="inline-flex h-9 w-fit items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white px-3 text-[10px] font-semibold text-slate-600 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <Download className="h-3.5 w-3.5" />
          Export Data Record
        </button>
      </div>

      <div className="mt-3 overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-card">
        <div className="hidden overflow-x-auto md:block">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-slate-100 bg-[#F8F9FA]">
                <TableHeading>Subject</TableHeading>
                <TableHeading>Assessment</TableHeading>
                <TableHeading align="right">Term 1</TableHeading>
                <TableHeading align="right">Term 2</TableHeading>
                <TableHeading align="right">Students Counted</TableHeading>
              </tr>
            </thead>
            <tbody>
              {TABLE_ROWS.map((row) => (
                <tr
                  key={row.id}
                  className="border-b border-slate-100 transition last:border-0 hover:bg-slate-50/70"
                >
                  <td className="px-4 py-3.5 text-[11px] font-semibold text-slate-800 sm:px-5">
                    {row.subject}
                  </td>
                  <td className="px-4 py-3.5 text-[11px] text-slate-500 sm:px-5">
                    {row.assessment}
                  </td>
                  <td className="px-4 py-3.5 text-right text-[11px] font-semibold tabular-nums text-slate-700 sm:px-5">
                    {row.term1}
                  </td>
                  <td className="px-4 py-3.5 text-right text-[11px] font-semibold tabular-nums text-slate-700 sm:px-5">
                    {row.term2}
                  </td>
                  <td className="px-4 py-3.5 text-right text-[11px] font-semibold tabular-nums text-slate-700 sm:px-5">
                    {row.count}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="divide-y divide-slate-100 md:hidden">
          {TABLE_ROWS.map((row) => (
            <div key={row.id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-[11px] font-bold text-slate-800">
                    {row.subject}
                  </p>
                  <p className="mt-0.5 text-[10px] text-slate-400">
                    {row.assessment}
                  </p>
                </div>
                <span className="rounded-full bg-[#F8F9FA] px-2 py-0.5 text-[9px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
                  {row.count} counted
                </span>
              </div>
              <div className="mt-3 grid grid-cols-2 gap-2">
                <div className="rounded-xl bg-[#F8F9FA] px-3 py-2">
                  <p className="text-[8px] font-semibold uppercase tracking-wide text-slate-400">
                    Term 1
                  </p>
                  <p className="mt-0.5 text-[11px] font-bold tabular-nums text-slate-700">
                    {row.term1}
                  </p>
                </div>
                <div className="rounded-xl bg-[#F8F9FA] px-3 py-2">
                  <p className="text-[8px] font-semibold uppercase tracking-wide text-slate-400">
                    Term 2
                  </p>
                  <p className="mt-0.5 text-[11px] font-bold tabular-nums text-slate-700">
                    {row.term2}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function TableHeading({
  children,
  align = "left",
}: {
  children: React.ReactNode;
  align?: "left" | "right";
}) {
  return (
    <th
      className={cn(
        "whitespace-nowrap px-4 py-2.5 text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-400 sm:px-5",
        align === "right" ? "text-right" : "text-left",
      )}
    >
      {children}
    </th>
  );
}

function PolicyBanner() {
  return (
    <section className="mt-6 flex gap-3 rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
        <ShieldCheck className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-xs font-bold text-slate-800">
          Assessment Policy Adherence
        </p>
        <p className="mt-1 max-w-4xl text-[11px] leading-relaxed text-slate-500">
          [Policy explanation] Scores are aggregated from published assessments
          only. Restricted learner-level records are not exposed in this view.
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
      <div className="grid grid-cols-5 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
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
        id="class-means-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="class-means-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="class-means-menu-title"
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
