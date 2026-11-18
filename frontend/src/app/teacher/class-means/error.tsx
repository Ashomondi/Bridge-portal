"use client";

import { AlertTriangle, RefreshCw } from "lucide-react";

/** Error recovery screen for the class means route. */
export default function ClassMeansError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main className="grid min-h-screen place-items-center bg-[#F4F7F6] px-4 py-10">
      <section className="w-full max-w-md rounded-3xl border border-slate-200/80 bg-white p-8 text-center shadow-card">
        <span className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-[#FEF3D6] text-[#8A6A25]">
          <AlertTriangle className="h-6 w-6" />
        </span>
        <h1 className="mt-5 text-xl font-bold tracking-tight text-slate-900">
          Class means could not load
        </h1>
        <p className="mt-2 text-sm leading-relaxed text-slate-600">
          The analytics source did not respond. Check your connection and try
          again.
        </p>
        <button
          type="button"
          onClick={reset}
          className="mt-5 inline-flex h-11 items-center justify-center gap-2 rounded-xl bg-brand-500 px-4 text-xs font-semibold text-white shadow-sm transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-2"
        >
          <RefreshCw className="h-4 w-4" />
          Retry
        </button>
      </section>
    </main>
  );
}
