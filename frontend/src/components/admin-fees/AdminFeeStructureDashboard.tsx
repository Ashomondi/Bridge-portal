"use client";

import { useEffect, useState, type ReactNode } from "react";
import {
  AlertTriangle,
  ArrowRight,
  Bell,
  Building2,
  CalendarDays,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  GraduationCap,
  Info,
  Link2,
  Lock,
  LogOut,
  Menu,
  Plus,
  ShieldCheck,
  Users,
  Wallet,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";

type NavItem = {
  label: string;
  icon: LucideIcon;
  badge?: string;
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Schools", icon: Building2, badge: "[X]" },
  { label: "Connections", icon: Link2, badge: "[X]" },
  { label: "Fee Structures", icon: Wallet, active: true },
  { label: "Unmatched Payments", icon: AlertTriangle, badge: "[X]" },
  { label: "System Audit", icon: ShieldCheck },
];

const GRADES = ["Grade 7", "Grade 8", "Grade 9", "Grade 10", "Grade 11", "Grade 12"];

const FEE_ROWS = [
  { id: "fee-1", title: "[Fee Item]", term1: "[Amount]", term2: "[Amount]", term3: "[Amount]", total: "[Amount]" },
  { id: "fee-2", title: "[Fee Item]", term1: "[Amount]", term2: "[Amount]", term3: "[Amount]", total: "[Amount]" },
  { id: "fee-3", title: "[Fee Item]", term1: "[Amount]", term2: "[Amount]", term3: "[Amount]", total: "[Amount]" },
  { id: "fee-4", title: "[Fee Item]", term1: "[Amount]", term2: "[Amount]", term3: "[Amount]", total: "[Amount]" },
  { id: "fee-5", title: "[Fee Item]", term1: "[Amount]", term2: "[Amount]", term3: "[Amount]", total: "[Amount]" },
];

const DONUT_SEGMENTS = [
  { id: "segment-1", label: "[Fee Item]", percent: "[X]%", color: "#007A78", value: 42 },
  { id: "segment-2", label: "[Fee Item]", percent: "[X]%", color: "#B9A3E0", value: 26 },
  { id: "segment-3", label: "[Fee Item]", percent: "[X]%", color: "#E8C97A", value: 18 },
  { id: "segment-4", label: "[Fee Item]", percent: "[X]%", color: "#CBD5E1", value: 14 },
];

export default function AdminFeeStructureDashboard() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [activeGrade, setActiveGrade] = useState("Grade 8");
  const [activeTab, setActiveTab] = useState<"structure" | "changes">("structure");

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
        href="#fees-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to fee structure
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[280px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar menuOpen={menuOpen} onOpenMenu={() => setMenuOpen(true)} />

      <div className="lg:pl-[280px]">
        <TopBar />

        <main
          id="fees-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PageHeader />
          <ViewTabs activeTab={activeTab} onChange={setActiveTab} />
          <FilterCard
            activeGrade={activeGrade}
            onGradeChange={setActiveGrade}
          />
          <CohortBanner grade={activeGrade} />

          <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.7fr)_minmax(340px,1fr)]">
            <FeeMatrixCard />
            <div className="min-w-0 space-y-4">
              <DonutCard />
              <AdjustmentCallout />
            </div>
          </div>
        </main>
      </div>

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
            <p className="text-[8px] font-semibold uppercase tracking-[0.14em] text-slate-400">
              Console v10.4
            </p>
            <span className="rounded-full bg-[#F0EBF9] px-1.5 py-0.5 text-[7px] font-bold uppercase tracking-wide text-[#6D4FA3] ring-1 ring-[#E4DBF5]">
              Operator
            </span>
          </div>
        </div>
      </div>

      <button
        type="button"
        className="mt-5 flex w-full items-center gap-2.5 rounded-xl border border-slate-200/80 bg-[#F8F9FA] p-3 text-left transition hover:border-brand-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
      >
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white text-brand-600 shadow-sm">
          <Building2 className="h-3.5 w-3.5" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-[8px] font-semibold uppercase tracking-[0.12em] text-slate-400">
            Active context
          </span>
          <span className="block truncate text-[11px] font-semibold text-slate-700">
            [Institution / Entity]
          </span>
        </span>
        <ChevronDown className="h-3.5 w-3.5 shrink-0 text-slate-400" />
      </button>

      <p className="mt-6 px-2 text-[8px] font-semibold uppercase tracking-[0.18em] text-slate-400">
        Operations rail
      </p>

      <nav className="mt-2 space-y-0.5" aria-label="Operations">
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href="#"
              onClick={onNavigate}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "group flex items-center gap-2.5 rounded-xl px-3 py-2.5 text-[11px] font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                item.active
                  ? "bg-[#E8F5F3] font-semibold text-brand-700 ring-1 ring-brand-100"
                  : "text-slate-500 hover:bg-slate-50 hover:text-slate-800",
              )}
            >
              <Icon
                className={cn(
                  "h-4 w-4 shrink-0",
                  item.active
                    ? "text-brand-600"
                    : "text-slate-400 group-hover:text-brand-600",
                )}
              />
              <span className="flex-1">{item.label}</span>
              {item.badge && (
                <span className="rounded-full bg-white px-1.5 py-0.5 text-[8px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
                  {item.badge}
                </span>
              )}
            </a>
          );
        })}
      </nav>

      <div className="mt-auto pt-6">
        <div className="flex items-center gap-2.5 border-t border-slate-100 pt-4">
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
            [AB]
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[11px] font-semibold text-slate-800">
              [Operator Name]
            </p>
            <p className="truncate text-[9px] text-slate-400">[Tier / Role]</p>
          </div>
          <button
            type="button"
            className="grid h-8 w-8 place-items-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Sign out"
          >
            <LogOut className="h-3.5 w-3.5" />
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
          aria-controls="fees-drawer"
          className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open navigation menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="min-w-0 flex-1 text-center">
          <p className="truncate text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            Internal infrastructure
          </p>
          <p className="truncate text-sm font-bold tracking-tight text-slate-900">
            Fee Structures
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

function TopBar() {
  return (
    <header className="sticky top-0 z-30 hidden border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:block print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-4 px-8 py-2.5">
        <div className="flex items-center gap-2.5">
          <span className="rounded-full bg-slate-100 px-2 py-1 text-[8px] font-semibold uppercase tracking-wide text-slate-500">
            Internal infrastructure
          </span>
          <span className="text-[9px] font-medium text-slate-400">
            [Gateway reference]
          </span>
        </div>
        <div className="flex items-center gap-2.5">
          <span className="inline-flex items-center gap-1.5 rounded-full bg-[#E2F0ED] px-2.5 py-1 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
            <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
            Daraja API Live
          </span>
          <button
            type="button"
            className="grid h-8 w-8 place-items-center rounded-full bg-brand-500 text-[9px] font-bold text-white ring-2 ring-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Open user profile"
          >
            [AB]
          </button>
        </div>
      </div>
    </header>
  );
}

function PageHeader() {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <p className="text-[8px] font-medium text-slate-400">
          Module <ChevronRight className="inline h-2.5 w-2.5" /> Entity code{" "}
          <ChevronRight className="inline h-2.5 w-2.5" /> Release tag
        </p>
        <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900">
          Fees · [Entity Name]
        </h1>
        <p className="mt-1 max-w-2xl text-[10px] leading-relaxed text-slate-500">
          [Description body] Annual billing matrices and operational release
          context.
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2.5 py-1 text-[8px] font-semibold text-brand-700 ring-1 ring-brand-100">
          <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
          Sync active
        </span>
        <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[8px] font-semibold text-slate-500">
          Audit v[X]
        </span>
      </div>
    </div>
  );
}

function ViewTabs({
  activeTab,
  onChange,
}: {
  activeTab: "structure" | "changes";
  onChange: (tab: "structure" | "changes") => void;
}) {
  return (
    <div className="mt-4 rounded-full bg-[#E8F5F3] p-1">
      <div className="grid grid-cols-2 gap-1" role="tablist" aria-label="Fee views">
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === "structure"}
          onClick={() => onChange("structure")}
          className={cn(
            "inline-flex items-center justify-center gap-1.5 rounded-full px-3 py-2 text-[10px] font-semibold transition",
            activeTab === "structure"
              ? "bg-white text-brand-700 shadow-sm"
              : "text-slate-400",
          )}
        >
          <Wallet className="h-3.5 w-3.5" />
          Fee Structure
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === "changes"}
          onClick={() => onChange("changes")}
          className={cn(
            "inline-flex items-center justify-center gap-1.5 rounded-full px-3 py-2 text-[10px] font-semibold transition",
            activeTab === "changes"
              ? "bg-white text-brand-700 shadow-sm"
              : "text-slate-400",
          )}
        >
          Fee Changes
          <span className="rounded-full bg-[#F0EBF9] px-1.5 py-0.5 text-[7px] text-[#6D4FA3]">
            [X] logged
          </span>
        </button>
      </div>
    </div>
  );
}

function FilterCard({
  activeGrade,
  onGradeChange,
}: {
  activeGrade: string;
  onGradeChange: (grade: string) => void;
}) {
  return (
    <section className="mt-3 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <FilterSelect label="Academic year" placeholder="[Academic Year]" />
          <FilterSelect label="Curriculum tier" placeholder="[Curriculum Tier]" />
        </div>
        <span className="inline-flex w-fit items-center gap-1 rounded-full bg-[#E2F0ED] px-2.5 py-1 text-[8px] font-semibold text-brand-700 ring-1 ring-brand-100">
          <Lock className="h-3 w-3" />
          Published [Date]
        </span>
      </div>

      <div className="mt-3 overflow-x-auto pb-1">
        <div className="flex min-w-max items-center gap-1.5">
          {GRADES.map((grade) => (
            <button
              key={grade}
              type="button"
              onClick={() => onGradeChange(grade)}
              className={cn(
                "rounded-full px-2.5 py-1 text-[9px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
                grade === activeGrade
                  ? "bg-brand-500 text-white shadow-sm"
                  : "bg-slate-50 text-slate-500 ring-1 ring-slate-200/80 hover:bg-white",
              )}
            >
              {grade}
            </button>
          ))}
          <span className="ml-2 text-[8px] font-medium text-slate-400">
            Cohort [Code]
          </span>
        </div>
      </div>
    </section>
  );
}

function FilterSelect({ label, placeholder }: { label: string; placeholder: string }) {
  return (
    <label className="relative block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-500">
        {label}
      </span>
      <select className="h-8 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-7 text-[9px] text-slate-600 outline-none focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10">
        <option>{placeholder}</option>
      </select>
      <ChevronDown className="pointer-events-none absolute bottom-2 right-2 h-3 w-3 text-slate-400" />
    </label>
  );
}

function CohortBanner({ grade }: { grade: string }) {
  return (
    <div className="mt-3 rounded-2xl bg-[#E2F0ED] p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-2.5">
          <Users className="h-4 w-4 shrink-0 text-brand-600" />
          <div>
            <p className="text-[11px] font-bold text-brand-800">
              {grade} has [X] active enrolled students
            </p>
            <p className="mt-0.5 text-[9px] text-slate-500">
              [Explanation subtext for the active cohort]
            </p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <span className="rounded-full bg-white px-2.5 py-1 text-[9px] font-semibold text-brand-700">
            Ledger base: [Amount]
          </span>
          <button
            type="button"
            className="inline-flex items-center gap-1 text-[9px] font-semibold text-brand-700 hover:underline"
          >
            Apply cohort adjustment
            <ArrowRight className="h-3 w-3" />
          </button>
        </div>
      </div>
    </div>
  );
}

function FeeMatrixCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Core Fee Matrix
          </h2>
          <p className="mt-0.5 text-[9px] text-slate-400">
            [Billing schedule subtext]
          </p>
        </div>
        <div className="flex gap-1.5">
          <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-500">
            Currency: KES
          </span>
          <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2 py-0.5 text-[8px] font-semibold text-brand-700">
            <Lock className="h-2.5 w-2.5" />
            Locked
          </span>
        </div>
      </div>

      <div className="mt-3 overflow-x-auto rounded-xl border border-slate-100">
        <table className="min-w-[560px] w-full border-collapse text-left">
          <thead>
            <tr className="border-b border-slate-100 bg-[#F8F9FA]">
              <TableHeading>Fee Item</TableHeading>
              <TableHeading align="right">Term 1</TableHeading>
              <TableHeading align="right">Term 2</TableHeading>
              <TableHeading align="right">Term 3</TableHeading>
              <TableHeading align="right">Annual Total</TableHeading>
            </tr>
          </thead>
          <tbody>
            {FEE_ROWS.map((row) => (
              <tr key={row.id} className="border-b border-slate-100 last:border-0">
                <td className="px-3 py-3">
                  <div className="flex items-center gap-2">
                    <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
                    <span className="text-[9px] font-semibold text-slate-700">
                      {row.title}
                    </span>
                  </div>
                </td>
                <td className="px-3 py-3 text-right text-[9px] tabular-nums text-slate-500">
                  {row.term1}
                </td>
                <td className="px-3 py-3 text-right text-[9px] tabular-nums text-slate-500">
                  {row.term2}
                </td>
                <td className="px-3 py-3 text-right text-[9px] tabular-nums text-slate-500">
                  {row.term3}
                </td>
                <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-slate-800">
                  {row.total}
                </td>
              </tr>
            ))}
            <tr className="bg-[#E2F0ED]">
              <td className="px-3 py-3 text-[9px] font-bold text-brand-800">
                Totals
              </td>
              <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                [Amount]
              </td>
              <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                [Amount]
              </td>
              <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                [Amount]
              </td>
              <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                [Amount]
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <CalendarDays className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          Due dates · Term 1 [Date] · Term 2 [Date] · Term 3 [Date]
        </p>
      </div>

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="inline-flex items-center gap-1 text-[8px] text-slate-400">
          <CheckCircle2 className="h-3 w-3 text-brand-600" />
          [Verification note]
        </p>
        <div className="flex gap-2">
          <button
            type="button"
            className="inline-flex h-8 items-center gap-1 rounded-lg border border-slate-200/80 bg-white px-2.5 text-[8px] font-semibold text-slate-600 transition hover:bg-slate-50"
          >
            <Plus className="h-3 w-3" />
            Add fee item
          </button>
          <button
            type="button"
            className="inline-flex h-8 items-center gap-1 rounded-lg bg-brand-500 px-2.5 text-[8px] font-bold text-white transition hover:bg-brand-600"
          >
            Publish structure
          </button>
        </div>
      </div>
    </section>
  );
}

function TableHeading({
  children,
  align = "left",
}: {
  children: ReactNode;
  align?: "left" | "right";
}) {
  return (
    <th
      className={cn(
        "whitespace-nowrap px-3 py-2.5 text-[8px] font-semibold uppercase tracking-[0.1em] text-slate-400",
        align === "right" ? "text-right" : "text-left",
      )}
    >
      {children}
    </th>
  );
}

function DonutCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Annual Distribution
        </h2>
        <span className="rounded-full bg-[#F0EBF9] px-2 py-0.5 text-[8px] font-semibold text-[#6D4FA3]">
          [Grade]
        </span>
      </div>

      <div className="mt-3 flex justify-center">
        <svg viewBox="0 0 120 120" className="h-40 w-40" role="img" aria-label="Annual fee distribution">
          <g transform="rotate(-90 60 60)">
            {DONUT_SEGMENTS.reduce<{ nodes: ReactNode[]; offset: number }>(
              (acc, segment, index) => {
                const circumference = 2 * Math.PI * 42;
                const length = (segment.value / 100) * circumference;
                acc.nodes.push(
                  <circle
                    key={segment.id}
                    cx="60"
                    cy="60"
                    r="42"
                    fill="none"
                    stroke={segment.color}
                    strokeWidth="14"
                    strokeDasharray={`${length} ${circumference}`}
                    strokeDashoffset={-acc.offset}
                  />,
                );
                acc.offset += length;
                return acc;
              },
              { nodes: [], offset: 0 },
            ).nodes}
          </g>
          <text x="60" y="56" textAnchor="middle" className="fill-slate-900 text-[11px] font-bold">
            [Total]
          </text>
          <text x="60" y="68" textAnchor="middle" className="fill-slate-400 text-[6px]">
            Annual
          </text>
        </svg>
      </div>

      <div className="mt-2 space-y-1.5">
        {DONUT_SEGMENTS.map((segment) => (
          <div key={segment.id} className="flex items-center gap-2">
            <span
              className="h-2 w-2 rounded-full"
              style={{ backgroundColor: segment.color }}
            />
            <span className="flex-1 text-[9px] text-slate-500">
              {segment.label}
            </span>
            <span className="text-[9px] font-semibold text-slate-700">
              {segment.percent}
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}

function AdjustmentCallout() {
  return (
    <section className="rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4">
      <div className="flex gap-2.5">
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white text-[#6D4FA3] shadow-sm">
          <Info className="h-3.5 w-3.5" />
        </span>
        <div>
          <p className="text-[11px] font-bold text-slate-800">
            Need a mid-term adjustment?
          </p>
          <p className="mt-1 text-[9px] leading-relaxed text-slate-500">
            [Policy note] Post-dispatch modifications require an audit reason and
            may create a new version.
          </p>
          <button
            type="button"
            className="mt-2 inline-flex items-center gap-1 text-[9px] font-semibold text-[#6D4FA3] hover:underline"
          >
            Switch to Changes Tab
            <ArrowRight className="h-3 w-3" />
          </button>
        </div>
      </div>
    </section>
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
        id="fees-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="fees-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="fees-menu-title"
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
