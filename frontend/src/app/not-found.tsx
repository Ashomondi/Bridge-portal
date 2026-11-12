import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#edf2f1] px-4">
      <div className="w-full max-w-md rounded-[28px] border border-slate-200 bg-white p-8 text-center shadow-[0_20px_60px_rgba(15,23,42,0.08)]">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-[#0d6e66] text-2xl font-bold text-white">
          I
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900">Page not found</h1>
        <p className="mt-3 text-sm text-slate-600">
          The page you requested could not be found.
        </p>
        <Link
          href="/login"
          className="mt-6 inline-flex rounded-xl bg-[#0d6e66] px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-[#0b5d58]"
        >
          Return to sign in
        </Link>
      </div>
    </div>
  );
}
