const summary = [
  { label: 'Classes', value: '6' },
  { label: 'Attendance', value: '94%' },
  { label: 'Coverage', value: '82%' },
];

export default function TeacherPage() {
  return (
    <main className="min-h-screen bg-slate-950 p-6 text-slate-100">
      <div className="mx-auto max-w-6xl">
        <header className="mb-8 border-b border-slate-800 pb-6">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-400">Teacher</p>
          <h1 className="mt-2 text-3xl font-bold">Classroom overview</h1>
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
            <h2 className="text-xl font-semibold">Today&apos;s timetable</h2>
            <ul className="mt-6 space-y-4 text-sm text-slate-300">
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>08:00 Mathematics</span><span className="text-emerald-400">Form 2A</span></li>
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>09:30 English</span><span className="text-emerald-400">Form 2B</span></li>
              <li className="flex items-center justify-between"><span>11:00 Biology</span><span className="text-emerald-400">Form 3C</span></li>
            </ul>
          </div>

          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Student notes</h2>
            <div className="mt-6 space-y-3 text-sm text-slate-300">
              <div className="rounded-2xl bg-slate-800 p-3">3 students need revision in algebra</div>
              <div className="rounded-2xl bg-slate-800 p-3">One absent due to medical check-up</div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
