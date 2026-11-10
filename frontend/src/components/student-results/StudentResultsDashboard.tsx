"use client";

import { useState } from "react";
import {
  Award,
  BarChart3,
  Bell,
  CalendarCheck,
  ChevronRight,
  FileText,
  GraduationCap,
  Home,
  Info,
  LogOut,
  MessageCircle,
  Quote,
  ShieldCheck,
  Sparkles,
  TrendingUp,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import type {
  AdjustmentNotice,
  PerformanceLevel,
  ResultsTerm,
  StudentResultsData,
  SubjectResult,
  TeacherComment,
} from "./student-results-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "#" },
  { label: "Results", icon: Award, href: "/parent/results", active: true },
  { label: "Attendance", icon: CalendarCheck, href: "#" },
  { label: "Fees", icon: Wallet, href: "#" },
  { label: "AI assistant", icon: Sparkles, href: "#" },
];

const BOTTOM_NAV = [
  { label: "Home", icon: Home, href: "#" },
  { label: "Results", icon: BarChart3, href: "/parent/results", active: true },
  { label: "Fees", icon: Wallet, href: "#" },
  { label: "Assistant", icon: MessageCircle, href: "#" },
];

const LEVEL_TONES: Record<PerformanceLevel["tone"], string> = {
  slate: "bg-slate-100 text-slate-600",
  lavender: "bg-[#F0EBF9] text-[#6D4FA3]",
  purple: "bg-[#D8CBF0] text-[#4F3585]",
  indigo: "bg-[#4E3A8A] text-white",
};

export default function StudentResultsDashboard({
  data,
}: {
  data: StudentResultsData | null;
}) {
  const students = data?.students ?? [];
  const terms = data?.terms ?? [];
  const results = data?.results ?? [];
  const levels = data?.levels ?? [];
  const adjustment = data?.adjustment ?? null;
  const teacherComment = data?.teacherComment ?? null;
  const portalName = data?.portal?.name?.trim() || "Parent portal";
  const session = data?.session ?? null;
  const guardianName = session?.fullName?.trim() || "Guardian";
  const relationship = session?.relationship?.trim() || "Parent";
  const contact = session?.contact?.trim() || "No contact on file";
  const guardianInitials =
    session?.initials?.trim() || initialsFromName(guardianName);

  const [selectedStudentId, setSelectedStudentId] = useState(
    data?.selectedStudentId ?? students[0]?.id ?? "",
  );
  const [selectedTermId, setSelectedTermId] = useState(
    data?.selectedTermId ??
      terms.find((term) => term.released)?.id ??
      terms[0]?.id ??
      "",
  );

  const selectedTerm =
    terms.find((term) => term.id === selectedTermId) ??
    terms.find((term) => term.released) ??
    terms[0] ??
    null;
  const releaseInfo = selectedTerm?.released
    ? selectedTerm.releaseDate
      ? `Released ${selectedTerm.releaseDate} · ${results.length} learning areas`
      : `Released · ${results.length} learning areas`
    : results.length > 0
      ? `${results.length} learning areas`
      : "No release information";

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <a
        href="#results-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-[60] focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-brand-700 focus:shadow-lg"
      >
        Skip to results
      </a>

      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[260px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          guardianName={guardianName}
          relationship={relationship}
          contact={contact}
          guardianInitials={guardianInitials}
        />
      </aside>

      <MobileAppBar
        portalName={portalName}
        guardianInitials={guardianInitials}
      />

      <div className="lg:pl-[260px]">
        <DesktopHeader
          portalName={portalName}
          guardianName={guardianName}
          relationship={relationship}
          guardianInitials={guardianInitials}
        />

        <main
          id="results-main"
          tabIndex={-1}
          className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7"
        >
          <StudentFilter
            students={students}
            selectedStudentId={selectedStudentId}
            onSelect={setSelectedStudentId}
          />
          <TermSelector
            terms={terms}
            selectedTermId={selectedTerm?.id ?? ""}
            onSelect={setSelectedTermId}
          />

          <div className="mt-4 grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_minmax(360px,1fr)]">
            <div className="min-w-0 space-y-4">
              <SubjectPerformanceCard
                results={results}
                releaseInfo={releaseInfo}
              />
              <AdjustmentNoticeCard adjustment={adjustment} />
              <TeacherCommentCard comment={teacherComment} />
            </div>

            <div className="min-w-0 xl:sticky xl:top-[84px] xl:self-start">
              <LevelsCard levels={levels} />
            </div>
          </div>
        </main>
      </div>

      <BottomNav />
    </div>
  );
}

function SidebarContent({
  guardianName,
  relationship,
  contact,
  guardianInitials,
}: {
  guardianName: string;
  relationship: string;
  contact: string;
  guardianInitials: string;
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
              {guardianInitials}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                {guardianName}
              </p>
              <p className="truncate text-[10px] text-slate-400">
                {relationship}
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                {contact}
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
  portalName,
  guardianInitials,
}: {
  portalName: string;
  guardianInitials: string;
}) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <div className="min-w-0 flex-1">
          <p className="truncate text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            {portalName}
          </p>
          <p className="truncate text-lg font-bold tracking-tight text-slate-900">
            Results
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
            {guardianInitials}
          </button>
        </div>
      </div>
    </header>
  );
}

function DesktopHeader({
  portalName,
  guardianName,
  relationship,
  guardianInitials,
}: {
  portalName: string;
  guardianName: string;
  relationship: string;
  guardianInitials: string;
}) {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="max-w-[220px] truncate text-slate-700">
            {portalName}
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>Results</span>
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
              {guardianInitials}
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                {guardianName}
              </p>
              <p className="text-[10px] text-slate-400">{relationship}</p>
            </div>
            <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

function StudentFilter({
  students,
  selectedStudentId,
  onSelect,
}: {
  students: StudentResultsData["students"];
  selectedStudentId: string;
  onSelect: (id: string) => void;
}) {
  return (
    <div
      className="-mx-4 overflow-x-auto px-4 pb-1 sm:mx-0 sm:px-0"
      role="tablist"
      aria-label="Student records"
    >
      {students.length > 0 ? (
        <div className="flex min-w-max gap-2">
          {students.map((student) => {
            const active = student.id === selectedStudentId;

            return (
              <button
                key={student.id}
                type="button"
                role="tab"
                aria-selected={active}
                onClick={() => onSelect(student.id)}
                className={cn(
                  "inline-flex items-center gap-2 rounded-full px-3 py-2 text-[11px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40 focus-visible:ring-offset-2",
                  active
                    ? "bg-brand-500 text-white shadow-md shadow-brand-500/20"
                    : "border border-slate-200/80 bg-white text-slate-500 hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700",
                )}
              >
                <span
                  className={cn(
                    "grid h-6 w-6 place-items-center rounded-full text-[8px] font-bold",
                    active
                      ? "bg-white/20 text-white"
                      : "bg-slate-100 text-slate-500",
                  )}
                >
                  {student.initials}
                </span>
                {student.name}
              </button>
            );
          })}
        </div>
      ) : (
        <span className="inline-flex items-center rounded-full border border-dashed border-slate-300 bg-white px-3 py-2 text-[11px] font-medium text-slate-500">
          No student records available
        </span>
      )}
    </div>
  );
}

function TermSelector({
  terms,
  selectedTermId,
  onSelect,
}: {
  terms: ResultsTerm[];
  selectedTermId: string;
  onSelect: (id: string) => void;
}) {
  return (
    <div className="mt-3 rounded-full bg-[#E8F5F3] p-1">
      <div
        className="grid gap-1"
        style={{
          gridTemplateColumns: `repeat(${Math.max(terms.length, 1)}, minmax(0, 1fr))`,
        }}
        role="tablist"
        aria-label="Terms"
      >
        {terms.length > 0 ? (
          terms.map((term) => {
            const active = term.released && term.id === selectedTermId;

            return (
              <button
                key={term.id}
                type="button"
                role="tab"
                aria-selected={active}
                disabled={!term.released}
                onClick={() => onSelect(term.id)}
                className={cn(
                  "rounded-full px-2 py-2 text-[10px] font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40",
                  active
                    ? "bg-white text-brand-700 shadow-sm"
                    : "text-slate-400",
                  !term.released &&
                    "cursor-not-allowed font-medium text-slate-400 opacity-70",
                )}
              >
                {term.label}
                {!term.released ? " · Not released" : ""}
              </button>
            );
          })
        ) : (
          <span className="col-span-full rounded-full px-2 py-2 text-center text-[10px] font-medium text-slate-400">
            No terms available
          </span>
        )}
      </div>
    </div>
  );
}

function SubjectPerformanceCard({
  results,
  releaseInfo,
}: {
  results: SubjectResult[];
  releaseInfo: string;
}) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-sm font-bold tracking-tight text-slate-900">
            Subject performance levels
          </h2>
          <p className="mt-1 text-[10px] text-slate-400">{releaseInfo}</p>
        </div>
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <BarChart3 className="h-4 w-4" />
        </span>
      </div>

      {results.length > 0 ? (
        <div className="mt-3">
          {results.map((row) => (
            <div
              key={row.id}
              className="flex items-center justify-between gap-3 border-b border-slate-100 py-3 last:border-0"
            >
              <div className="min-w-0">
                <div className="flex items-center gap-1.5">
                  <p className="truncate text-[11px] font-bold text-slate-800">
                    {row.subject}
                  </p>
                  {row.corrected && (
                    <span className="shrink-0 rounded-full bg-[#FEF3D6] px-1.5 py-0.5 text-[8px] font-semibold text-[#8A6A25] ring-1 ring-[#EBD7A5]">
                      Corrected
                    </span>
                  )}
                </div>
                <p className="mt-0.5 flex items-center gap-1 text-[9px] font-medium text-slate-400">
                  <TrendingUp className="h-3 w-3 text-brand-500" />
                  {row.performanceLevel}
                </p>
              </div>

              <div className="flex shrink-0 items-center gap-0.5 rounded-full bg-slate-50 p-0.5 ring-1 ring-slate-200/80">
                {(["T1", "T2", "T3"] as const).map((term) => (
                  <TermPill
                    key={term}
                    label={term}
                    active={row.activeTerms.includes(term)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          icon={BarChart3}
          title="No subject results available"
          body="Released subject performance levels will appear here when the school publishes them."
          className="mt-3"
        />
      )}

      <div className="mt-3 rounded-xl bg-[#E2F0ED] px-3 py-2.5">
        <div className="flex flex-wrap gap-x-4 gap-y-1 text-[9px] font-medium text-brand-700">
          <span>T1 · Term 1</span>
          <span>T2 · Term 2</span>
          <span>T3 · Term 3</span>
        </div>
      </div>
    </section>
  );
}

function TermPill({ label, active = false }: { label: string; active?: boolean }) {
  return (
    <span
      className={cn(
        "grid h-6 min-w-7 place-items-center rounded-full px-1.5 text-[8px] font-bold",
        active ? "bg-brand-500 text-white" : "text-slate-300",
      )}
    >
      {label}
    </span>
  );
}

function AdjustmentNoticeCard({
  adjustment,
}: {
  adjustment: AdjustmentNotice | null;
}) {
  return (
    <section className="rounded-2xl border border-[#E4DBF5] bg-[#F0EBF9] p-4">
      <div className="flex gap-3">
        <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-[#6D4FA3] shadow-sm">
          <FileText className="h-4 w-4" />
        </span>
        <div className="min-w-0 flex-1">
          {adjustment ? (
            <>
              <p className="text-xs font-bold text-slate-800">
                {adjustment.title}
              </p>
              <p className="mt-1 text-[11px] leading-relaxed text-slate-500">
                {adjustment.description}
              </p>
              <div className="mt-3 flex items-center justify-between gap-3">
                {adjustment.noteHref ? (
                  <a
                    href={adjustment.noteHref}
                    className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-[#D7C8F2] bg-white px-2.5 text-[9px] font-semibold text-[#6D4FA3] transition hover:bg-[#F7F3FD] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
                  >
                    <FileText className="h-3 w-3" />
                    Read note
                  </a>
                ) : (
                  <span />
                )}
                <span className="text-[9px] font-medium text-slate-400">
                  {adjustment.timestamp}
                </span>
              </div>
            </>
          ) : (
            <div className="rounded-xl border border-dashed border-[#D7C8F2] bg-white/60 px-4 py-5 text-center">
              <p className="text-[11px] font-semibold text-slate-700">
                No grade adjustments recorded
              </p>
              <p className="mx-auto mt-1 max-w-xs text-[9px] leading-relaxed text-slate-500">
                Correction notes will appear here when the school records an
                adjustment.
              </p>
            </div>
          )}
        </div>
      </div>
    </section>
  );
}

function TeacherCommentCard({ comment }: { comment: TeacherComment | null }) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          Class teacher&apos;s comment
        </h2>
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-[#F0EBF9] text-[#6D4FA3]">
          <Quote className="h-4 w-4" />
        </span>
      </div>

      {comment ? (
        <>
          <div className="mt-3 rounded-xl bg-[#E8F5F3] p-4">
            <p className="text-[11px] italic leading-relaxed text-slate-600">
              “{comment.comment}”
            </p>
          </div>

          <div className="mt-4 flex items-center gap-2.5 border-t border-slate-100 pt-3">
            <span className="grid h-9 w-9 place-items-center rounded-full bg-slate-900 text-[9px] font-bold text-white">
              {comment.initials}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-[11px] font-bold text-slate-800">
                {comment.teacherName}
              </p>
              <p className="text-[9px] text-slate-400">{comment.role}</p>
            </div>
            {comment.verified && (
              <span className="inline-flex items-center gap-1 rounded-full bg-[#E2F0ED] px-2 py-0.5 text-[8px] font-semibold text-brand-700 ring-1 ring-brand-100">
                <ShieldCheck className="h-3 w-3" />
                Verified
              </span>
            )}
          </div>
        </>
      ) : (
        <EmptyState
          icon={Quote}
          title="No class teacher comment available"
          body="The class teacher's comment will appear here after results are released."
          className="mt-3"
        />
      )}
    </section>
  );
}

function LevelsCard({ levels }: { levels: PerformanceLevel[] }) {
  return (
    <section className="rounded-2xl border border-slate-200/80 bg-white p-4 shadow-card sm:p-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-bold tracking-tight text-slate-900">
          How levels work
        </h2>
        <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-50 text-brand-600">
          <GraduationCap className="h-4 w-4" />
        </span>
      </div>

      {levels.length > 0 ? (
        <div className="mt-3 grid grid-cols-2 gap-2">
          {levels.map((level) => (
            <div
              key={level.id}
              className={cn("rounded-xl p-3", LEVEL_TONES[level.tone])}
            >
              <p className="text-[8px] font-semibold uppercase tracking-wide opacity-60">
                {level.level}
              </p>
              <p className="mt-1 text-[11px] font-bold">{level.title}</p>
              <p className="mt-1 text-[9px] leading-relaxed opacity-70">
                {level.description}
              </p>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState
          icon={GraduationCap}
          title="Level guide not available"
          body="Performance level definitions will appear here when the school publishes them."
          className="mt-3"
        />
      )}

      <div className="mt-3 flex gap-2 rounded-xl bg-[#E2F0ED] p-3">
        <Info className="h-3.5 w-3.5 shrink-0 text-brand-600" />
        <p className="text-[9px] leading-relaxed text-brand-800">
          Levels describe achievement against learning outcomes, not comparison
          with other learners.
        </p>
      </div>
    </section>
  );
}

function EmptyState({
  icon: Icon,
  title,
  body,
  className,
  tone = "slate",
}: {
  icon: LucideIcon;
  title: string;
  body: string;
  className?: string;
  tone?: "slate" | "lavender";
}) {
  return (
    <div
      className={cn(
        "rounded-xl border border-dashed px-4 py-6 text-center",
        tone === "slate"
          ? "border-slate-200 bg-slate-50/60"
          : "border-[#D7C8F2] bg-white/60",
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

function BottomNav() {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-4 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
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

function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "G";
}
