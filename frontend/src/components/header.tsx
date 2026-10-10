type HeaderProps = {
  isSidebarOpen: boolean
  onMenuClick: () => void
}

export function Header({ isSidebarOpen, onMenuClick }: HeaderProps) {
  return (
    <header className="sticky top-0 z-20 flex h-16 w-full items-center justify-between border-b border-slate-800 bg-slate-900/95 px-5 backdrop-blur md:px-10">
      <button
        type="button"
        aria-controls="sidebar"
        aria-expanded={isSidebarOpen}
        onClick={onMenuClick}
        className="inline-flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 px-3.5 py-2 text-sm font-medium text-slate-300 transition-colors hover:border-slate-600 hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400"
      >
        <svg
          aria-hidden="true"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.75"
          className="size-4"
        >
          <path strokeLinecap="round" d="M4 7h16M4 12h16M4 17h16" />
        </svg>
        Menu
      </button>

      <div className="flex items-center gap-2 text-xs text-slate-500">
        <span className="size-1.5 rounded-full bg-emerald-400" />
        Local workspace
      </div>
    </header>
  )
}
