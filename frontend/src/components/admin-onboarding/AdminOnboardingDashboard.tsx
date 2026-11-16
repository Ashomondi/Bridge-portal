"use client";

import { useEffect, useMemo, useState, type ReactNode } from "react";
import Link from "next/link";
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
  href?: string;
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  {
    label: "Schools",
    icon: Building2,
    countKey: "schools",
    href: "/admin/schools",
    active: true,
  },
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
  const registry = data?.registry ?? null;
  const [searchQuery, setSearchQuery] = useState("");
  const [regionFilter, setRegionFilter] = useState("");
  const regionOptions = useMemo(() => {
    const provided = data?.header.regionOptions ?? [];

    if (provided.length > 0) {
      return provided;
    }

    const counties = Array.from(
      new Set(
        (registry?.rows ?? [])
          .map((row) => row.county.trim())
          .filter(Boolean),
      ),
    ).sort((left, right) => left.localeCompare(right));

    return counties.map((county) => ({ value: county, label: county }));
  }, [data?.header.regionOptions, registry?.rows]);

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
            regionOptions={regionOptions}
            searchQuery={searchQuery}
            regionFilter={regionFilter}
            searchDisabled={registry === null}
            onSearchChange={setSearchQuery}
            onRegionChange={setRegionFilter}
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
            registry={registry}
            searchQuery={searchQuery}
            regionFilter={regionFilter}
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
            Bridge <span className="text-brand-600">Portal</span>
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

      <div className="mt-5 flex w-full items-center gap-2.5 rounded-xl border border-slate-200/80 bg-[#F8F9FA] p-3 text-left">
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white text-brand-600 shadow-sm">
          <Building2 className="h-3.5 w-3.5" aria-hidden="true" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-[8px] font-semibold uppercase tracking-[0.12em] text-slate-500">
            Active context
          </span>
          <span className="block truncate text-[11px] font-semibold text-slate-700">
            {activeEntity || "No entity selected"}
          </span>
        </span>
      </div>

      <p className="mt-6 px-2 text-[8px] font-semibold uppercase tracking-[0.18em] text-slate-400">
        Operations rail
      </p>

      <nav className="mt-2 space-y-0.5" aria-label="Operations">
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;
          const count = item.countKey ? counts[item.countKey] : null;
          const className = cn(
            "group flex items-center gap-2.5 rounded-xl px-3 py-2.5 text-[11px] font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
            item.active
              ? "bg-[#E8F5F3] font-semibold text-brand-700 ring-1 ring-brand-100"
              : "text-slate-500",
            item.href
              ? "hover:bg-slate-50 hover:text-slate-800"
              : "cursor-not-allowed opacity-70",
          );
          const content = (
            <>
              <Icon
                aria-hidden="true"
                className={cn(
                  "h-4 w-4 shrink-0",
                  item.active ? "text-brand-600" : "text-slate-500",
                )}
              />
              <span className="flex-1">{item.label}</span>
              {count !== null && (
                <span className="rounded-full bg-white px-1.5 py-0.5 text-[8px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
                  {count.toLocaleString("en-GB")}
                </span>
              )}
            </>
          );

          if (!item.href) {
            return (
              <div
                key={item.label}
                aria-disabled="true"
                title="This operations route is not connected yet."
                className={className}
              >
                {content}
              </div>
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
            disabled
            aria-disabled="true"
            title="Sign out is not connected yet."
            className="grid h-8 w-8 cursor-not-allowed place-items-center rounded-lg text-slate-500 opacity-60"
            aria-label="Sign out unavailable"
          >
            <LogOut className="h-3.5 w-3.5" aria-hidden="true" />
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
          className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
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
            disabled
            aria-disabled="true"
            title="Notifications are not connected yet."
            className="grid h-10 w-10 cursor-not-allowed place-items-center rounded-xl text-slate-500 opacity-60"
            aria-label="Notifications unavailable"
          >
            <Bell className="h-5 w-5" aria-hidden="true" />
          </button>
          <span
            title="Operator profile"
            className="grid h-9 w-9 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white"
          >
            {operatorInitials}
          </span>
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
          <span
            title="Operator profile"
            className="grid h-8 w-8 place-items-center rounded-full bg-brand-500 text-[9px] font-bold text-white ring-2 ring-white"
          >
            {operatorInitials}
          </span>
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
  searchQuery,
  regionFilter,
  searchDisabled,
  onSearchChange,
  onRegionChange,
}: {
  enrolledCount: number | null;
  regionOptions: { value: string; label: string }[];
  searchQuery: string;
  regionFilter: string;
  searchDisabled: boolean;
  onSearchChange: (value: string) => void;
  onRegionChange: (value: string) => void;
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
            <span className="sr-only">Search schools</span>
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-500"
              aria-hidden="true"
            />
            <input
              type="search"
              value={searchQuery}
              onChange={(event) => onSearchChange(event.target.value)}
              disabled={searchDisabled}
              aria-disabled={searchDisabled || undefined}
              placeholder={
                searchDisabled ? "Search unavailable" : "Search schools..."
              }
              className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8F9FA] pl-8 pr-3 text-[10px] text-slate-700 outline-none transition placeholder:text-slate-500 focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20 disabled:cursor-not-allowed disabled:opacity-60"
            />
          </label>
          <label className="relative shrink-0 sm:w-40">
            <span className="sr-only">Region or category</span>
            <select
              value={regionFilter}
              onChange={(event) => onRegionChange(event.target.value)}
              disabled={regionOptions.length === 0}
              className="h-9 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8F9FA] pl-3 pr-8 text-[10px] font-semibold text-slate-600 outline-none focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20 disabled:cursor-not-allowed disabled:opacity-60"
            >
              <option value="">
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
            <ChevronDown
              className="pointer-events-none absolute right-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-500"
              aria-hidden="true"
            />
          </label>
          <button
            type="button"
            disabled
            aria-disabled="true"
            title="School provisioning is not connected yet."
            className="inline-flex h-9 shrink-0 cursor-not-allowed items-center justify-center gap-1.5 rounded-lg bg-brand-500 px-3 text-[10px] font-semibold text-white opacity-60 shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
          >
            <Plus className="h-3.5 w-3.5" aria-hidden="true" />
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
            {steps.map((step) => (
              <span
                key={step.id}
                className={cn(
                  "h-1.5 w-6 rounded-full",
                  step.kind === "completed" || step.kind === "api"
                    ? "bg-brand-500"
                    : step.kind === "progress"
                      ? "bg-[#D6A84A]"
                      : "bg-[#B9A3E0]",
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
        disabled
        aria-disabled="true"
        title="Checklist actions are not connected yet."
        className={cn(
          "inline-flex h-7 cursor-not-allowed shrink-0 items-center gap-1 rounded-lg px-2 text-[8px] font-semibold opacity-70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
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
            <Building2 className="h-4 w-4" aria-hidden="true" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Provision School
          </h2>
        </div>
        <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-600">
          Form Drawer
        </span>
      </div>

      <div
        role="status"
        className="mt-4 flex gap-2.5 rounded-xl border border-[#E8C97A] bg-[#FEF3D6] p-3"
      >
        <AlertTriangle
          className="mt-0.5 h-3.5 w-3.5 shrink-0 text-[#8A6A25]"
          aria-hidden="true"
        />
        <div className="min-w-0">
          <p className="text-[9px] font-bold text-[#6D4FA3]">
            Provisioning is unavailable
          </p>
          <p className="mt-0.5 text-[8px] leading-relaxed text-slate-600">
            School creation will be enabled once the backend provisioning
            endpoint is connected.
          </p>
        </div>
      </div>

      <div className="mt-4">
        <div className="rounded-lg bg-[#F8F9FA] px-2.5 py-1.5">
          <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-slate-500">
            School details
          </p>
        </div>
        <div className="mt-2 space-y-2">
          <FormInput
            label="School code"
            placeholder="Enter school code"
            maxLength={50}
            disabled
          />
          <FormInput
            label="School name"
            placeholder="Enter school name"
            disabled
          />
          <FormSelect label="County" placeholder="Select county" disabled />
          <div className="grid grid-cols-2 gap-2">
            <FormInput
              label="Contact phone"
              placeholder="Enter phone"
              disabled
            />
            <FormInput
              label="Contact email"
              placeholder="Enter email"
              type="email"
              disabled
            />
          </div>
          <FormInput
            label="Physical address"
            placeholder="Enter address"
            disabled
          />
          <FormInput
            label="Office hours"
            placeholder="e.g. Mon–Fri, 8:00–17:00"
            disabled
          />
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
          <FormInput
            label="Full name"
            placeholder="Enter full name"
            disabled
          />
          <label className="block">
            <span className="mb-1 block text-[8px] font-semibold text-slate-500">
              Phone
            </span>
            <div className="flex h-9 items-center rounded-lg border border-slate-200/80 bg-[#F8FAF9] p-0.5 opacity-70">
              <span className="inline-flex h-8 items-center rounded-md bg-white px-2 text-[9px] font-bold text-brand-700 shadow-sm">
                +254
              </span>
              <input
                type="tel"
                placeholder="Enter phone number"
                disabled
                aria-disabled="true"
                className="h-8 min-w-0 flex-1 cursor-not-allowed bg-transparent px-2 text-[10px] text-slate-700 outline-none placeholder:text-slate-500"
              />
            </div>
            <span className="mt-1 block text-[7px] text-slate-500">
              An SMS OTP will be sent before the principal account is activated.
            </span>
          </label>
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2 border-t border-slate-100 pt-3">
        <button
          type="button"
          disabled
          aria-disabled="true"
          title="Provisioning is not connected yet."
          className="h-9 flex-1 cursor-not-allowed rounded-lg border border-slate-200/80 bg-white text-[9px] font-semibold text-slate-600 opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
        >
          Save as draft
        </button>
        <button
          type="button"
          disabled
          aria-disabled="true"
          title="Provisioning is not connected yet."
          className="inline-flex h-9 flex-1 cursor-not-allowed items-center justify-center gap-1 rounded-lg bg-brand-500 text-[9px] font-bold text-white opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
        >
          <UserRound className="h-3 w-3" aria-hidden="true" />
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
  disabled = false,
  maxLength,
}: {
  label: string;
  placeholder: string;
  type?: string;
  disabled?: boolean;
  maxLength?: number;
}) {
  return (
    <label className="block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-600">
        {label}
      </span>
      <input
        type={type}
        placeholder={placeholder}
        disabled={disabled}
        aria-disabled={disabled || undefined}
        maxLength={maxLength}
        className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] px-2.5 text-[10px] text-slate-700 outline-none transition placeholder:text-slate-500 focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:opacity-70"
      />
    </label>
  );
}

function FormSelect({
  label,
  placeholder,
  disabled = false,
}: {
  label: string;
  placeholder: string;
  disabled?: boolean;
}) {
  return (
    <label className="relative block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-600">
        {label}
      </span>
      <select
        disabled={disabled}
        aria-disabled={disabled || undefined}
        className="h-9 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-7 text-[10px] text-slate-700 outline-none focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:opacity-70"
      >
        <option>{placeholder}</option>
      </select>
      <ChevronDown
        className="pointer-events-none absolute bottom-2.5 right-2 h-3 w-3 text-slate-500"
        aria-hidden="true"
      />
    </label>
  );
}

const REGISTRY_PAGE_SIZE = 10;

function RegistryTable({
  registry,
  searchQuery,
  regionFilter,
}: {
  registry: AdminOnboardingData["registry"] | null;
  searchQuery: string;
  regionFilter: string;
}) {
  const [page, setPage] = useState(1);
  const rows = useMemo(() => registry?.rows ?? [], [registry?.rows]);
  const normalizedQuery = searchQuery.trim().toLowerCase();
  const hasFilters = normalizedQuery.length > 0 || regionFilter.length > 0;
  const filteredRows = useMemo(
    () =>
      rows.filter((row) => {
        const matchesRegion = !regionFilter || row.county === regionFilter;
        const searchable = [
          row.school,
          row.county,
          row.type,
          row.principal,
          row.statusLabel,
        ]
          .join(" ")
          .toLowerCase();

        return matchesRegion && (!normalizedQuery || searchable.includes(normalizedQuery));
      }),
    [rows, normalizedQuery, regionFilter],
  );
  const totalPages = Math.max(
    1,
    Math.ceil(filteredRows.length / REGISTRY_PAGE_SIZE),
  );
  const currentPage = Math.min(page, totalPages);
  const visibleRows = filteredRows.slice(
    (currentPage - 1) * REGISTRY_PAGE_SIZE,
    currentPage * REGISTRY_PAGE_SIZE,
  );
  const loadedCount = hasFilters
    ? filteredRows.length
    : registry?.totalCount ?? 0;
  const loadedCountyCount = hasFilters
    ? new Set(filteredRows.map((row) => row.county)).size
    : registry?.countyCount ?? 0;
  const subtitle =
    registry === null
      ? "Live sync · registry data unavailable"
      : hasFilters
        ? `Filtered registry · ${filteredRows.length} matching ${
            filteredRows.length === 1 ? "school" : "schools"
          }`
        : rows.length === 0
          ? "Live sync · no registry records"
          : `Live sync · ${registry.totalCount.toLocaleString("en-GB")} schools loaded`;

  useEffect(() => {
    setPage(1);
  }, [searchQuery, regionFilter]);

  return (
    <section className="mt-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="flex items-center gap-2 text-sm font-bold tracking-tight text-slate-900">
            <FileText className="h-4 w-4 text-brand-600" aria-hidden="true" />
            Master Registry
          </h2>
          <p className="mt-0.5 text-[8px] text-slate-500">{subtitle}</p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => exportRegistryCsv(filteredRows)}
            disabled={filteredRows.length === 0}
            aria-disabled={filteredRows.length === 0}
            title={
              filteredRows.length === 0
                ? "There are no registry rows to export."
                : "Export the filtered registry as CSV."
            }
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-slate-200/80 bg-white px-2.5 text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <Download className="h-3 w-3" aria-hidden="true" />
            Export CSV
          </button>
          <button
            type="button"
            disabled
            aria-disabled="true"
            title="Advanced registry filters are not connected yet."
            className="inline-flex h-8 cursor-not-allowed items-center gap-1.5 rounded-lg border border-slate-200/80 bg-white px-2.5 text-[9px] font-semibold text-slate-600 opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
          >
            <Filter className="h-3 w-3" aria-hidden="true" />
            Filters
          </button>
        </div>
      </div>

      <div className="mt-2.5 overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-card">
        {registry === null ? (
          <div className="p-4">
            <EmptyState
              icon={FileText}
              title="Registry unavailable"
              body="Registry data could not be loaded. Try again once the onboarding service is connected."
            />
          </div>
        ) : visibleRows.length > 0 ? (
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
                  {visibleRows.map((row) => (
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
                          disabled
                          aria-disabled="true"
                          title="School management is not connected yet."
                          className="cursor-not-allowed text-[9px] font-semibold text-brand-600 opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
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
              {visibleRows.map((row) => (
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
                    <span className="text-slate-500">
                      County:{" "}
                      <span className="text-slate-600">{row.county}</span>
                    </span>
                    <span className="text-slate-500">
                      Type: <span className="text-slate-600">{row.type}</span>
                    </span>
                    <span className="text-slate-500">
                      Principal:{" "}
                      <span className="text-slate-600">{row.principal}</span>
                    </span>
                    <span className="text-slate-500">
                      Created:{" "}
                      <span className="text-slate-600">{row.created}</span>
                    </span>
                  </div>
                  <div className="mt-2 text-right">
                    <button
                      type="button"
                      disabled
                      aria-disabled="true"
                      title="School management is not connected yet."
                      className="cursor-not-allowed text-[9px] font-semibold text-brand-600 opacity-60"
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
              title={
                hasFilters
                  ? "No schools match your search"
                  : "No registry records available"
              }
              body={
                hasFilters
                  ? "Try a different school name, county or region filter."
                  : "Onboarded schools will appear here when they are registered."
              }
            />
          </div>
        )}

        {registry !== null && (
          <div className="flex flex-col gap-3 border-t border-slate-100 bg-slate-50/30 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-[8px] text-slate-500">
              Showing{" "}
              <span className="font-semibold text-slate-700">
                {visibleRows.length}
              </span>{" "}
              of{" "}
              <span className="font-semibold text-slate-700">
                {loadedCount}
              </span>{" "}
              active schools across{" "}
              <span className="font-semibold text-slate-700">
                {loadedCountyCount}
              </span>{" "}
              counties
            </p>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={() => setPage((current) => Math.max(1, current - 1))}
                disabled={currentPage <= 1}
                aria-disabled={currentPage <= 1}
                className="inline-flex h-7 items-center gap-1 rounded-lg px-2 text-[8px] font-semibold text-slate-500 hover:bg-white disabled:cursor-not-allowed disabled:opacity-40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
              >
                <ChevronLeft className="h-3 w-3" aria-hidden="true" />
                Previous
              </button>
              {Array.from({ length: totalPages }, (_, index) => index + 1).map(
                (pageNumber) => (
                  <button
                    key={pageNumber}
                    type="button"
                    onClick={() => setPage(pageNumber)}
                    aria-current={
                      pageNumber === currentPage ? "page" : undefined
                    }
                    className={cn(
                      "grid h-7 w-7 place-items-center rounded-lg text-[8px] font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
                      pageNumber === currentPage
                        ? "bg-brand-500 text-white"
                        : "text-slate-500 hover:bg-white",
                    )}
                  >
                    {pageNumber}
                  </button>
                ),
              )}
              <button
                type="button"
                onClick={() =>
                  setPage((current) => Math.min(totalPages, current + 1))
                }
                disabled={currentPage >= totalPages}
                aria-disabled={currentPage >= totalPages}
                className="inline-flex h-7 items-center gap-1 rounded-lg px-2 text-[8px] font-semibold text-slate-500 hover:bg-white disabled:cursor-not-allowed disabled:opacity-40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
              >
                Next
                <ChevronRight className="h-3 w-3" aria-hidden="true" />
              </button>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}

/** Escapes CSV cells and neutralises spreadsheet formula injection. */
function csvCell(value: string) {
  const text = value ?? "";
  const safe = /^[=+\-@\t\r]/.test(text) ? `'${text}` : text;

  return `"${safe.replace(/"/g, '""')}"`;
}

/** Downloads the provided registry rows as a CSV file. */
function exportRegistryCsv(rows: AdminRegistryRow[]) {
  const header = [
    "School Name",
    "County",
    "Type",
    "Principal",
    "Status",
    "Progress",
    "Created",
  ];
  const lines = [
    header.map(csvCell).join(","),
    ...rows.map((row) =>
      [
        row.school,
        row.county,
        row.type,
        row.principal,
        row.statusLabel,
        row.progress ?? "",
        row.created,
      ]
        .map(csvCell)
        .join(","),
    ),
  ];
  const blob = new Blob([`\uFEFF${lines.join("\r\n")}`], {
    type: "text/csv;charset=utf-8;",
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `school-registry-${new Date()
    .toISOString()
    .slice(0, 10)}.csv`;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
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
            className="grid h-9 w-9 place-items-center rounded-xl bg-slate-50 text-slate-500 transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
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
