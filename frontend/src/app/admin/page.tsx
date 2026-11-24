const stats = [
  { label: 'Active staff', value: '184' },
  { label: 'Students', value: '1,260' },
  { label: 'Pending approvals', value: '12' },
];

// Admin portal page
export default function AdminPage() {
  return (
    <main className="min-h-screen bg-slate-950 p-6 text-slate-100">
      <div className="mx-auto max-w-6xl">
        <header className="mb-8 border-b border-slate-800 pb-6">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-400">
              Admin
              <span className="ml-2 rounded-full border border-amber-400/30 bg-amber-400/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-amber-300">Preview</span>
            </p>
            <h1 className="mt-2 text-3xl font-bold">School Operations</h1>
          </div>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          {stats.map((stat) => (
            <div key={stat.label} className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
              <p className="text-sm text-slate-400">{stat.label}</p>
              <p className="mt-3 text-3xl font-bold text-white">{stat.value}</p>
            </div>
          ))}
        </section>

        <section className="mt-8 grid gap-6 lg:grid-cols-[1.5fr_1fr]">
          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Recent activity</h2>
            <ul className="mt-6 space-y-4 text-sm text-slate-300">
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>New student admissions</span><span className="text-emerald-400">24</span></li>
              <li className="flex items-center justify-between border-b border-slate-800 pb-3"><span>Staff onboarding QR codes</span><span className="text-emerald-400">8</span></li>
              <li className="flex items-center justify-between"><span>Compliance checks</span><span className="text-emerald-400">Completed</span></li>
            </ul>
          </div>

          <div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
            <h2 className="text-xl font-semibold">Priority tasks</h2>
            <div className="mt-6 space-y-3 text-sm text-slate-300">
              <div className="rounded-2xl bg-slate-800 p-3">Approve onboarding requests</div>
              <div className="rounded-2xl bg-slate-800 p-3">Review staff access updates</div>
              <div className="rounded-2xl bg-slate-800 p-3">Monitor fee reconciliation</div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
