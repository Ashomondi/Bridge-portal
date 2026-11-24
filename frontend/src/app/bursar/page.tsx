const summary = [
  { label: 'Collected this term', value: 'KSh 2.4M' },
  { label: 'Outstanding', value: 'KSh 410K' },
  { label: 'Reconciled', value: '94%' },
];

export default function BursarPage() {
  return (
    <main className="min-h-screen bg-slate-950 p-6 text-slate-100">
      <div className="mx-auto max-w-6xl">
        <header className="mb-8 border-b border-slate-800 pb-6">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-400">
            Bursar
            <span className="ml-2 rounded-full border border-amber-400/30 bg-amber-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-amber-300">Preview</span>
          </p>
          <h1 className="mt-2 text-3xl font-bold">Fee and reconciliation</h1>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          {summary.map((item) => (
            <div key={item.label} className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
              <p className="text-sm text-slate-400">{item.label}</p>
              <p className="mt-3 text-3xl font-bold text-white">{item.value}</p>
            </div>
          ))}
        </section>

        <section className="mt-8 grid gap-6 lg:grid-cols-[1.5fr_1fr]">
          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Recent transactions</h2>
            <ul className="mt-6 space-y-4 text-sm text-slate-300">
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>STK settlement</span><span className="text-emerald-400">KSh 56,000</span></li>
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>Cash deposit</span><span className="text-emerald-400">KSh 18,500</span></li>
              <li className="flex items-center justify-between"><span>Balance transfer</span><span className="text-emerald-400">KSh 12,000</span></li>
            </ul>
          </div>

          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Follow-up</h2>
            <div className="mt-6 space-y-3 text-sm text-slate-300">
              <div className="rounded-2xl bg-slate-800 p-3">12 fee reminders are due this week</div>
              <div className="rounded-2xl bg-slate-800 p-3">2 reconciliation mismatches need review</div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
