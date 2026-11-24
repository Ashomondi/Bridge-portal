const stats = [
  { label: 'Books issued', value: '438' },
  { label: 'Overdue', value: '17' },
  { label: 'Catalogued', value: '8,290' },
];

export default function LibrarianPage() {
  return (
    <main className="min-h-screen bg-slate-950 p-6 text-slate-100">
      <div className="mx-auto max-w-6xl">
        <header className="mb-8 border-b border-slate-800 pb-6">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-400">
            Librarian
            <span className="ml-2 rounded-full border border-amber-400/30 bg-amber-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-amber-300">Preview</span>
          </p>
          <h1 className="mt-2 text-3xl font-bold">Library operations</h1>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          {stats.map((item) => (
            <div key={item.label} className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
              <p className="text-sm text-slate-400">{item.label}</p>
              <p className="mt-3 text-3xl font-bold text-white">{item.value}</p>
            </div>
          ))}
        </section>

        <section className="mt-8 grid gap-6 lg:grid-cols-[1.5fr_1fr]">
          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Loan activity</h2>
            <ul className="mt-6 space-y-4 text-sm text-slate-300">
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>History of Kenya</span><span className="text-emerald-400">Due in 2 days</span></li>
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>Chemistry practicals</span><span className="text-emerald-400">Checked out</span></li>
              <li className="flex items-center justify-between"><span>Mathematics revision guide</span><span className="text-emerald-400">Returned</span></li>
            </ul>
          </div>

          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Needs attention</h2>
            <div className="mt-6 space-y-3 text-sm text-slate-300">
              <div className="rounded-2xl bg-slate-800 p-3">7 books require catalog verification</div>
              <div className="rounded-2xl bg-slate-800 p-3">2 students have overdue loans</div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
