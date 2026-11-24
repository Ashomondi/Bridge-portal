const stats = [
  { label: 'Visits today', value: '18' },
  { label: 'Recovery rates', value: '92%' },
  { label: 'Pending alerts', value: '4' },
];

export default function NursePage() {
  return (
    <main className="min-h-screen bg-slate-950 p-6 text-slate-100">
      <div className="mx-auto max-w-6xl">
        <header className="mb-8 border-b border-slate-800 pb-6">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-400">
            Nurse
            <span className="ml-2 rounded-full border border-amber-400/30 bg-amber-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-amber-300">Preview</span>
          </p>
          <h1 className="mt-2 text-3xl font-bold">Sickbay overview</h1>
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
            <h2 className="text-xl font-semibold">Consultation log</h2>
            <ul className="mt-6 space-y-4 text-sm text-slate-300">
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>Headache follow-up</span><span className="text-emerald-400">Resolved</span></li>
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>Medication review</span><span className="text-emerald-400">Scheduled</span></li>
              <li className="flex items-center justify-between"><span>Fevers observed</span><span className="text-emerald-400">2</span></li>
            </ul>
          </div>

          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Alerts</h2>
            <div className="mt-6 space-y-3 text-sm text-slate-300">
              <div className="rounded-2xl bg-slate-800 p-3">Asthma review for 3 students</div>
              <div className="rounded-2xl bg-slate-800 p-3">2 parents requested follow-up calls</div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
