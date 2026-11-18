"use client";

import { useEffect, useMemo, useState, type ReactNode } from "react";
import Link from "next/link";
import {
  Activity,
  AlertTriangle,
  Bell,
  Building2,
  Check,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock3,
  CreditCard,
  FileText,
  GraduationCap,
  Link2,
  LogOut,
  Menu,
  RefreshCw,
  Search,
  ShieldCheck,
  Wallet,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import {
  filterUnmatchedTransactions,
  type AdminUnmatchedData,
  type PaymentInspectorData,
} from "./admin-unmatched-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  countKey?: "schools" | "connections" | "unmatchedPayments";
  href?: string;
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Schools", icon: Building2, countKey: "schools", href: "/admin/schools" },
  {
    label: "Connections",
    icon: Link2,
    countKey: "connections",
    href: "/admin/connections",
  },
  { label: "Fee Structures", icon: Wallet },
  {
    label: "Unmatched Payments",
    icon: AlertTriangle,
    countKey: "unmatchedPayments",
    href: "/admin/unmatched-payments",
    active: true,
  },
  { label: "System Audit", icon: ShieldCheck },
];

const METRIC_ICONS = [Clock3, CheckCircle2, CreditCard];

export default function AdminUnmatchedPaymentsDashboard({
  data,
}: {
  data: AdminUnmatchedData | null;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<"unmatched" | "matched" | "all">(
    "unmatched",
  );
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
  const transactions = useMemo(
    () => data?.transactions ?? [],
    [data?.transactions],
  );
  const metrics = data?.metrics ?? [];
  const tabs = useMemo(
    () => data?.tabs ?? { unmatched: null, matched: null, all: null },
    [data?.tabs],
  );
  const [searchQuery, setSearchQuery] = useState("");
  const visibleTransactions = useMemo(
    () => filterUnmatchedTransactions(transactions, activeTab, searchQuery),
    [transactions, activeTab, searchQuery],
  );
  const [selectedId, setSelectedId] = useState(
    visibleTransactions[0]?.id ?? "",
  );
  const selected = useMemo(
    () =>
      visibleTransactions.find((row) => row.id === selectedId) ?? null,
    [selectedId, visibleTransactions],
  );
  const inspector = selectedId
    ? data?.inspectors?.[selectedId] ?? null
    : null;

  useEffect(() => {
    if (visibleTransactions.length === 0) {
      if (selectedId) {
        setSelectedId("");
      }
      return;
    }

    if (!visibleTransactions.some((row) => row.id === selectedId)) {
      setSelectedId(visibleTransactions[0].id);
    }
  }, [visibleTransactions, selectedId]);

  useEffect(() => {
    const tab = tabs[activeTab];

    if (tab !== null && tab > 0) {
      return;
    }

    const fallback = (["unmatched", "matched", "all"] as const).find(
      (candidate) => (tabs[candidate] ?? 0) > 0,
    );

    if (fallback && fallback !== activeTab) {
      setActiveTab(fallback);
    }
  }, [activeTab, tabs]);

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
        href="#unmatched-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to unmatched payments
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
          id="unmatched-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PageHeader header={data?.header ?? null} />
          <PreviewNotice />
          <ViewTabs activeTab={activeTab} tabs={tabs} onChange={setActiveTab} />
          <FilterToolbar
            termLabel={data?.filter?.termLabel ?? null}
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
          />

          <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.55fr)_minmax(360px,1fr)]">
            <div className="min-w-0 space-y-4">
              <IncomingStreamTable
                transactions={visibleTransactions}
                selectedId={selectedId}
                trust={data?.trust ?? null}
                onSelect={setSelectedId}
              />
              <MetricsGrid metrics={metrics} />
              <AuditGuarantee trust={data?.trust ?? null} />
            </div>
            <div className="min-w-0 xl:sticky xl:top-[84px] xl:self-start">
              <PaymentInspector
                key={selected?.id ?? "no-payment"}
                transaction={selected}
                inspector={inspector}
              />
            </div>
          </div>

          <QueueFooter queue={data?.queue ?? null} />
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
  counts: AdminUnmatchedData["counts"];
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
          aria-controls="unmatched-drawer"
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
            Unmatched Payments
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
          <span
            title="Operator profile"
            className="grid h-8 w-8 place-items-center rounded-full bg-brand-500 text-[9px] font-bold text-white ring-2 ring-white"
          >
            {operatorInitials}
          </span>
        </div>
      </div>
    </header>
  );
}

function PageHeader({
  header,
}: {
  header: AdminUnmatchedData["header"] | null;
}) {
  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <p className="text-[8px] font-medium text-slate-400">
          Ledger reference <ChevronRight className="inline h-2.5 w-2.5" />{" "}
          Gateway code
        </p>
        <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900">
          Unmatched payments · {header?.entityName || "No entity selected"}
        </h1>
        <p className="mt-1 max-w-2xl text-[10px] leading-relaxed text-slate-500">
          {header?.description || "No reconciliation description available."}
        </p>
      </div>
      <span
        className={cn(
          "inline-flex w-fit items-center gap-1.5 rounded-full px-3 py-1.5 text-[9px] font-semibold ring-1",
          header?.queueHealth !== null && header?.queueHealth !== undefined
            ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
            : "bg-slate-100 text-slate-500 ring-slate-200",
        )}
      >
        <span
          className={cn(
            "h-1.5 w-1.5 rounded-full",
            header?.queueHealth !== null && header?.queueHealth !== undefined
              ? "bg-brand-500"
              : "bg-slate-400",
          )}
        />
        {header?.queueHealth !== null && header?.queueHealth !== undefined
          ? `Queue health: ${header.queueHealth} exceptions pending`
          : "Queue health unavailable"}
      </span>
    </div>
  );
}

function PreviewNotice() {
  return (
    <section
      role="status"
      className="mt-3 flex gap-2.5 rounded-xl border border-[#E8C97A] bg-[#FEF3D6] p-3"
    >
      <AlertTriangle
        className="mt-0.5 h-3.5 w-3.5 shrink-0 text-[#8A6A25]"
        aria-hidden="true"
      />
      <div className="min-w-0">
        <p className="text-[9px] font-bold text-[#6D4FA3]">Preview only</p>
        <p className="mt-0.5 text-[8px] leading-relaxed text-slate-600">
          Holding, assigning and refreshing transactions are disabled until the
          reconciliation write endpoints are connected.
        </p>
      </div>
    </section>
  );
}

function ViewTabs({
  activeTab,
  tabs,
  onChange,
}: {
  activeTab: "unmatched" | "matched" | "all";
  tabs: AdminUnmatchedData["tabs"];
  onChange: (tab: "unmatched" | "matched" | "all") => void;
}) {
  const items = [
    { id: "unmatched" as const, label: "Unmatched", count: tabs.unmatched },
    { id: "matched" as const, label: "Matched", count: tabs.matched },
    { id: "all" as const, label: "All", count: tabs.all },
  ];

  return (
    <div className="mt-4 rounded-full bg-[#E8F5F3] p-1">
      <div className="grid grid-cols-3 gap-1" role="tablist" aria-label="Payment views">
        {items.map((tab) => {
          const disabled = tab.count === null || tab.count === 0;

          return (
            <button
              key={tab.id}
              type="button"
              role="tab"
              aria-selected={activeTab === tab.id}
              disabled={disabled}
              aria-disabled={disabled}
              title={
                disabled
                  ? tab.count === null
                    ? `${tab.label} count unavailable`
                    : `No ${tab.label.toLowerCase()} payments`
                  : `Show ${tab.label.toLowerCase()} payments`
              }
              onClick={() => onChange(tab.id)}
              className={cn(
                "inline-flex items-center justify-center gap-1.5 rounded-full px-3 py-2 text-[10px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
                activeTab === tab.id
                  ? "bg-white text-brand-700 shadow-sm"
                  : "text-slate-500",
                disabled && "cursor-not-allowed opacity-50",
              )}
            >
              {tab.label}
              <span className="rounded-full bg-[#FEF3D6] px-1.5 py-0.5 text-[7px] text-[#8A6A25]">
                {tab.count === null
                  ? "Unavailable"
                  : tab.count.toLocaleString("en-GB")}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function FilterToolbar({
  termLabel,
  searchQuery,
  onSearchChange,
}: {
  termLabel: string | null;
  searchQuery: string;
  onSearchChange: (value: string) => void;
}) {
  return (
    <section className="mt-3 flex flex-col gap-2 rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card sm:flex-row sm:items-center">
      <label className="relative shrink-0 sm:w-52">
        <span className="sr-only">Academic term</span>
        <select
          disabled={!termLabel}
          className="h-9 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-7 text-[9px] font-semibold text-slate-600 outline-none focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <option>{termLabel || "No active term"}</option>
        </select>
        <ChevronDown
          className="pointer-events-none absolute right-2 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-500"
          aria-hidden="true"
        />
      </label>
      <label className="relative min-w-0 flex-1">
        <span className="sr-only">Search transactions</span>
        <Search
          className="pointer-events-none absolute left-2.5 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-500"
          aria-hidden="true"
        />
        <input
          type="search"
          value={searchQuery}
          onChange={(event) => onSearchChange(event.target.value)}
          placeholder="Search transactions..."
          className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-8 pr-2.5 text-[9px] text-slate-700 outline-none placeholder:text-slate-500 focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20"
        />
      </label>
      <button
        type="button"
        disabled
        aria-disabled="true"
        title="Refreshing transactions is not connected yet."
        className="grid h-9 w-9 shrink-0 cursor-not-allowed place-items-center rounded-lg border border-slate-200/80 bg-white text-slate-500 opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
        aria-label="Refresh transactions unavailable"
      >
        <RefreshCw className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
    </section>
  );
}

function IncomingStreamTable({
  transactions,
  selectedId,
  trust,
  onSelect,
}: {
  transactions: AdminUnmatchedData["transactions"];
  selectedId: string;
  trust: AdminUnmatchedData["trust"];
  onSelect: (id: string) => void;
}) {
  const pollingLabel =
    trust?.livePolling === true
      ? "Live polling active"
      : trust?.livePolling === false
        ? "Live polling inactive"
        : "Polling status unavailable";
  const signatureLabel =
    trust?.signatureVerification === true
      ? "Cryptographic signature verification active"
      : trust?.signatureVerification === false
        ? "Signature verification inactive"
        : "Signature verification status unavailable";

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Incoming Paybill Stream
          </h2>
          <p className="mt-0.5 text-[8px] text-slate-500">{pollingLabel}</p>
        </div>
        <span className="inline-flex w-fit items-center gap-1 rounded-full bg-[#FEF3D6] px-2.5 py-1 text-[8px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
          <AlertTriangle className="h-3 w-3" aria-hidden="true" />
          {transactions.length} requires manual triage
        </span>
      </div>

      {transactions.length > 0 ? (
        <div className="mt-3 overflow-x-auto rounded-xl border border-slate-100">
          <table className="min-w-[650px] w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-slate-100 bg-[#F8F9FA]">
                <TableHeading>Select</TableHeading>
                <TableHeading>Received</TableHeading>
                <TableHeading align="right">Amount</TableHeading>
                <TableHeading>Reference</TableHeading>
                <TableHeading>Payer Note</TableHeading>
                <TableHeading>Phone</TableHeading>
              </tr>
            </thead>
            <tbody>
              {transactions.map((row) => {
                const selected = selectedId === row.id;

                return (
                  <tr
                    key={row.id}
                    className={cn(
                      "border-b border-slate-100 last:border-0",
                      selected ? "bg-[#F1FAF8]" : "hover:bg-slate-50/70",
                    )}
                  >
                    <td className="px-3 py-3">
                      <button
                        type="button"
                        onClick={() => onSelect(row.id)}
                        aria-pressed={selected}
                        aria-label={`Select payment ${row.reference}`}
                        className={cn(
                          "inline-flex h-7 items-center gap-1 rounded-lg border px-2 text-[8px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2",
                          selected
                            ? "border-brand-600 bg-brand-500 text-white"
                            : "border-slate-200/80 bg-white text-slate-600 hover:bg-slate-50",
                        )}
                      >
                        {selected && (
                          <Check className="h-2.5 w-2.5" aria-hidden="true" />
                        )}
                        {selected ? "Selected" : "Select"}
                      </button>
                    </td>
                    <td className="px-3 py-3 text-[9px] font-medium text-slate-600">
                      {row.received}
                    </td>
                    <td className="px-3 py-3 text-right text-[10px] font-bold tabular-nums text-brand-700">
                      {row.amount}
                    </td>
                    <td className="px-3 py-3 text-[9px] font-semibold text-slate-600">
                      {row.reference}
                    </td>
                    <td className="px-3 py-3">
                      <span className="rounded-lg bg-[#F0EBF9] px-2 py-1 text-[8px] font-medium text-[#6D4FA3]">
                        {row.note}
                      </span>
                    </td>
                    <td className="px-3 py-3 text-[9px] text-slate-500">
                      {row.phone}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      ) : (
        <EmptyState
          icon={FileText}
          title="No incoming transactions"
          body="Unmatched payments will appear here when they are received."
          className="mt-3"
        />
      )}

      <div className="mt-2.5 flex flex-col gap-1 text-[8px] text-slate-500 sm:flex-row sm:items-center sm:justify-between">
        <span
          className={cn(
            "inline-flex items-center gap-1",
            trust?.signatureVerification === true
              ? "text-brand-700"
              : trust?.signatureVerification === false
                ? "text-[#8A6A25]"
                : "text-slate-500",
          )}
        >
          {trust?.signatureVerification === false ? (
            <AlertTriangle className="h-3 w-3" aria-hidden="true" />
          ) : (
            <ShieldCheck className="h-3 w-3" aria-hidden="true" />
          )}
          {signatureLabel}
        </span>
        <span>{transactions.length} visible transactions</span>
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

function MetricsGrid({ metrics }: { metrics: AdminUnmatchedData["metrics"] }) {
  if (metrics.length === 0) {
    return (
      <EmptyState
        icon={Activity}
        title="No reconciliation metrics available"
        body="Queue metrics will appear here when transactions are being processed."
      />
    );
  }

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      {metrics.map((metric, index) => {
        const Icon = METRIC_ICONS[index % METRIC_ICONS.length];
        return (
          <article
            key={metric.id}
            className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card"
          >
            <div className="flex items-center justify-between gap-2">
              <p className="text-[8px] font-semibold uppercase tracking-[0.1em] text-slate-400">
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

function AuditGuarantee({ trust }: { trust: AdminUnmatchedData["trust"] }) {
  const assignmentAudit = trust?.assignmentAudit ?? null;
  const label =
    assignmentAudit === true
      ? "Every assignment is timestamped and signed by the operator."
      : assignmentAudit === false
        ? "Assignment audit signing is disabled."
        : "Assignment audit status unavailable.";
  const Icon = assignmentAudit === true ? ShieldCheck : AlertTriangle;

  return (
    <div
      className={cn(
        "flex gap-2.5 rounded-2xl border p-3.5",
        assignmentAudit === true
          ? "border-brand-100 bg-[#E2F0ED]"
          : assignmentAudit === false
            ? "border-[#EBD7A5] bg-[#FEF3D6]"
            : "border-slate-200 bg-slate-50",
      )}
    >
      <Icon
        className={cn(
          "h-4 w-4 shrink-0",
          assignmentAudit === true
            ? "text-brand-600"
            : assignmentAudit === false
              ? "text-[#8A6A25]"
              : "text-slate-500",
        )}
        aria-hidden="true"
      />
      <p
        className={cn(
          "text-[9px] leading-relaxed",
          assignmentAudit === true
            ? "text-brand-800"
            : assignmentAudit === false
              ? "text-[#8A6A25]"
              : "text-slate-600",
        )}
      >
        {label}
      </p>
    </div>
  );
}

function PaymentInspector({
  transaction,
  inspector,
}: {
  transaction: AdminUnmatchedData["transactions"][number] | null;
  inspector: PaymentInspectorData | null;
}) {
  const [reason, setReason] = useState("");

  if (!transaction) {
    return (
      <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
            <CreditCard className="h-4 w-4" aria-hidden="true" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Active Payment Inspector
          </h2>
        </div>
        <EmptyState
          icon={CreditCard}
          title="No payment selected"
          body="Select an incoming transaction to inspect and reconcile it."
          className="mt-4"
        />
      </section>
    );
  }

  const student = inspector?.student ?? null;
  const ledger = inspector?.ledger ?? null;
  const matched = transaction.status === "matched";
  const heading = matched
    ? `Matched payment ${transaction.reference}`
    : `Assign payment ${transaction.reference}`;

  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
            <CreditCard className="h-4 w-4" aria-hidden="true" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Active Payment Inspector
          </h2>
        </div>
        <span
          className={cn(
            "rounded-full px-2 py-0.5 text-[8px] font-semibold ring-1",
            matched
              ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
              : "bg-[#FEF3D6] text-[#8A6A25] ring-[#EBD7A5]",
          )}
        >
          {matched ? "Matched" : "Unmatched"}
        </span>
      </div>

      <div className="mt-4 rounded-xl border border-slate-100 bg-[#F8FAF9] p-3">
        <p className="text-[10px] font-bold text-slate-800">{heading}</p>
        <p className="mt-2 text-2xl font-bold tracking-tight text-brand-700">
          {transaction.amount}
        </p>
        <p className="mt-0.5 text-[8px] text-slate-500">
          From {transaction.phone} · ID: {transaction.reference}
        </p>
      </div>

      <div className="mt-3 rounded-xl bg-[#F0EBF9] p-3">
        <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-[#6D4FA3]">
          Payer reference note
        </p>
        <p className="mt-1 text-[10px] font-semibold text-slate-700">
          {transaction.note}
        </p>
      </div>

      <div className="mt-3">
        <label className="relative block">
          <span className="sr-only">Search student directory</span>
          <Search
            className="pointer-events-none absolute left-2.5 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-500"
            aria-hidden="true"
          />
          <input
            type="search"
            disabled
            aria-disabled="true"
            placeholder="Student search unavailable until connected"
            className="h-9 w-full cursor-not-allowed rounded-lg border border-slate-200/80 bg-slate-100 pl-8 pr-8 text-[9px] text-slate-700 outline-none placeholder:text-slate-500 disabled:opacity-70"
          />
        </label>

        {student ? (
          <div className="mt-2 rounded-xl border border-slate-100 bg-white p-3">
            <div className="flex items-center gap-2">
              <span className="grid h-9 w-9 place-items-center rounded-full bg-[#F0EBF9] text-[9px] font-bold text-[#6D4FA3]">
                {student.initials}
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-[10px] font-bold text-slate-800">
                  {student.name}
                </p>
                <p className="text-[8px] text-brand-600">
                  {student.admission}
                </p>
                <p className="truncate text-[8px] text-slate-500">
                  {student.grade} · {student.boarding}
                </p>
              </div>
              <button
                type="button"
                disabled
                aria-disabled="true"
                title="Removing a student is not connected yet."
                className="grid h-6 w-6 cursor-not-allowed place-items-center rounded-full bg-slate-100 text-slate-500 opacity-60"
                aria-label="Remove selected student unavailable"
              >
                <X className="h-3 w-3" aria-hidden="true" />
              </button>
            </div>
            <div className="mt-2 border-t border-slate-100 pt-2 text-[8px] text-slate-500">
              Guardian: {student.guardianName} · {student.guardianPhone}
            </div>
          </div>
        ) : (
          <EmptyState
            icon={Search}
            title="No directory match selected"
            body="Search and select a student to continue the reconciliation."
            className="mt-2"
          />
        )}
      </div>

      {ledger ? (
        <div className="mt-3 rounded-xl bg-[#E2F0ED] p-3">
          <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-brand-700">
            Live ledger impact
          </p>
          <p className="mt-1 text-[9px] text-slate-600">
            {ledger.projectionNote}
          </p>
          <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-white">
            <div
              className="h-full rounded-full bg-brand-500"
              style={{ width: `${Math.max(0, Math.min(100, ledger.progress))}%` }}
            />
          </div>
          <div className="mt-1.5 flex justify-between text-[7px] text-slate-500">
            <span>Cleared {ledger.cleared}</span>
            <span>Total {ledger.total}</span>
          </div>
        </div>
      ) : (
        <EmptyState
          icon={Activity}
          title="No ledger preview"
          body="A balance projection will appear when a student is selected."
          className="mt-3"
        />
      )}

      <label className="mt-3 block">
        <span className="mb-1 block text-[8px] font-semibold text-slate-500">
          Reason for manual assignment{" "}
          {inspector?.auditReasonRequired && (
            <span className="text-rose-500">*</span>
          )}
        </span>
        <input
          type="text"
          value={reason}
          onChange={(event) => setReason(event.target.value)}
          placeholder="Enter audit reason"
          className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] px-2.5 text-[10px] text-slate-700 outline-none placeholder:text-slate-500 focus:border-brand-600 focus:bg-white focus:ring-4 focus:ring-brand-600/20"
        />
        <span className="mt-1 block text-[7px] text-slate-500">
          {matched
            ? "This payment is already matched; no audit reason is required."
            : "Required for the audit log."}
        </span>
      </label>

      {inspector?.overpaymentWarning && (
        <div className="mt-3 flex gap-2 rounded-xl bg-[#FEF3D6] p-2.5">
          <AlertTriangle
            className="h-3.5 w-3.5 shrink-0 text-[#8A6A25]"
            aria-hidden="true"
          />
          <p className="text-[8px] leading-relaxed text-[#8A6A25]">
            {inspector.overpaymentWarning}
          </p>
        </div>
      )}

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row">
        <button
          type="button"
          disabled
          aria-disabled="true"
          title="Holding payments is not connected yet."
          className="h-9 flex-1 cursor-not-allowed rounded-lg border border-slate-200/80 bg-white text-[9px] font-semibold text-slate-600 opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
        >
          Hold for later
        </button>
        <button
          type="button"
          disabled
          aria-disabled="true"
          title={
            matched
              ? "This payment is already matched."
              : "Assigning payments is not connected yet."
          }
          className="inline-flex h-9 flex-1 cursor-not-allowed items-center justify-center gap-1 rounded-lg bg-brand-500 text-[9px] font-bold text-white opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
        >
          {!matched && <Check className="h-3 w-3" aria-hidden="true" />}
          {matched ? "Already matched" : "Assign payment"}
        </button>
      </div>
    </section>
  );
}

function QueueFooter({ queue }: { queue: AdminUnmatchedData["queue"] }) {
  const available = Boolean(queue?.status?.trim());
  const Icon = available ? CheckCircle2 : Clock3;

  return (
    <div
      className={cn(
        "mt-4 flex items-center justify-between gap-3 rounded-2xl border p-3.5",
        available
          ? "border-brand-100 bg-[#E2F0ED]"
          : "border-slate-200 bg-slate-50",
      )}
    >
      <p
        className={cn(
          "inline-flex items-center gap-2 text-[9px] font-medium",
          available ? "text-brand-800" : "text-slate-600",
        )}
      >
        <Icon
          className={cn(
            "h-3.5 w-3.5",
            available ? "text-brand-600" : "text-slate-500",
          )}
          aria-hidden="true"
        />
        {queue?.status || "Queue status unavailable"}
      </p>
      {queue?.allClear && (
        <span className="rounded-full bg-white px-2.5 py-1 text-[8px] font-semibold text-brand-700">
          All Clear Ready
        </span>
      )}
    </div>
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
  counts: AdminUnmatchedData["counts"];
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
        id="unmatched-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="unmatched-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="unmatched-menu-title"
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
