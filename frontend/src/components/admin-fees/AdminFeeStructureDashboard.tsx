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
import type { AdminFeesData } from "./admin-fees-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  countKey?: "schools" | "connections" | "unmatchedPayments";
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Schools", icon: Building2, countKey: "schools" },
  { label: "Connections", icon: Link2, countKey: "connections" },
  { label: "Fee Structures", icon: Wallet, active: true },
  { label: "Unmatched Payments", icon: AlertTriangle, countKey: "unmatchedPayments" },
  { label: "System Audit", icon: ShieldCheck },
];

export default function AdminFeeStructureDashboard({
  data,
}: {
  data: AdminFeesData | null;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<"structure" | "changes">("structure");
  const session = data?.session ?? null;
  const operatorName = session?.operatorName?.trim() || "Operator";
  const tier = session?.tier?.trim() || "Staff";
  const initials = session?.initials?.trim() || initialsFromName(operatorName);
  const consoleData = data?.console ?? {
    version: "Console",
    environment: "Internal infrastructure",
    gatewayRef: null,
    gatewayLive: false,
  };
  const counts = data?.counts ?? {
    schools: null,
    connections: null,
    unmatchedPayments: null,
  };
  const grades = data?.grades ?? [];
  const [activeGrade, setActiveGrade] = useState(
    data?.selectedGrade ?? grades[0] ?? "",
  );

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
    if (grades.length > 0 && !grades.includes(activeGrade)) {
      setActiveGrade(grades[0]);
    }
  }, [activeGrade, grades]);

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#fees-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to fee structure
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[280px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          activeEntity={data?.activeEntity ?? null}
          consoleVersion={consoleData.version}
          counts={counts}
          operatorName={operatorName}
          tier={tier}
          initials={initials}
        />
      </aside>

      <MobileAppBar
        operatorInitials={initials}
        menuOpen={menuOpen}
        onOpenMenu={() => setMenuOpen(true)}
      />

      <div className="lg:pl-[280px]">
        <TopBar
          environment={consoleData.environment}
          gatewayRef={consoleData.gatewayRef}
          gatewayLive={consoleData.gatewayLive}
          operatorInitials={initials}
        />

        <main
          id="fees-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PageHeader header={data?.header ?? null} />
          <ViewTabs
            activeTab={activeTab}
            changesLogged={data?.changesLogged ?? null}
            onChange={setActiveTab}
          />
          <FilterCard
            filters={data?.filters ?? null}
            grades={grades}
            activeGrade={activeGrade}
            onGradeChange={setActiveGrade}
          />
          <CohortBanner grade={activeGrade} cohort={data?.cohort ?? null} />

          <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.7fr)_minmax(340px,1fr)]">
            <FeeMatrixCard matrix={data?.matrix ?? null} />
            <div className="min-w-0 space-y-4">
              <DonutCard distribution={data?.distribution ?? null} />
              <AdjustmentCallout adjustment={data?.adjustment ?? null} />
            </div>
          </div>
        </main>
      </div>

      {menuOpen && (
        <MobileDrawer
          activeEntity={data?.activeEntity ?? null}
          consoleVersion={consoleData.version}
          counts={counts}
          operatorName={operatorName}
          tier={tier}
          initials={initials}
          onClose={() => setMenuOpen(false)}
        />
      )}
    </div>
  );
}

function SidebarContent({
  activeEntity,
  consoleVersion,
  counts,
  operatorName,
  tier,
  initials,
  onNavigate,
}: {
  activeEntity: string | null;
  consoleVersion: string;
  counts: AdminFeesData["counts"];
  operatorName: string;
  tier: string;
  initials: string;
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
            <p className="text-[8px] font-semibold uppercase tracking-[0.14em] text-slate-400">
              {consoleVersion}
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
            {activeEntity || "No entity selected"}
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
          const count = item.countKey ? counts[item.countKey] : null;

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
              {count !== null && (
                <span className="rounded-full bg-white px-1.5 py-0.5 text-[8px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
                  {count.toLocaleString("en-GB")}
                </span>
              )}
            </a>
          );
        })}
      </nav>

      <div className="mt-auto pt-6">
        <div className="flex items-center gap-2.5 border-t border-slate-100 pt-4">
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
            {initials}
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[11px] font-semibold text-slate-800">
              {operatorName}
            </p>
            <p className="truncate text-[9px] text-slate-400">{tier}</p>
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
  operatorInitials,
  menuOpen,
  onOpenMenu,
}: {
  operatorInitials: string;
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
            {operatorInitials}
          </button>
        </div>
      </div>
    </header>
  );
}

function TopBar({
  environment,
  gatewayRef,
  gatewayLive,
  operatorInitials,
}: {
  environment: string;
  gatewayRef: string | null;
  gatewayLive: boolean;
  operatorInitials: string;
}) {
  return (
    <header className="sticky top-0 z-30 hidden border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:block print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-4 px-8 py-2.5">
        <div className="flex items-center gap-2.5">
          <span className="rounded-full bg-slate-100 px-2 py-1 text-[8px] font-semibold uppercase tracking-wide text-slate-500">
            {environment}
          </span>
          <span className="text-[9px] font-medium text-slate-400">
            {gatewayRef || "No gateway reference"}
          </span>
        </div>
        <div className="flex items-center gap-2.5">
          <span
            className={cn(
              "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[9px] font-semibold ring-1",
              gatewayLive
                ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
                : "bg-slate-100 text-slate-500 ring-slate-200",
            )}
          >
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                gatewayLive ? "bg-brand-500" : "bg-slate-400",
              )}
            />
            {gatewayLive ? "Daraja API Live" : "Gateway status unavailable"}
          </span>
          <button
            type="button"
            className="grid h-8 w-8 place-items-center rounded-full bg-brand-500 text-[9px] font-bold text-white ring-2 ring-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Open user profile"
          >
            {operatorInitials}
          </button>
        </div>
      </div>
    </header>
  );
}

function PageHeader({ header }: { header: AdminFeesData["header"] | null }) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <p className="text-[8px] font-medium text-slate-400">
          Module <ChevronRight className="inline h-2.5 w-2.5" /> Entity code{" "}
          <ChevronRight className="inline h-2.5 w-2.5" /> Release tag
        </p>
        <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900">
          Fees · {header?.entityName || "No entity selected"}
        </h1>
        <p className="mt-1 max-w-2xl text-[10px] leading-relaxed text-slate-500">
          {header?.description ||
            "No fee structure description available."}
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <span
          className={cn(
            "inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-[8px] font-semibold ring-1",
            header?.syncLabel
              ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
              : "bg-slate-100 text-slate-500 ring-slate-200",
          )}
        >
          <span
            className={cn(
              "h-1.5 w-1.5 rounded-full",
              header?.syncLabel ? "bg-brand-500" : "bg-slate-400",
            )}
          />
          {header?.syncLabel || "Sync status unavailable"}
        </span>
        <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[8px] font-semibold text-slate-500">
          Audit {header?.auditVersion || "not available"}
        </span>
      </div>
    </div>
  );
}

function ViewTabs({
  activeTab,
  changesLogged,
  onChange,
}: {
  activeTab: "structure" | "changes";
  changesLogged: number | null;
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
            {changesLogged === null ? "—" : `${changesLogged} logged`}
          </span>
        </button>
      </div>
    </div>
  );
}

function FilterCard({
  filters,
  grades,
  activeGrade,
  onGradeChange,
}: {
  filters: AdminFeesData["filters"] | null;
  grades: string[];
  activeGrade: string;
  onGradeChange: (grade: string) => void;
}) {
  return (
    <section className="mt-3 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <FilterSelect
            label="Academic year"
            placeholder="Select academic year"
            options={filters?.academicYears ?? []}
          />
          <FilterSelect
            label="Curriculum tier"
            placeholder="Select curriculum tier"
            options={filters?.curriculumTiers ?? []}
          />
        </div>
        <span
          className={cn(
            "inline-flex w-fit items-center gap-1 rounded-full px-2.5 py-1 text-[8px] font-semibold ring-1",
            filters?.published
              ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
              : "bg-slate-100 text-slate-500 ring-slate-200",
          )}
        >
          <Lock className="h-3 w-3" />
          {filters?.published
            ? `Published ${filters.publishedAt || ""}`.trim()
            : "Not published"}
        </span>
      </div>

      <div className="mt-3 overflow-x-auto pb-1">
        {grades.length > 0 ? (
          <div className="flex min-w-max items-center gap-1.5">
            {grades.map((grade) => (
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
          </div>
        ) : (
          <p className="text-[9px] text-slate-400">
            No grade cohorts available.
          </p>
        )}
      </div>
    </section>
  );
}

function FilterSelect({
  label,
  placeholder,
  options,
}: {
  label: string;
  placeholder: string;
  options: { value: string; label: string }[];
}) {
  return (
    <label className="relative block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-500">
        {label}
      </span>
      <select
        disabled={options.length === 0}
        className="h-8 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-7 text-[9px] text-slate-600 outline-none focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10 disabled:cursor-not-allowed disabled:opacity-60"
      >
        <option>
          {options.length > 0 ? placeholder : "No options available"}
        </option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute bottom-2 right-2 h-3 w-3 text-slate-400" />
    </label>
  );
}

function CohortBanner({
  grade,
  cohort,
}: {
  grade: string;
  cohort: AdminFeesData["cohort"];
}) {
  return (
    <div className="mt-3 rounded-2xl bg-[#E2F0ED] p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-2.5">
          <Users className="h-4 w-4 shrink-0 text-brand-600" />
          <div>
            <p className="text-[11px] font-bold text-brand-800">
              {cohort
                ? `${cohort.gradeName} has ${
                    cohort.studentCount === null
                      ? "no"
                      : cohort.studentCount.toLocaleString("en-GB")
                  } active enrolled students`
                : `${grade || "No grade"} cohort summary unavailable`}
            </p>
            <p className="mt-0.5 text-[9px] text-slate-500">
              {cohort?.description || "No cohort metadata available."}
            </p>
          </div>
        </div>
        {cohort && (
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-full bg-white px-2.5 py-1 text-[9px] font-semibold text-brand-700">
              Ledger base: {cohort.ledgerBase || "Not available"}
            </span>
            <button
              type="button"
              className="inline-flex items-center gap-1 text-[9px] font-semibold text-brand-700 hover:underline"
            >
              Apply cohort adjustment
              <ArrowRight className="h-3 w-3" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function FeeMatrixCard({
  matrix,
}: {
  matrix: AdminFeesData["matrix"] | null;
}) {
  const rows = matrix?.rows ?? [];
  const dueDates = matrix?.dueDates ?? [];

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Core Fee Matrix
          </h2>
          <p className="mt-0.5 text-[9px] text-slate-400">
            Billing schedule and structure
          </p>
        </div>
        <div className="flex gap-1.5">
          <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-500">
            Currency: {matrix?.currency || "—"}
          </span>
          <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2 py-0.5 text-[8px] font-semibold text-brand-700">
            <Lock className="h-2.5 w-2.5" />
            Locked
          </span>
        </div>
      </div>

      {rows.length > 0 ? (
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
              {rows.map((row) => (
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
              {matrix?.totals && (
                <tr className="bg-[#E2F0ED]">
                  <td className="px-3 py-3 text-[9px] font-bold text-brand-800">
                    Totals
                  </td>
                  <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                    {matrix.totals.term1}
                  </td>
                  <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                    {matrix.totals.term2}
                  </td>
                  <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                    {matrix.totals.term3}
                  </td>
                  <td className="px-3 py-3 text-right text-[9px] font-bold tabular-nums text-brand-800">
                    {matrix.totals.total}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      ) : (
        <EmptyState
          icon={Wallet}
          title="No fee items available"
          body="The fee matrix will appear here when a structure is selected."
          className="mt-3"
        />
      )}

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <CalendarDays className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          {dueDates.length > 0
            ? dueDates.map((item) => `${item.term}: ${item.date}`).join(" · ")
            : "No term due dates available"}
        </p>
      </div>

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="inline-flex items-center gap-1 text-[8px] text-slate-400">
          <CheckCircle2 className="h-3 w-3 text-brand-600" />
          {matrix?.verificationNote || "Verification note unavailable"}
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

function DonutCard({
  distribution,
}: {
  distribution: AdminFeesData["distribution"] | null;
}) {
  const segments = distribution?.segments ?? [];

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Annual Distribution
        </h2>
        <span className="rounded-full bg-[#F0EBF9] px-2 py-0.5 text-[8px] font-semibold text-[#6D4FA3]">
          {distribution?.gradeTag || "No grade"}
        </span>
      </div>

      {segments.length > 0 ? (
        <>
          <div className="mt-3 flex justify-center">
            <svg viewBox="0 0 120 120" className="h-40 w-40" role="img" aria-label="Annual fee distribution">
              <g transform="rotate(-90 60 60)">
                {segments.reduce<{ nodes: ReactNode[]; offset: number }>(
                  (acc, segment) => {
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
              <text
                x="60"
                y="56"
                textAnchor="middle"
                className="fill-slate-900 text-[11px] font-bold"
              >
                {distribution?.total || "—"}
              </text>
              <text x="60" y="68" textAnchor="middle" className="fill-slate-400 text-[6px]">
                Annual
              </text>
            </svg>
          </div>

          <div className="mt-2 space-y-1.5">
            {segments.map((segment) => (
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
        </>
      ) : (
        <EmptyState
          icon={Info}
          title="No distribution data available"
          body="The annual fee distribution will appear here when the matrix is loaded."
          className="mt-3"
        />
      )}
    </section>
  );
}

function AdjustmentCallout({
  adjustment,
}: {
  adjustment: AdminFeesData["adjustment"] | null;
}) {
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
            {adjustment?.policyNote ||
              "No adjustment policy is available for this fee structure."}
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

function MobileDrawer({
  activeEntity,
  consoleVersion,
  counts,
  operatorName,
  tier,
  initials,
  onClose,
}: {
  activeEntity: string | null;
  consoleVersion: string;
  counts: AdminFeesData["counts"];
  operatorName: string;
  tier: string;
  initials: string;
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
          <SidebarContent
            activeEntity={activeEntity}
            consoleVersion={consoleVersion}
            counts={counts}
            operatorName={operatorName}
            tier={tier}
            initials={initials}
            onNavigate={onClose}
          />
        </div>
      </div>
    </div>
  );
}

function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "OP";
}
