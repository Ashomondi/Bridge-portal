export default function ClassMeansLoading() {
  return (
    <main className="min-h-screen bg-[#F4F7F6] px-4 py-8 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-[1440px] animate-pulse">
        <div className="h-3 w-28 rounded-full bg-slate-200" />
        <div className="mt-4 h-9 w-56 rounded-xl bg-slate-200" />
        <div className="mt-5 h-20 rounded-2xl bg-white shadow-card" />
        <div className="mt-4 h-16 rounded-2xl bg-white shadow-card" />
        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <div
              key={index}
              className="h-44 rounded-2xl border border-slate-200/80 bg-white shadow-card"
            />
          ))}
        </div>
        <div className="mt-6 h-64 rounded-2xl border border-slate-200/80 bg-white shadow-card" />
      </div>
    </main>
  );
}
