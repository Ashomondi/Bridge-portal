"use client";

import { useEffect, useState, type ReactNode } from "react";
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Bell,
  Building2,
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Clock3,
  Download,
  ExternalLink,
  FileText,
  Filter,
  GraduationCap,
  Link2,
  LogOut,
  Menu,
  Plus,
  Search,
  ShieldCheck,
  UserRound,
  Wallet,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import type {
  AdminMetric,
  AdminOnboardingData,
  AdminRegistryRow,
  MetricKind,
  OnboardingStep,
} from "./admin-onboarding-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  countKey?: "schools" | "connections" | "unmatchedPayments";
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Schools", icon: Building2, countKey: "schools", active: true },
  { label: "Connections", icon: Link2, countKey: "connections" },
  { label: "Fee Structures", icon: Wallet },
  { label: "Unmatched Payments", icon: AlertTriangle, countKey: "unmatchedPayments" },
  { label: "System Audit", icon: ShieldCheck },
];

const METRIC_ICONS: Record<MetricKind, LucideIcon> = {
  latency: Activity,
  health: ShieldCheck,
  velocity: Clock3,
};

export default function AdminOnboardingDashboard({
  data,
}: {
  data: AdminOnboardingData | null;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const session = data?.session ?? null;
  const operatorName = session?.operatorName?.trim() || "Operator";
  const tier = session?.tier?.trim() || "Staff";
  const initials = session?.initials?.trim() || initialsFromName(operatorName);
  const consoleData = data?.console ?? {
    version: "Console",
    environment: "Internal infrastructure",
    gatewayRef: null,
    gatewayLive: false,
    syncId: null,
    sessionLabel: null,
  };
  const counts = data?.counts ?? {
    schools: null,
    connections: null,
    unmatchedPayments: null,
  };

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
        href="#admin-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to admin dashboard
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
          sessionLabel={consoleData.sessionLabel}
          syncId={consoleData.syncId}
          activeEntity={data?.activeEntity ?? null}
          operatorInitials={initials}
        />

        <main
          id="admin-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PageHeaderCard
            enrolledCount={data?.header.enrolledCount ?? null}
            regionOptions={data?.header.regionOptions ?? []}
          />

          <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.7fr)_minmax(360px,1fr)]">
            <div className="min-w-0 space-y-4">
              <ChecklistCard onboarding={data?.onboarding ?? null} />
              <MetricsGrid metrics={data?.metrics ?? []} />
            </div>
            <div className="min-w-0">
              <ProvisionForm />
            </div>
          </div>

          <RegistryTable
            registry={
              data?.registry ?? {
                rows: [],
                visibleCount: 0,
                totalCount: 0,
                countyCount: 0,
                totalPages: 1,
                currentPage: 1,
              }
            }
          />
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
  counts: AdminOnboardingData["counts"];
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
          aria-controls="admin-drawer"
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
            Schools
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
  sessionLabel,
  syncId,
  activeEntity,
  operatorInitials,
}: {
  environment: string;
  gatewayRef: string | null;
  gatewayLive: boolean;
  sessionLabel: string | null;
  syncId: string | null;
  activeEntity: string | null;
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

      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-4 border-t border-slate-200/60 px-8 py-2">
        <div className="flex items-center gap-1.5 text-[9px] font-medium text-slate-400">
          <span className="text-slate-600">Operations</span>
          <ChevronRight className="h-3 w-3" />
          <span className="text-slate-600">Schools</span>
          <ChevronRight className="h-3 w-3" />
          <span>{activeEntity || "No entity selected"}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="rounded-full bg-white px-2.5 py-1 text-[8px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
            {sessionLabel || "No operator session"}
          </span>
          <span className="text-[8px] font-medium text-slate-400">
            {syncId || "No sync ID"}
          </span>
        </div>
      </div>
    </header>
  );
}

function PageHeaderCard({
  enrolledCount,
  regionOptions,
}: {
  enrolledCount: number | null;
  regionOptions: { value: string; label: string }[];
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-slate-900 sm:text-2xl">
              Schools &amp; Institutional Onboarding
            </h1>
            <span className="rounded-full bg-[#F0EBF9] px-2 py-0.5 text-[8px] font-semibold text-[#6D4FA3] ring-1 ring-[#E4DBF5]">
              {enrolledCount === null
                ? "No enrollment data"
                : `${enrolledCount.toLocaleString("en-GB")} enrolled`}
            </span>
          </div>
          <p className="mt-1 max-w-2xl text-[10px] leading-relaxed text-slate-500">
            Manage school onboarding, payment connections and registry state
            across the platform.
          </p>
        </div>

        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <label className="relative min-w-0 flex-1 sm:w-56">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
            <input
              type="search"
              placeholder="Search schools..."
              className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8F9FA] pl-8 pr-3 text-[10px] text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
            />
          </label>
          <label className="relative shrink-0 sm:w-40">
            <span className="sr-only">Region or category</span>
            <select
              disabled={regionOptions.length === 0}
              className="h-9 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8F9FA] pl-3 pr-8 text-[10px] font-semibold text-slate-600 outline-none focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10 disabled:cursor-not-allowed disabled:opacity-60"
            >
              <option>
                {regionOptions.length > 0
                  ? "All regions"
                  : "No regions available"}
              </option>
              {regionOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
          </label>
          <button
            type="button"
            className="inline-flex h-9 shrink-0 items-center justify-center gap-1.5 rounded-lg bg-brand-500 px-3 text-[10px] font-semibold text-white shadow-sm transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <Plus className="h-3.5 w-3.5" />
            Create school
          </button>
        </div>
      </div>
    </section>
  );
}

function ChecklistCard({
  onboarding,
}: {
  onboarding: AdminOnboardingData["onboarding"];
}) {
  const steps = onboarding?.steps ?? [];
  const completed = steps.filter(
    (step) => step.kind === "completed" || step.kind === "api",
  ).length;
  const percent = steps.length
    ? Math.round((completed / steps.length) * 100)
    : 0;

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-[8px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            {onboarding?.streamLabel || "Onboarding stream"}
          </p>
          <h2 className="mt-1 text-sm font-bold tracking-tight text-slate-900">
            {onboarding?.entityName || "No entity selected"}
          </h2>
          <p className="mt-0.5 text-[9px] text-slate-400">
            {onboarding?.entityMeta || "No onboarding metadata available"}
          </p>
        </div>
        <div className="text-right">
          <p className="text-[9px] font-semibold text-slate-600">
            {steps.length > 0
              ? `${completed} of ${steps.length} steps completed (${percent}%)`
              : "No onboarding steps"}
          </p>
          <div className="mt-2 flex justify-end gap-1">
            {steps.map((step, index) => (
              <span
                key={step.id}
                className={cn(
                  "h-1.5 w-6 rounded-full",
                  index < completed ? "bg-brand-500" : "bg-slate-100",
                )}
              />
            ))}
          </div>
        </div>
      </div>

      {steps.length > 0 ? (
        <div className="mt-4 space-y-2">
          {steps.map((step) => (
            <StepCard key={step.id} step={step} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={CheckCircle2}
          title="No onboarding steps available"
          body="Checklist steps will appear here when an institution is selected."
          className="mt-4"
        />
      )}
    </section>
  );
}

function StepCard({ step }: { step: OnboardingStep }) {
  const completed = step.kind === "completed" || step.kind === "api";
  const api = step.kind === "api";
  const progress = step.kind === "progress";

  return (
    <article
      className={cn(
        "flex items-center gap-3 rounded-xl border p-3",
        completed && "border-brand-100 bg-[#E8F5F3]",
        progress && "border-[#EBD7A5] bg-[#FFF6E5]",
        step.kind === "pending" && "border-[#E4DBF5] bg-[#F0EBF9]",
      )}
    >
      <span
        className={cn(
          "grid h-7 w-7 shrink-0 place-items-center rounded-full bg-white",
          completed && "text-brand-600",
          progress && "text-[#8A6A25]",
          step.kind === "pending" && "text-[#6D4FA3]",
        )}
      >
        {completed && <CheckCircle2 className="h-4 w-4" />}
        {progress && <Clock3 className="h-4 w-4" />}
        {step.kind === "pending" && <AlertTriangle className="h-4 w-4" />}
      </span>

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <p className="text-[10px] font-bold text-slate-800">{step.title}</p>
          <span
            className={cn(
              "rounded-full bg-white px-1.5 py-0.5 text-[7px] font-semibold ring-1",
              completed && "text-brand-700 ring-brand-100",
              progress && "text-[#8A6A25] ring-[#EBD7A5]",
              step.kind === "pending" && "text-[#6D4FA3] ring-[#E4DBF5]",
            )}
          >
            {api
              ? "Verified"
              : completed
                ? "Done"
                : progress
                  ? "In progress"
                  : "Not yet"}
          </span>
        </div>
        <p className="mt-0.5 truncate text-[8px] text-slate-500">
          {step.meta || "No metadata"}
        </p>
        <p className="mt-0.5 text-[7px] font-medium text-slate-400">
          {step.timestamp || "Timestamp unavailable"}
        </p>
      </div>

      <button
        type="button"
        className={cn(
          "inline-flex h-7 shrink-0 items-center gap-1 rounded-lg px-2 text-[8px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
          completed && "bg-white text-brand-700 ring-1 ring-brand-100",
          progress && "bg-brand-500 text-white",
          step.kind === "pending" && "bg-[#6D4FA3] text-white",
        )}
      >
        {step.action}
        {api && <ExternalLink className="h-2.5 w-2.5" />}
        {!api && !completed && <ArrowRight className="h-2.5 w-2.5" />}
      </button>
    </article>
  );
}

function MetricsGrid({ metrics }: { metrics: AdminMetric[] }) {
  if (metrics.length === 0) {
    return (
      <EmptyState
        icon={Activity}
        title="No metrics available"
        body="Gateway and onboarding metrics will appear here when data is available."
      />
    );
  }

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      {metrics.map((metric) => {
        const Icon = METRIC_ICONS[metric.kind];

        return (
          <article
            key={metric.id}
            className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card"
          >
            <div className="flex items-center justify-between gap-2">
              <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-slate-400">
                {metric.label}
              </p>
              <Icon className="h-3.5 w-3.5 text-brand-500" />
            </div>
            <p className="mt-2 text-2xl font-bold tracking-tight text-slate-900">
              {metric.value || "—"}
            </p>
            <p className="mt-1 text-[8px] text-slate-400">
              {metric.note || "No supporting note"}
            </p>
          </article>
        );
      })}
    </div>
  );
}

function ProvisionForm() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
            <Building2 className="h-4 w-4" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Provision School
          </h2>
        </div>
        <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-500">
          Form Drawer
        </span>
      </div>

      <div className="mt-4">
        <div className="rounded-lg bg-[#F8F9FA] px-2.5 py-1.5">
          <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-slate-500">
            School details
          </p>
        </div>
        <div className="mt-2 space-y-2">
          <FormInput label="School name" placeholder="Enter school name" />
          <div className="grid grid-cols-2 gap-2">
            <FormSelect label="County" placeholder="Select county" />
            <FormSelect label="Type" placeholder="Select type" />
          </div>
          <div className="grid grid-cols-2 gap-2">
            <FormInput label="Contact phone" placeholder="Enter phone" />
            <FormInput label="Contact email" placeholder="Enter email" />
          </div>
          <FormInput label="Physical address" placeholder="Enter address" />
        </div>
      </div>

      <div className="mt-4">
        <div className="flex items-center justify-between gap-2 rounded-lg bg-[#F8F9FA] px-2.5 py-1.5">
          <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-slate-500">
            Principal details
          </p>
          <span className="rounded-full bg-[#E2F0ED] px-1.5 py-0.5 text-[7px] font-semibold text-brand-700">
            Requires SMS OTP
          </span>
        </div>
        <div className="mt-2 space-y-2">
          <FormInput label="Full name" placeholder="Enter full name" />
          <label className="block">
            <span className="mb-1 block text-[8px] font-semibold text-slate-500">
              Phone
            </span>
            <div className="flex h-9 items-center rounded-lg border border-slate-200/80 bg-[#F8FAF9] p-0.5">
              <span className="inline-flex h-8 items-center rounded-md bg-white px-2 text-[9px] font-bold text-brand-700 shadow-sm">
                +254
              </span>
              <input
                type="tel"
                placeholder="Enter phone number"
                className="h-8 min-w-0 flex-1 bg-transparent px-2 text-[10px] text-slate-700 outline-none placeholder:text-slate-400"
              />
            </div>
            <span className="mt-1 block text-[7px] text-slate-400">
              An SMS OTP will be sent before the principal account is activated.
            </span>
          </label>
          <FormInput label="Email" placeholder="Enter email" type="email" />
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2 border-t border-slate-100 pt-3">
        <button
          type="button"
          className="h-9 flex-1 rounded-lg border border-slate-200/80 bg-white text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          Save as draft
        </button>
        <button
          type="button"
          className="inline-flex h-9 flex-1 items-center justify-center gap-1 rounded-lg bg-brand-500 text-[9px] font-bold text-white transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <UserRound className="h-3 w-3" />
          Create school &amp; principal
        </button>
      </div>
    </section>
  );
}

function FormInput({
  label,
  placeholder,
  type = "text",
}: {
  label: string;
  placeholder: string;
  type?: string;
}) {
  return (
    <label className="block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-500">
        {label}
      </span>
      <input
        type={type}
        placeholder={placeholder}
        className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] px-2.5 text-[10px] text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
      />
    </label>
  );
}

function FormSelect({
  label,
  placeholder,
}: {
  label: string;
  placeholder: string;
}) {
  return (
    <label className="relative block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-500">
        {label}
      </span>
      <select className="h-9 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-7 text-[10px] text-slate-700 outline-none focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10">
        <option>{placeholder}</option>
      </select>
      <ChevronDown className="pointer-events-none absolute bottom-2.5 right-2 h-3 w-3 text-slate-400" />
    </label>
  );
}

function RegistryTable({
  registry,
}: {
  registry: AdminOnboardingData["registry"];
}) {
  const rows = registry.rows;

  return (
    <section className="mt-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="flex items-center gap-2 text-sm font-bold tracking-tight text-slate-900">
            <FileText className="h-4 w-4 text-brand-600" />
            Master Registry
          </h2>
          <p className="mt-0.5 text-[8px] text-slate-400">
            Live sync · no registry data loaded
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-slate-200/80 bg-white px-2.5 text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <Download className="h-3 w-3" />
            Export CSV
          </button>
          <button
            type="button"
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-slate-200/80 bg-white px-2.5 text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          >
            <Filter className="h-3 w-3" />
            Filters
          </button>
        </div>
      </div>

      <div className="mt-2.5 overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-card">
        {rows.length > 0 ? (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="border-b border-slate-100 bg-[#F8F9FA]">
                    <TableHeading>School Name</TableHeading>
                    <TableHeading>County</TableHeading>
                    <TableHeading>Type</TableHeading>
                    <TableHeading>Principal</TableHeading>
                    <TableHeading>Onboarding Status</TableHeading>
                    <TableHeading>Created Date</TableHeading>
                    <TableHeading align="right">Actions</TableHeading>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => (
                    <tr
                      key={row.id}
                      className="border-b border-slate-100 transition last:border-0 hover:bg-slate-50/70"
                    >
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <span
                            className={cn(
                              "h-1.5 w-1.5 rounded-full",
                              row.status === "completed" && "bg-brand-500",
                              row.status === "live" && "bg-[#4E8A6E]",
                              row.status === "initiated" && "bg-[#B9A3E0]",
                            )}
                          />
                          <span className="text-[10px] font-semibold text-slate-800">
                            {row.school}
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-[9px] text-slate-500">
                        {row.county}
                      </td>
                      <td className="px-4 py-3 text-[9px] text-slate-500">
                        {row.type}
                      </td>
                      <td className="px-4 py-3 text-[9px] text-slate-500">
                        {row.principal}
                      </td>
                      <td className="px-4 py-3">
                        <StatusBadge row={row} />
                      </td>
                      <td className="px-4 py-3 text-[9px] text-slate-500">
                        {row.created}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <button
                          type="button"
                          className="text-[9px] font-semibold text-brand-600 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
                        >
                          Manage
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="divide-y divide-slate-100 md:hidden">
              {rows.map((row) => (
                <div key={row.id} className="p-3.5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span
                        className={cn(
                          "h-1.5 w-1.5 rounded-full",
                          row.status === "completed" && "bg-brand-500",
                          row.status === "live" && "bg-[#4E8A6E]",
                          row.status === "initiated" && "bg-[#B9A3E0]",
                        )}
                      />
                      <p className="text-[10px] font-bold text-slate-800">
                        {row.school}
                      </p>
                    </div>
                    <StatusBadge row={row} />
                  </div>
                  <div className="mt-2 grid grid-cols-2 gap-2 text-[8px]">
                    <span className="text-slate-400">
                      County:{" "}
                      <span className="text-slate-600">{row.county}</span>
                    </span>
                    <span className="text-slate-400">
                      Type: <span className="text-slate-600">{row.type}</span>
                    </span>
                    <span className="text-slate-400">
                      Principal:{" "}
                      <span className="text-slate-600">{row.principal}</span>
                    </span>
                    <span className="text-slate-400">
                      Created:{" "}
                      <span className="text-slate-600">{row.created}</span>
                    </span>
                  </div>
                  <div className="mt-2 text-right">
                    <button
                      type="button"
                      className="text-[9px] font-semibold text-brand-600 hover:underline"
                    >
                      Manage
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </>
        ) : (
          <div className="p-4">
            <EmptyState
              icon={FileText}
              title="No registry records available"
              body="Onboarded schools will appear here when they are registered."
            />
          </div>
        )}

        <div className="flex flex-col gap-3 border-t border-slate-100 bg-slate-50/30 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-[8px] text-slate-500">
            Showing{" "}
            <span className="font-semibold text-slate-700">
              {registry.visibleCount}
            </span>{" "}
            of{" "}
            <span className="font-semibold text-slate-700">
              {registry.totalCount}
            </span>{" "}
            active schools across{" "}
            <span className="font-semibold text-slate-700">
              {registry.countyCount}
            </span>{" "}
            counties
          </p>
          <div className="flex items-center gap-1">
            <button
              type="button"
              className="inline-flex h-7 items-center gap-1 rounded-lg px-2 text-[8px] font-semibold text-slate-500 hover:bg-white disabled:opacity-40"
              disabled={registry.currentPage <= 1}
            >
              <ChevronLeft className="h-3 w-3" />
              Previous
            </button>
            {Array.from(
              { length: Math.max(registry.totalPages, 1) },
              (_, index) => index + 1,
            ).map((page) => (
              <button
                key={page}
                type="button"
                className={cn(
                  "grid h-7 w-7 place-items-center rounded-lg text-[8px] font-semibold",
                  page === registry.currentPage
                    ? "bg-brand-500 text-white"
                    : "text-slate-500 hover:bg-white",
                )}
              >
                {page}
              </button>
            ))}
            <button
              type="button"
              className="inline-flex h-7 items-center gap-1 rounded-lg px-2 text-[8px] font-semibold text-slate-500 hover:bg-white disabled:opacity-40"
              disabled={registry.currentPage >= registry.totalPages}
            >
              Next
              <ChevronRight className="h-3 w-3" />
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}

function StatusBadge({ row }: { row: AdminRegistryRow }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-full px-2 py-0.5 text-[8px] font-semibold ring-1",
        row.status === "completed" &&
          "bg-[#E2F0ED] text-brand-700 ring-brand-100",
        row.status === "live" &&
          "bg-[#E8F5F3] text-[#3E7A5E] ring-[#CBE6DB]",
        row.status === "initiated" &&
          "bg-[#F0EBF9] text-[#6D4FA3] ring-[#E4DBF5]",
      )}
    >
      {row.statusLabel}
      {row.progress ? ` · ${row.progress}` : ""}
    </span>
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
        "whitespace-nowrap px-4 py-2.5 text-[8px] font-semibold uppercase tracking-[0.1em] text-slate-400",
        align === "right" ? "text-right" : "text-left",
      )}
    >
      {children}
    </th>
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
        "rounded-2xl border border-dashed border-slate-200 bg-slate-50/60 px-4 py-8 text-center",
        className,
      )}
    >
      <span className="mx-auto grid h-10 w-10 place-items-center rounded-xl bg-white text-slate-400 shadow-sm">
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
  counts: AdminOnboardingData["counts"];
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
        id="admin-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="admin-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="admin-menu-title"
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
