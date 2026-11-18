<<<<<<< HEAD
const quickStats = [
	{ label: 'Overall performance', value: 'B+' },
	{ label: 'Attendance', value: '96%' },
	{ label: 'Fees due', value: 'KSh 12,500' },
];

export default function ParentPage() {
	return (
		<main className="min-h-screen bg-slate-950 p-6 text-slate-100">
			<div className="mx-auto max-w-6xl">
				<header className="mb-8 border-b border-slate-800 pb-6">
					<p className="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-400">
						Parent
					</p>
					<h1 className="mt-2 text-3xl font-bold">Student dashboard</h1>
				</header>

				<section className="grid gap-4 md:grid-cols-3">
					{quickStats.map((stat) => (
						<div
							key={stat.label}
							className="rounded-2xl border border-slate-800 bg-slate-900 p-5"
						>
							<p className="text-sm text-slate-400">{stat.label}</p>
							<p className="mt-3 text-3xl font-bold text-white">
								{stat.value}
							</p>
						</div>
					))}
				</section>

				<section className="mt-8 grid gap-6 lg:grid-cols-[1.5fr_1fr]">
					<div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
						<h2 className="text-xl font-semibold">Recent academic updates</h2>
						<ul className="mt-6 space-y-4 text-sm text-slate-300">
							<li className="flex items-center justify-between border-b border-slate-800 pb-3">
								<span>Mathematics</span>
								<span className="text-emerald-400">A-</span>
							</li>
							<li className="flex items-center justify-between border-b border-slate-800 pb-3">
								<span>English</span>
								<span className="text-emerald-400">B+</span>
							</li>
							<li className="flex items-center justify-between">
								<span>Science</span>
								<span className="text-emerald-400">A</span>
							</li>
						</ul>
					</div>

					<div className="rounded-3xl border border-slate-800 bg-slate-900 p-6">
						<h2 className="text-xl font-semibold">Assistant</h2>
						<div className="mt-6 rounded-2xl bg-slate-800 p-4 text-sm text-slate-300">
							“Your child has no pending disciplinary concerns. Fee payment
							reminder sent for this month.”
						</div>
					</div>
				</section>
			</div>
		</main>
	);
=======
import ParentHomeDashboard from "@/components/parent-home/ParentHomeDashboard";
import { loadParentHome } from "@/components/parent-home/parent-home-source";

export const dynamic = "force-dynamic";

export default async function ParentHomePage({
  searchParams,
}: {
  searchParams: { studentId?: string | string[] };
}) {
  const data = await loadParentHome({
    studentId:
      typeof searchParams.studentId === "string"
        ? searchParams.studentId
        : undefined,
  });

  return <ParentHomeDashboard data={data} />;
>>>>>>> 7e300ec074132fac5cb72ca2428a933bfcbc5172
}
