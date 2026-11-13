"use client";

import { useEffect, useState, type ReactNode } from "react";
import {
  AlertTriangle,
  ArrowRight,
  Bell,
  Building2,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock3,
  Copy,
  CreditCard,
  Download,
  GraduationCap,
  Key,
  Link2,
  Lock,
  LogOut,
  Menu,
  MessageCircle,
  RefreshCw,
  Send,
  ShieldCheck,
  Smartphone,
  Wallet,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import type {
  AdminConnectionsData,
  ConnectionsAuditRow,
} from "./admin-connections-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  countKey?: "schools" | "connections" | "unmatchedPayments";
  active?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Schools", icon: Building2, countKey: "schools" },
  { label: "Connections", icon: Link2, countKey: "connections", active: true },
  { label: "Fee Structures", icon: Wallet },
  { label: "Unmatched Payments", icon: AlertTriangle, countKey: "unmatchedPayments" },
  { label: "System Audit", icon: ShieldCheck },
];

export default function AdminConnectionsDashboard({
  data,
}: {
  data: AdminConnectionsData | null;
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
  };
  const counts = data?.counts ?? {
    schools: null,
    connections: null,
    unmatchedPayments: null,
  };
  const [paymentType, setPaymentType] = useState<"paybill" | "till">(
    data?.daraja?.paymentType ?? "paybill",
  );
  const [environment, setEnvironment] = useState<"sandbox" | "live">(
    data?.daraja?.environment ?? "sandbox",
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
        href="#connections-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to connections
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
          id="connections-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <PolicyNotice />
          <PageHeader header={data?.header ?? null} />

          <div className="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
            <WhatsAppCard whatsapp={data?.whatsapp ?? null} />
            <DarajaCard
              daraja={data?.daraja ?? null}
              paymentType={paymentType}
              environment={environment}
              onPaymentTypeChange={setPaymentType}
              onEnvironmentChange={setEnvironment}
            />
          </div>

          <AuditTable rows={data?.auditRows ?? []} />
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
  counts: AdminConnectionsData["counts"];
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
          aria-controls="connections-drawer"
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
            Connections
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

function PolicyNotice() {
  return (
    <section className="flex flex-col gap-3 rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex gap-3">
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
          <ShieldCheck className="h-4 w-4" />
        </span>
        <div className="min-w-0">
          <p className="text-xs font-bold text-slate-800">Tenant policy notice</p>
          <p className="mt-1 max-w-3xl text-[10px] leading-relaxed text-slate-500">
            Credentials are write-only and never returned in plaintext after
            they are saved.
          </p>
        </div>
      </div>
      <span className="inline-flex w-fit shrink-0 items-center gap-1 rounded-full bg-white px-2.5 py-1 text-[8px] font-semibold text-[#6D4FA3] ring-1 ring-[#E4DBF5]">
        <Lock className="h-3 w-3" />
        PCI &amp; Meta Compliant
      </span>
    </section>
  );
}

function PageHeader({
  header,
}: {
  header: AdminConnectionsData["header"] | null;
}) {
  return (
    <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p className="text-[8px] font-medium text-slate-400">
          Tenant link reference <ChevronRight className="inline h-2.5 w-2.5" />{" "}
          Scope
        </p>
        <h1 className="mt-1 text-2xl font-bold tracking-tight text-slate-900">
          Connections
        </h1>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <span className="inline-flex items-center gap-1 rounded-full bg-white px-2.5 py-1 text-[9px] font-semibold text-slate-500 ring-1 ring-slate-200/80">
          <Smartphone className="h-3 w-3 text-slate-400" />
          {header?.regionLabel || "No region selected"}
        </span>
        <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[9px] font-semibold text-slate-500">
          {header?.accountId || "No account ID"}
        </span>
        <button
          type="button"
          className="grid h-8 w-8 place-items-center rounded-full border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Refresh connections"
        >
          <RefreshCw className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  );
}

function WhatsAppCard({
  whatsapp,
}: {
  whatsapp: AdminConnectionsData["whatsapp"];
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-center gap-2">
          <span className="grid h-9 w-9 place-items-center rounded-xl bg-[#E2F0ED] text-brand-700">
            <MessageCircle className="h-4 w-4" />
          </span>
          <div>
            <h2 className="text-sm font-bold tracking-tight text-slate-900">
              WhatsApp Bot Integration
            </h2>
            <p className="mt-0.5 text-[8px] text-slate-400">
              API version and gateway subtype not loaded
            </p>
          </div>
        </div>
        <div className="flex flex-wrap gap-1.5">
          <span
            className={cn(
              "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[8px] font-semibold ring-1",
              whatsapp?.connected
                ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
                : "bg-slate-100 text-slate-500 ring-slate-200",
            )}
          >
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                whatsapp?.connected ? "bg-brand-500" : "bg-slate-400",
              )}
            />
            {whatsapp?.connected ? "Connected" : "Not connected"}
          </span>
          <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[8px] font-semibold text-slate-500">
            {whatsapp?.receiving ? "Receiving messages" : "Not receiving"}
          </span>
        </div>
      </div>

      <div className="mt-4 space-y-2.5">
        <FormInput
          label="Business display name"
          placeholder="Enter business display name"
          defaultValue={whatsapp?.displayName ?? ""}
        />
        <label className="block">
          <span className="mb-1 block text-[8px] font-semibold text-slate-500">
            Display number
          </span>
          <div className="flex h-9 items-center rounded-lg border border-slate-200/80 bg-[#F8FAF9] p-0.5">
            <span className="inline-flex h-8 items-center rounded-md bg-white px-2 text-[9px] font-bold text-brand-700 shadow-sm">
              +254
            </span>
            <input
              type="tel"
              defaultValue={whatsapp?.displayNumber ?? ""}
              placeholder="Enter display number"
              className="h-8 min-w-0 flex-1 bg-transparent px-2 text-[10px] text-slate-700 outline-none placeholder:text-slate-400"
            />
          </div>
        </label>
        <label className="block">
          <span className="mb-1 block text-[8px] font-semibold text-slate-500">
            Inbound webhook URL
          </span>
          <div className="flex h-9 items-center rounded-lg border border-slate-200/80 bg-[#F8FAF9] pl-2.5 pr-1">
            <input
              readOnly
              value={whatsapp?.webhookUrl ?? ""}
              placeholder="No webhook URL configured"
              className="h-8 min-w-0 flex-1 bg-transparent text-[10px] text-slate-500 outline-none placeholder:text-slate-400"
            />
            <button
              type="button"
              disabled={!whatsapp?.webhookUrl}
              className="inline-flex h-7 items-center gap-1 rounded-md bg-white px-2 text-[8px] font-semibold text-brand-700 shadow-sm ring-1 ring-slate-200/80 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            >
              <Copy className="h-2.5 w-2.5" />
              Copy
            </button>
          </div>
        </label>
        <FormInput
          label="Webhook verification token"
          placeholder="Enter verification token"
          type="password"
          helper={`Token rotation · ${
            whatsapp?.tokenRotatedAt || "No rotation recorded"
          }`}
        />
      </div>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          {whatsapp?.lastTestStatus || "No test delivery recorded"}
        </p>
      </div>

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row">
        <button
          type="button"
          className="inline-flex h-9 flex-1 items-center justify-center gap-1.5 rounded-lg border border-slate-200/80 bg-white text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <Send className="h-3 w-3" />
          Send test message
        </button>
        <button
          type="button"
          className="inline-flex h-9 flex-1 items-center justify-center gap-1 rounded-lg bg-brand-500 text-[9px] font-bold text-white transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          Save WhatsApp settings
        </button>
      </div>
    </section>
  );
}

function DarajaCard({
  daraja,
  paymentType,
  environment,
  onPaymentTypeChange,
  onEnvironmentChange,
}: {
  daraja: AdminConnectionsData["daraja"];
  paymentType: "paybill" | "till";
  environment: "sandbox" | "live";
  onPaymentTypeChange: (value: "paybill" | "till") => void;
  onEnvironmentChange: (value: "sandbox" | "live") => void;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-center gap-2">
          <span className="grid h-9 w-9 place-items-center rounded-xl bg-[#F0EBF9] text-[#6D4FA3]">
            <CreditCard className="h-4 w-4" />
          </span>
          <div>
            <h2 className="text-sm font-bold tracking-tight text-slate-900">
              M-Pesa / Daraja Integration
            </h2>
            <p className="mt-0.5 text-[8px] text-slate-400">
              Gateway credentials not loaded
            </p>
          </div>
        </div>
        <div className="flex rounded-full bg-slate-50 p-0.5 ring-1 ring-slate-200/80">
          <button
            type="button"
            onClick={() => onPaymentTypeChange("paybill")}
            className={cn(
              "rounded-full px-2.5 py-1 text-[8px] font-semibold transition",
              paymentType === "paybill"
                ? "bg-white text-brand-700 shadow-sm"
                : "text-slate-400",
            )}
          >
            Paybill
          </button>
          <button
            type="button"
            onClick={() => onPaymentTypeChange("till")}
            className={cn(
              "rounded-full px-2.5 py-1 text-[8px] font-semibold transition",
              paymentType === "till"
                ? "bg-white text-brand-700 shadow-sm"
                : "text-slate-400",
            )}
          >
            Till
          </button>
        </div>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-2">
        <FormInput
          label="Business shortcode"
          placeholder="Enter shortcode"
          defaultValue={daraja?.shortcode ?? ""}
        />
        <div>
          <p className="mb-1 text-[8px] font-semibold text-slate-500">
            Environment
          </p>
          <div className="flex h-9 rounded-lg bg-[#F8FAF9] p-0.5 ring-1 ring-slate-200/80">
            <button
              type="button"
              onClick={() => onEnvironmentChange("sandbox")}
              className={cn(
                "flex-1 rounded-md text-[8px] font-semibold",
                environment === "sandbox"
                  ? "bg-white text-brand-700 shadow-sm"
                  : "text-slate-400",
              )}
            >
              Sandbox
            </button>
            <button
              type="button"
              onClick={() => onEnvironmentChange("live")}
              className={cn(
                "flex-1 rounded-md text-[8px] font-semibold",
                environment === "live"
                  ? "bg-[#FEF3D6] text-[#8A6A25] shadow-sm"
                  : "text-slate-400",
              )}
            >
              Live
            </button>
          </div>
          <p className="mt-1 text-[7px] text-[#8A6A25]">
            {environment === "live"
              ? "Live mode uses real payment credentials."
              : "Sandbox mode does not move real money."}
          </p>
        </div>
      </div>

      <div className="mt-2.5 space-y-2.5">
        <FormInput
          label="Online passkey (STK key)"
          placeholder={
            daraja?.passkeySet ? "••••••••" : "Enter online passkey"
          }
          type="password"
          helper="Stored as a write-only secret."
          icon={Key}
        />
        <div className="grid grid-cols-2 gap-2">
          <FormInput
            label="Consumer key"
            placeholder={
              daraja?.consumerKeySet ? "••••••••" : "Enter consumer key"
            }
            type="password"
          />
          <FormInput
            label="Consumer secret"
            placeholder={
              daraja?.consumerSecretSet ? "••••••••" : "Enter consumer secret"
            }
            type="password"
          />
        </div>
      </div>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          {daraja?.lastTestStatus || "No transaction test recorded"}
        </p>
      </div>

      <div className="mt-3 flex flex-col gap-2 border-t border-slate-100 pt-3 sm:flex-row">
        <button
          type="button"
          className="inline-flex h-9 flex-1 items-center justify-center gap-1.5 rounded-lg border border-slate-200/80 bg-white text-[9px] font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <ArrowRight className="h-3 w-3" />
          Run test payment
        </button>
        <button
          type="button"
          className="inline-flex h-9 flex-1 items-center justify-center gap-1 rounded-lg bg-brand-500 text-[9px] font-bold text-white transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          Save Daraja settings
        </button>
      </div>
    </section>
  );
}

function FormInput({
  label,
  placeholder,
  type = "text",
  helper,
  icon: Icon,
  defaultValue,
}: {
  label: string;
  placeholder: string;
  type?: string;
  helper?: string;
  icon?: LucideIcon;
  defaultValue?: string;
}) {
  return (
    <label className="block">
      <span className="mb-1 block text-[8px] font-semibold text-slate-500">
        {label}
      </span>
      <div className="relative">
        {Icon && (
          <Icon className="pointer-events-none absolute left-2.5 top-1/2 h-3 w-3 -translate-y-1/2 text-slate-400" />
        )}
        <input
          type={type}
          defaultValue={defaultValue}
          placeholder={placeholder}
          className={cn(
            "h-9 w-full rounded-lg border border-slate-200/80 bg-[#F8FAF9] pr-2.5 text-[10px] text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10",
            Icon ? "pl-8" : "pl-2.5",
          )}
        />
      </div>
      {helper && (
        <span className="mt-1 block text-[7px] text-slate-400">{helper}</span>
      )}
    </label>
  );
}

function AuditTable({ rows }: { rows: ConnectionsAuditRow[] }) {
  return (
    <section className="mt-4 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
            <Clock3 className="h-4 w-4" />
          </span>
          <div>
            <h2 className="text-sm font-bold tracking-tight text-slate-900">
              Connection Audit History
            </h2>
            <p className="mt-0.5 text-[8px] text-slate-400">
              Cryptographic write-only audit trail
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2.5 py-1 text-[8px] font-semibold text-brand-700 ring-1 ring-brand-100">
            <Lock className="h-3 w-3" />
            Secrets masked · write-only audit
          </span>
          <button
            type="button"
            disabled={rows.length === 0}
            className="grid h-8 w-8 place-items-center rounded-lg border border-slate-200/80 bg-white text-slate-500 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Download audit CSV"
          >
            <Download className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-slate-100">
        {rows.length > 0 ? (
          <>
            <div className="hidden overflow-x-auto sm:block">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="border-b border-slate-100 bg-[#F8F9FA]">
                    <TableHeading>Timestamp</TableHeading>
                    <TableHeading>Operator</TableHeading>
                    <TableHeading>Event Detail</TableHeading>
                    <TableHeading align="right">Secrets Status</TableHeading>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => (
                    <tr
                      key={row.id}
                      className="border-b border-slate-100 last:border-0"
                    >
                      <td className="px-3 py-3 text-[9px] text-slate-500">
                        {row.timestamp}
                      </td>
                      <td className="px-3 py-3">
                        <div className="flex items-center gap-2">
                          <span className="grid h-6 w-6 place-items-center rounded-full bg-[#F0EBF9] text-[7px] font-bold text-[#6D4FA3]">
                            {row.initials}
                          </span>
                          <div>
                            <p className="text-[9px] font-semibold text-slate-700">
                              {row.operator}
                            </p>
                            <p className="text-[7px] text-slate-400">
                              {row.role}
                            </p>
                          </div>
                        </div>
                      </td>
                      <td className="px-3 py-3">
                        <span className="inline-flex items-center gap-1.5 text-[9px] text-slate-600">
                          <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
                          {row.event}
                        </span>
                      </td>
                      <td className="px-3 py-3 text-right">
                        <AuditBadge row={row} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="divide-y divide-slate-100 sm:hidden">
              {rows.map((row) => (
                <div key={row.id} className="p-3">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-[8px] text-slate-400">
                      {row.timestamp}
                    </span>
                    <AuditBadge row={row} />
                  </div>
                  <div className="mt-2 flex items-center gap-2">
                    <span className="grid h-6 w-6 place-items-center rounded-full bg-[#F0EBF9] text-[7px] font-bold text-[#6D4FA3]">
                      {row.initials}
                    </span>
                    <div className="min-w-0">
                      <p className="text-[9px] font-semibold text-slate-700">
                        {row.operator}
                      </p>
                      <p className="text-[7px] text-slate-400">{row.role}</p>
                    </div>
                  </div>
                  <p className="mt-2 inline-flex items-center gap-1.5 text-[8px] text-slate-600">
                    <span className="h-1.5 w-1.5 rounded-full bg-brand-500" />
                    {row.event}
                  </p>
                </div>
              ))}
            </div>
          </>
        ) : (
          <div className="px-4 py-8 text-center">
            <Clock3 className="mx-auto h-4 w-4 text-slate-400" />
            <p className="mt-2 text-[11px] font-semibold text-slate-700">
              No audit records available
            </p>
            <p className="mx-auto mt-1 max-w-xs text-[9px] leading-relaxed text-slate-500">
              Connection changes will appear here when they are recorded.
            </p>
          </div>
        )}
      </div>
    </section>
  );
}

function AuditBadge({ row }: { row: ConnectionsAuditRow }) {
  return (
    <span
      className={cn(
        "inline-flex rounded-full px-2 py-0.5 text-[8px] font-semibold ring-1",
        row.status === "masked" &&
          "bg-slate-100 text-slate-500 ring-slate-200",
        row.status === "rotated" &&
          "bg-[#FEF3D6] text-[#8A6A25] ring-[#EBD7A5]",
        row.status === "verified" &&
          "bg-[#E2F0ED] text-brand-700 ring-brand-100",
      )}
    >
      {row.statusLabel}
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
        "whitespace-nowrap px-3 py-2.5 text-[8px] font-semibold uppercase tracking-[0.1em] text-slate-400",
        align === "right" ? "text-right" : "text-left",
      )}
    >
      {children}
    </th>
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
  counts: AdminConnectionsData["counts"];
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
        id="connections-drawer"
        role="dialog"
        aria-modal="true"
        aria-labelledby="connections-menu-title"
        tabIndex={-1}
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl outline-none animate-drawer-in motion-reduce:animate-none"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <h2
            id="connections-menu-title"
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
