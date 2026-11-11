"use client";

import { useEffect, useMemo, useState, type ReactNode, type SVGProps } from "react";
import {
  AlertTriangle,
  Bell,
  Building2,
  Check,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock3,
  CreditCard,
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

type NavItem = {
  label: string;
  icon: LucideIcon;
  badge?: string;
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Schools", icon: Building2, badge: "[X]" },
  { label: "Connections", icon: Link2, badge: "[X]" },
  { label: "Fee Structures", icon: Wallet },
  { label: "Unmatched Payments", icon: AlertTriangle, badge: "[X]", active: true },
  { label: "System Audit", icon: ShieldCheck },
];

const TRANSACTIONS = [
  {
    id: "transaction-1",
    received: "[Timestamp]",
    amount: "[Amount]",
    reference: "[Reference]",
    note: "[Payer Note]",
    phone: "[Phone]",
  },
  {
    id: "transaction-2",
    received: "[Timestamp]",
    amount: "[Amount]",
    reference: "[Reference]",
    note: "[Payer Note]",
    phone: "[Phone]",
  },
  {
    id: "transaction-3",
    received: "[Timestamp]",
    amount: "[Amount]",
    reference: "[Reference]",
    note: "[Payer Note]",
    phone: "[Phone]",
  },
  {
    id: "transaction-4",
    received: "[Timestamp]",
    amount: "[Amount]",
    reference: "[Reference]",
    note: "[Payer Note]",
    phone: "[Phone]",
  },
  {
    id: "transaction-5",
    received: "[Timestamp]",
    amount: "[Amount]",
    reference: "[Reference]",
    note: "[Payer Note]",
    phone: "[Phone]",
  },
];

const METRICS = [
  { id: "pending", label: "Pending allocation", value: "[X]", icon: Clock3 },
  { id: "matched", label: "Matched this term", value: "[Y]", icon: CheckCircle2 },
  { id: "triage", label: "Avg triage time", value: "[Z]", icon: ActivityIcon },
];

function ActivityIcon(props: SVGProps<SVGSVGElement>) {
  return <CreditCard {...props} />;
}

export default function AdminUnmatchedPaymentsDashboard() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<"unmatched" | "matched" | "all">(
    "unmatched",
  );
  const [selectedId, setSelectedId] = useState(TRANSACTIONS[0].id);
  const selected = useMemo(
    () => TRANSACTIONS.find((row) => row.id === selectedId) ?? TRANSACTIONS[0],
    [selectedId],
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

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#unmatched-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to unmatched payments
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[280px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar menuOpen={menuOpen} onOpenMenu={() => setMenuOpen(true)} />

      <div className="lg:pl-[280px]">
        <TopBar />

        <main
          id="unmatched-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PageHeader />
          <ViewTabs activeTab={activeTab} onChange={setActiveTab} />
          <FilterToolbar />

          <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.55fr)_minmax(360px,1fr)]">
            <div className="min-w-0 space-y-4">
              <IncomingStreamTable
                selectedId={selectedId}
                onSelect={setSelectedId}
              />
              <MetricsGrid />
              <AuditGuarantee />
            </div>
            <div className="min-w-0 xl:sticky xl:top-[84px] xl:self-start">
              <PaymentInspector transaction={selected} />
            </div>
          </div>

          <QueueFooter />
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
          aria-controls="unmatched-drawer"
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
            Unmatched Payments
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
          Ledger reference <ChevronRight className="inline h-2.5 w-2.5" />{" "}
          Gateway code
        </p>
        <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900">
          Unmatched payments · [Entity Name]
        </h1>
        <p className="mt-1 max-w-2xl text-[10px] leading-relaxed text-slate-500">
          [Description body] Review incoming direct payments and reconcile them
          against student ledgers.
        </p>
      </div>
      <span className="inline-flex w-fit items-center gap-1.5 rounded-full bg-[#E2F0ED] px-3 py-1.5 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
        <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
        Queue health: [X] exceptions pending
      </span>
    </div>
  );
}

function ViewTabs({
  activeTab,
  onChange,
}: {
  activeTab: "unmatched" | "matched" | "all";
  onChange: (tab: "unmatched" | "matched" | "all") => void;
}) {
  const tabs = [
    { id: "unmatched" as const, label: "Unmatched", count: "[X]" },
    { id: "matched" as const, label: "Matched", count: "[Y]" },
    { id: "all" as const, label: "All", count: "[Z]" },
  ];

  return (
    <div className="mt-4 rounded-full bg-[#E8F5F3] p-1">
      <div className="grid grid-cols-3 gap-1" role="tablist" aria-label="Payment views">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={activeTab === tab.id}
            onClick={() => onChange(tab.id)}
            className={cn(
              "inline-flex items-center justify-center gap-1.5 rounded-full px-3 py-2 text-[10px] font-semibold transition",
              activeTab === tab.id
                ? "bg-white text-brand-700 shadow-sm"
                : "text-slate-400",
            )}
          >
            {tab.label}
            <span className="rounded-full bg-[#FEF3D6] px-1.5 py-0.5 text-[7px] text-[#8A6A25]">
              {tab.count}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}

function FilterToolbar() {
  return (
    <section className="mt-3 flex flex-col gap-2 rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card sm:flex-row sm:items-center">
      <label className="relative shrink-0 sm:w-52">
        <select className="h-9 w-full appearance-none rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-7 text-[9px] font-semibold text-slate-600 outline-none focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10">
          <option>[Current Academic Term]</option>
        </select>
        <ChevronDown className="pointer-events-none absolute right-2 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-400" />
      </label>
      <label className="relative min-w-0 flex-1">
        <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-400" />
        <input
          type="search"
          placeholder="Search transactions..."
          className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-8 pr-2.5 text-[9px] text-slate-700 outline-none placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
        />
      </label>
      <button
        type="button"
        className="grid h-9 w-9 shrink-0 place-items-center rounded-lg border border-slate-200/80 bg-white text-slate-500 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        aria-label="Refresh transactions"
      >
        <RefreshCw className="h-3.5 w-3.5" />
      </button>
    </section>
  );
}

function IncomingStreamTable({
  selectedId,
  onSelect,
}: {
  selectedId: string;
  onSelect: (id: string) => void;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Incoming Paybill Stream
          </h2>
          <p className="mt-0.5 text-[8px] text-slate-400">
            Live polling active · [Timestamp]
          </p>
        </div>
        <span className="inline-flex w-fit items-center gap-1 rounded-full bg-[#FEF3D6] px-2.5 py-1 text-[8px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
          <AlertTriangle className="h-3 w-3" />
          [X] requires manual triage
        </span>
      </div>

      <div className="mt-3 overflow-x-auto rounded-xl border border-slate-100">
        <table className="min-w-[600px] w-full border-collapse text-left">
          <thead>
            <tr className="border-b border-slate-100 bg-[#F8F9FA]">
              <TableHeading>Received</TableHeading>
              <TableHeading align="right">Amount</TableHeading>
              <TableHeading>Reference</TableHeading>
              <TableHeading>Payer Note</TableHeading>
              <TableHeading>Phone</TableHeading>
            </tr>
          </thead>
          <tbody>
            {TRANSACTIONS.map((row) => (
              <tr
                key={row.id}
                onClick={() => onSelect(row.id)}
                className={cn(
                  "cursor-pointer border-b border-slate-100 transition last:border-0",
                  selectedId === row.id
                    ? "bg-[#F1FAF8]"
                    : "hover:bg-slate-50/70",
                )}
              >
                <td className="px-3 py-3">
                  <p className="text-[9px] font-medium text-slate-600">
                    {row.received}
                  </p>
                  <p className="text-[7px] text-slate-400">[Timestamp]</p>
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
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-2.5 flex flex-col gap-1 text-[8px] text-slate-400 sm:flex-row sm:items-center sm:justify-between">
        <span className="inline-flex items-center gap-1">
          <ShieldCheck className="h-3 w-3 text-brand-600" />
          Cryptographic signature verification active
        </span>
        <span>[X] visible transactions</span>
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

function MetricsGrid() {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
      {METRICS.map((metric) => {
        const Icon = metric.icon;
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
              {metric.value}
            </p>
            <p className="mt-1 text-[8px] text-slate-400">
              [Secondary explanation]
            </p>
          </article>
        );
      })}
    </div>
  );
}

function AuditGuarantee() {
  return (
    <div className="flex gap-2.5 rounded-2xl border border-brand-100 bg-[#E2F0ED] p-3.5">
      <ShieldCheck className="h-4 w-4 shrink-0 text-brand-600" />
      <p className="text-[9px] leading-relaxed text-brand-800">
        [Immutable audit guarantee] Every assignment is timestamped and signed
        by the operator.
      </p>
    </div>
  );
}

function PaymentInspector({
  transaction,
}: {
  transaction: (typeof TRANSACTIONS)[number];
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
            <CreditCard className="h-4 w-4" />
          </span>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Active Payment Inspector
          </h2>
        </div>
        <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-500">
          ID: {transaction.reference}
        </span>
      </div>

      <div className="mt-4 rounded-xl border border-slate-100 bg-[#F8FAF9] p-3">
        <p className="text-[10px] font-bold text-slate-800">
          Assign payment {transaction.reference}
        </p>
        <p className="mt-2 text-2xl font-bold tracking-tight text-brand-700">
          {transaction.amount}
        </p>
        <p className="mt-0.5 text-[8px] text-slate-400">
          From {transaction.phone}
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
          <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-400" />
          <input
            type="search"
            placeholder="Search student name or admission number"
            className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-8 pr-8 text-[9px] text-slate-700 outline-none placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
          />
        </label>
        <div className="mt-2 rounded-xl border border-slate-100 bg-white p-3">
          <div className="flex items-center gap-2">
            <span className="grid h-9 w-9 place-items-center rounded-full bg-[#F0EBF9] text-[9px] font-bold text-[#6D4FA3]">
              [AB]
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-[10px] font-bold text-slate-800">
                [Student Name]
              </p>
              <p className="text-[8px] text-brand-600">[Admission #]</p>
              <p className="truncate text-[8px] text-slate-400">
                [Grade] · [Boarding status]
              </p>
            </div>
            <button
              type="button"
              className="grid h-6 w-6 place-items-center rounded-full bg-slate-100 text-slate-400 hover:bg-slate-200"
              aria-label="Remove selected student"
            >
              <X className="h-3 w-3" />
            </button>
          </div>
          <div className="mt-2 border-t border-slate-100 pt-2 text-[8px] text-slate-400">
            Guardian: [Guardian Name] · [Phone]
          </div>
        </div>
      </div>

      <div className="mt-3 rounded-xl bg-[#E2F0ED] p-3">
        <p className="text-[8px] font-semibold uppercase tracking-[0.12em] text-brand-700">
          Live ledger impact
        </p>
        <p className="mt-1 text-[9px] text-slate-600">
          Balance projection: [Amount] cleared of [Amount] total fees.
        </p>
        <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-white">
          <div className="h-full w-[62%] rounded-full bg-brand-500" />
        </div>
        <div className="mt-1.5 flex justify-between text-[7px] text-slate-400">
          <span>Cleared [Amount]</span>
          <span>Total [Amount]</span>
        </div>
      </div>

      <label className="mt-3 block">
        <span className="mb-1 block text-[8px] font-semibold text-slate-500">
          Reason for manual assignment <span className="text-rose-500">*</span>
        </span>
        <input
          type="text"
          placeholder="Enter audit reason"
          className="h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] px-2.5 text-[10px] text-slate-700 outline-none placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
        />
        <span className="mt-1 block text-[7px] text-slate-400">
          Required for the audit log.
        </span>
      </label>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#FEF3D6] p-2.5">
        <AlertTriangle className="h-3.5 w-3.5 shrink-0 text-[#8A6A25]" />
        <p className="text-[8px] leading-relaxed text-[#8A6A25]">
          [Overpayment warning] Excess credit will be carried forward on the
          ledger.
        </p>
      </div>

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row">
        <button
          type="button"
          className="h-9 flex-1 rounded-lg border border-slate-200/80 bg-white text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50"
        >
          Hold for later
        </button>
        <button
          type="button"
          className="inline-flex h-9 flex-1 items-center justify-center gap-1 rounded-lg bg-brand-500 text-[9px] font-bold text-white transition hover:bg-brand-600"
        >
          <Check className="h-3 w-3" />
          Assign payment
        </button>
      </div>
    </section>
  );
}

function QueueFooter() {
  return (
    <div className="mt-4 flex items-center justify-between gap-3 rounded-2xl border border-brand-100 bg-[#E2F0ED] p-3.5">
      <p className="inline-flex items-center gap-2 text-[9px] font-medium text-brand-800">
        <CheckCircle2 className="h-3.5 w-3.5" />
        [Empty state configuration status] Queue is clear.
      </p>
      <span className="rounded-full bg-white px-2.5 py-1 text-[8px] font-semibold text-brand-700">
        All Clear Ready
      </span>
    </div>
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
