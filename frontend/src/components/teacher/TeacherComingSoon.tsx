import Link from "next/link";
import { ArrowLeft, CalendarDays, Construction } from "lucide-react";

/**
 * Placeholder page for teacher routes that are not implemented yet.
 * Keeps internal navigation valid until the full screen is available.
 */
export default function TeacherComingSoon({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <main className="grid min-h-screen place-items-center bg-[#F4F7F6] px-4 py-10 text-slate-900">
      <section className="w-full max-w-lg rounded-3xl border border-slate-200/80 bg-white p-8 text-center shadow-card">
        <span className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-[#E2F0ED] text-brand-700">
          <Construction className="h-6 w-6" />
        </span>
        <p className="mt-5 text-[10px] font-semibold uppercase tracking-[0.18em] text-brand-600">
          Teacher portal
        </p>
        <h1 className="mt-2 text-2xl font-bold tracking-tight text-slate-900">
          {title}
        </h1>
        <p className="mx-auto mt-3 max-w-sm text-sm leading-relaxed text-slate-600">
          {description}
        </p>
        <div className="mt-6 flex flex-col justify-center gap-2 sm:flex-row">
          <Link
            href="/teacher"
            className="inline-flex h-11 items-center justify-center gap-2 rounded-xl bg-brand-500 px-4 text-xs font-semibold text-white shadow-sm transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
          >
            <CalendarDays className="h-4 w-4" />
            Open timetable
          </Link>
          <Link
            href="/teacher/class-means"
            className="inline-flex h-11 items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white px-4 text-xs font-semibold text-slate-600 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to class means
          </Link>
        </div>
      </section>
    </main>
  );
}
