/**
 * Static fake dashboard that mirrors the real AdminLayout structure.
 * sidebar (w-56) + header (h-14) + content skeleton.
 * Zero API, zero hooks, zero session — pure decorative background for login page.
 */
export function FakeDashboard() {
  return (
    <div className="flex h-screen select-none bg-background" aria-hidden="true">
      {/* Fake sidebar */}
      <aside className="flex w-56 shrink-0 flex-col border-r border-border bg-card px-3 py-4">
        <div className="mb-6 h-5 w-20 rounded bg-muted" />
        <div className="flex flex-col gap-1.5">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="flex items-center gap-2.5 rounded-md px-3 py-2">
              <div className="h-4 w-4 rounded bg-muted/80" />
              <div className="h-3 rounded bg-muted/60" style={{ width: `${50 + (i % 3) * 20}%` }} />
            </div>
          ))}
        </div>
      </aside>

      {/* Main area */}
      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
        {/* Fake header */}
        <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-card px-8">
          <div className="h-4 w-24 rounded bg-muted" />
          <div className="flex items-center gap-3">
            <div className="h-6 w-6 rounded-full bg-muted" />
            <div className="h-6 w-16 rounded bg-muted" />
          </div>
        </header>

        {/* Fake content */}
        <main className="flex-1 overflow-hidden p-8 opacity-60">
          {/* Stats row */}
          <div className="grid grid-cols-4 gap-4">
            {['12,580', '¥48,200', '23', '1,247'].map((val, i) => (
              <div key={i} className="rounded-lg border border-border bg-card p-4">
                <div className="h-3 w-16 rounded bg-muted/70" />
                <div className="mt-2 h-5 w-20 rounded bg-muted/50" />
              </div>
            ))}
          </div>

          {/* Fake chart */}
          <div className="mt-6 rounded-lg border border-border bg-card p-6">
            <div className="mb-4 h-4 w-28 rounded bg-muted/70" />
            <div className="flex h-32 items-end gap-1">
              {[35, 50, 42, 65, 58, 72, 68, 80, 75, 90, 85, 78, 82, 95, 88, 70, 76, 83, 91, 86].map(
                (h, i) => (
                  <div
                    key={i}
                    className="flex-1 rounded-sm bg-primary/20"
                    style={{ height: `${h}%` }}
                  />
                ),
              )}
            </div>
          </div>

          {/* Fake table */}
          <div className="mt-6 rounded-lg border border-border bg-card p-6">
            <div className="mb-4 h-4 w-20 rounded bg-muted/70" />
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="mb-3 flex items-center gap-4">
                <div className="h-3 w-1/5 rounded bg-muted/50" />
                <div className="h-3 w-1/4 rounded bg-muted/40" />
                <div className="h-3 w-1/6 rounded bg-muted/30" />
                <div className="h-3 w-12 rounded bg-primary/15" />
              </div>
            ))}
          </div>
        </main>
      </div>
    </div>
  )
}
