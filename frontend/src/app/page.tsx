import Link from 'next/link';

const roles = [
	{
		name: 'Admin',
		href: '/admin',
		description: 'School operations and approvals',
	},
	{
		name: 'Parent',
		href: '/parent',
		description: 'Student progress and payments',
	},
	{
		name: 'Teacher',
		href: '/teacher',
		description: 'Classroom delivery and attendance',
	},
	{
		name: 'Nurse',
		href: '/nurse',
		description: 'Health and wellbeing tracking',
	},
	{
		name: 'Librarian',
		href: '/librarian',
		description: 'Books, loans and returns',
	},
	{
		name: 'Bursar',
		href: '/bursar',
		description: 'Fee records and reconciliation',
	},
];

export default function HomePage() {
	return (
		<main className="min-h-screen bg-slate-950 text-slate-100">
			<div className="mx-auto max-w-6xl px-6 py-16">
				<header className="mb-12 flex items-center justify-between gap-4 border-b border-slate-800 pb-6">
					<div>
						<p className="text-xs font-semibold uppercase tracking-[0.25em] text-emerald-400">
							Ikonex Academy
						</p>
						<h1 className="mt-2 text-3xl font-bold">
							Institutional Portal
						</h1>
					</div>
					<Link
						href="/login"
						className="rounded-full bg-emerald-500 px-5 py-2.5 text-sm font-semibold text-slate-950 transition hover:bg-emerald-400"
					>
						Sign in
					</Link>
				</header>

				<section className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
					{roles.map((role) => (
						<Link
							key={role.name}
							href={role.href}
							className="group rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/30 transition hover:border-emerald-500/70 hover:bg-slate-900"
						>
							<div className="mb-4 inline-flex rounded-2xl bg-emerald-500/10 px-3 py-2 text-emerald-400">
								{role.name}
							</div>
							<h2 className="text-xl font-semibold text-white">
								{role.name} workspace
							</h2>
							<p className="mt-2 text-sm text-slate-300">
								{role.description}
							</p>
							<div className="mt-6 text-sm font-medium text-emerald-400 transition group-hover:translate-x-1">
								Open dashboard →
							</div>
						</Link>
					))}
				</section>
			</div>
		</main>
	);
}
