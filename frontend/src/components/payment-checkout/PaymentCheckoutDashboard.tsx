"use client";

import { useState, type ReactNode } from "react";
import {
  ArrowLeft,
  ArrowRight,
  Award,
  Bot,
  CalendarCheck,
  Check,
  ChevronRight,
  CreditCard,
  FlaskConical,
  GraduationCap,
  Home,
  Info,
  Lock,
  LogOut,
  Pencil,
  ShieldCheck,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import {
  DEFAULT_PAYMENT_OPTION,
  PAYMENT_OPTIONS,
  type PaymentOption,
} from "./payment-checkout-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "/parent" },
  { label: "Results", icon: Award, href: "/parent/results" },
  { label: "Attendance", icon: CalendarCheck, href: "/parent/attendance" },
  { label: "Fees", icon: Wallet, href: "#", active: true },
  { label: "AI assistant", icon: Bot, href: "#" },
];

export default function PaymentCheckoutDashboard() {
  const [selectedOption, setSelectedOption] = useState(
    DEFAULT_PAYMENT_OPTION,
  );
  const [customAmount, setCustomAmount] = useState("");

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#payment-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to payment
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar />

      <div className="lg:pl-[260px]">
        <DesktopHeader />

        <main
          id="payment-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1200px] px-4 pb-10 pt-20 sm:px-6 lg:px-8 lg:pt-7"
        >
          <div className="hidden lg:block">
            <h1 className="text-2xl font-bold tracking-tight text-slate-900">
              Make Payment
            </h1>
            <p className="mt-1 text-xs text-slate-500">
              Secure M-Pesa checkout for the selected student account.
            </p>
          </div>

          <StudentBanner />

          <div className="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.45fr)_minmax(340px,1fr)]">
            <div className="min-w-0 space-y-4">
              <AmountCard
                selectedOption={selectedOption}
                customAmount={customAmount}
                onSelect={setSelectedOption}
                onCustomAmountChange={setCustomAmount}
              />
              <PhoneCard />
            </div>

            <div className="min-w-0 space-y-4 xl:sticky xl:top-[84px] xl:self-start">
              <SecurityNotice />
              <SandboxNotice />
              <PaymentActions />
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

function SidebarContent() {
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
              Parent portal
            </p>
            <span className="rounded-full bg-brand-50 px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wide text-brand-700 ring-1 ring-brand-100">
              Parent
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

      <div className="mt-auto pt-6">
        <div className="rounded-2xl border border-slate-200/80 bg-slate-50/70 p-3">
          <div className="flex items-center gap-2.5">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
              [AB]
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                [Guardian Name]
              </p>
              <p className="truncate text-[10px] text-slate-400">
                [Relationship]
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                [Email / Phone]
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

function MobileAppBar() {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <div className="flex min-w-0 items-center gap-2">
          <a
            href="/parent"
            className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Back to fees"
          >
            <ArrowLeft className="h-5 w-5" />
          </a>
          <p className="truncate text-lg font-bold tracking-tight text-slate-900">
            Make Payment
          </p>
        </div>

        <button
          type="button"
          className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open user profile"
        >
          [AB]
        </button>
      </div>
    </header>
  );
}

function DesktopHeader() {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1200px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="max-w-[220px] truncate text-slate-700">
            [Portal Name]
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Fees</span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Make Payment</span>
        </div>

        <button
          type="button"
          className="flex items-center gap-2.5 rounded-xl border border-slate-200/80 bg-white py-1.5 pl-1.5 pr-3 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
        >
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-[10px] font-bold text-white">
            [AB]
          </span>
          <div className="min-w-0 text-left">
            <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
              [Guardian Name]
            </p>
            <p className="text-[10px] text-slate-400">[Relationship]</p>
          </div>
          <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
        </button>
      </div>
    </header>
  );
}

function StudentBanner() {
  return (
    <div className="mt-4 flex items-center gap-2.5 rounded-full bg-[#F0EBF9] px-3 py-2.5 lg:mt-5">
      <span className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-white text-[#6D4FA3] shadow-sm">
        <GraduationCap className="h-3.5 w-3.5" />
      </span>
      <p className="min-w-0 truncate text-[10px] font-semibold text-[#6D4FA3]">
        Billing to [Student Name] · [Admission #]
      </p>
    </div>
  );
}

function AmountCard({
  selectedOption,
  customAmount,
  onSelect,
  onCustomAmountChange,
}: {
  selectedOption: string;
  customAmount: string;
  onSelect: (id: string) => void;
  onCustomAmountChange: (value: string) => void;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-start justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          How much would you like to pay?
        </h2>
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <Wallet className="h-4 w-4" />
        </span>
      </div>

      <div className="mt-3 space-y-2" role="radiogroup" aria-label="Payment amount">
        {PAYMENT_OPTIONS.map((option) => (
          <PaymentOptionRow
            key={option.id}
            option={option}
            selected={selectedOption === option.id}
            onSelect={() => onSelect(option.id)}
          >
            {option.custom && selectedOption === option.id && (
              <div className="mt-2.5 flex items-center gap-2 rounded-xl bg-white/80 p-2 ring-1 ring-[#E4DBF5]">
                <span className="pl-1 text-[10px] font-semibold text-slate-500">
                  KES
                </span>
                <input
                  type="text"
                  inputMode="numeric"
                  value={customAmount}
                  onChange={(event) =>
                    onCustomAmountChange(event.target.value)
                  }
                  placeholder="[Enter amount]"
                  className="h-9 min-w-0 flex-1 rounded-lg border border-slate-200/80 bg-white px-2.5 text-xs font-semibold text-slate-700 outline-none placeholder:text-slate-400 focus:border-brand-300 focus:ring-4 focus:ring-brand-500/10"
                />
              </div>
            )}
          </PaymentOptionRow>
        ))}
      </div>

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <Info className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          Overpayments are applied as a credit to the student&apos;s fee ledger
          and remain available for future charges.
        </p>
      </div>
    </section>
  );
}

function PaymentOptionRow({
  option,
  selected,
  onSelect,
  children,
}: {
  option: PaymentOption;
  selected: boolean;
  onSelect: () => void;
  children?: ReactNode;
}) {
  return (
    <div
      className={cn(
        "rounded-xl border transition",
        selected
          ? "border-[#D7C8F2] bg-[#F0EBF9]"
          : "border-slate-200/80 bg-white hover:bg-slate-50/60",
      )}
    >
      <label className="flex cursor-pointer items-center gap-2.5 p-3">
        <input
          type="radio"
          name="payment-amount"
          value={option.id}
          checked={selected}
          onChange={onSelect}
          className="sr-only"
        />
        <span
          className={cn(
            "grid h-4 w-4 shrink-0 place-items-center rounded-full border-2",
            selected
              ? "border-brand-500 bg-white"
              : "border-slate-300 bg-white",
          )}
          aria-hidden="true"
        >
          {selected && <span className="h-2 w-2 rounded-full bg-brand-500" />}
        </span>
        <span className="flex-1 text-[11px] font-semibold text-slate-700">
          {option.label}
        </span>
        {option.recommended && (
          <span className="rounded-full bg-white px-2 py-0.5 text-[8px] font-semibold text-[#6D4FA3] ring-1 ring-[#E4DBF5]">
            Recommended
          </span>
        )}
        {option.custom ? (
          <Pencil className="h-3.5 w-3.5 text-slate-400" />
        ) : (
          <span className="text-[11px] font-bold tabular-nums text-slate-800">
            {option.amount}
          </span>
        )}
      </label>
      {children && <div className="px-3 pb-3">{children}</div>}
    </div>
  );
}

function PhoneCard() {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          M-Pesa number to charge
        </h2>
        <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2 py-0.5 text-[8px] font-semibold text-brand-700 ring-1 ring-brand-100">
          <Check className="h-2.5 w-2.5" />
          Verified
        </span>
      </div>

      <label className="mt-3 flex h-11 items-center rounded-xl bg-[#E8F5F3] p-1">
        <span className="inline-flex h-9 items-center rounded-lg bg-white px-2.5 text-[10px] font-bold text-brand-700 shadow-sm">
          +254
        </span>
        <span className="sr-only">M-Pesa phone number</span>
        <input
          type="tel"
          inputMode="tel"
          placeholder="[Phone Number]"
          className="h-9 min-w-0 flex-1 bg-transparent px-2.5 text-xs font-semibold text-slate-700 outline-none placeholder:text-slate-400"
        />
      </label>

      <p className="mt-2.5 text-[9px] leading-relaxed text-slate-400">
        An STK prompt will be sent to this number. To charge a different
        guardian&apos;s phone, update the verified number first.
      </p>
    </section>
  );
}

function SecurityNotice() {
  return (
    <section className="flex gap-3 rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
        <Lock className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-xs font-bold text-slate-800">Secure PIN entry</p>
        <p className="mt-1 text-[10px] leading-relaxed text-slate-500">
          You&apos;ll enter your M-Pesa PIN on your phone. We never see, prompt,
          or store your PIN.
        </p>
      </div>
    </section>
  );
}

function SandboxNotice() {
  return (
    <div className="flex items-center gap-2 px-1 text-[9px] font-medium text-slate-400">
      <FlaskConical className="h-3 w-3 shrink-0" />
      Demo environment · Daraja sandbox. No real money moves.
    </div>
  );
}

function PaymentActions() {
  return (
    <div className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card">
      <button
        type="button"
        className="inline-flex h-11 w-full items-center justify-center gap-2 rounded-xl bg-brand-500 text-[11px] font-bold text-white shadow-lg shadow-brand-500/20 transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
      >
        <CreditCard className="h-4 w-4" />
        Send M-Pesa prompt
        <ArrowRight className="h-4 w-4" />
      </button>
      <a
        href="/parent"
        className="mt-3 flex items-center justify-center text-[10px] font-semibold text-slate-500 transition hover:text-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
      >
        Cancel and return to fees
      </a>
      <p className="mt-3 flex items-center justify-center gap-1 text-[8px] font-medium text-slate-400">
        <ShieldCheck className="h-3 w-3" />
        Secure checkout · 256-bit encryption
      </p>
    </div>
  );
}
