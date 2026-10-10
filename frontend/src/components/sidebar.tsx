import { Link, NavLink } from "react-router"
import { Profile } from "../services/profile"

type SidebarProps = {
  isOpen: boolean
  profiles: Profile[]
}

export function Sidebar({ isOpen, profiles }: SidebarProps) {
  const navigation: { label: string; to?: string }[] = [
    { label: "Home", to: "/" },
    { label: "Jobs", to: "/jobs" },
    { label: "Profile", to: "/profile" },
  ]

  return (
    <aside
      id="sidebar"
      className={`fixed top-16 bottom-0 left-0 z-10 w-64 flex-col border-r border-slate-800 bg-slate-950 text-slate-100 flex transition-transform duration-200 ease-out ${
        isOpen ? "translate-x-0" : "-translate-x-full"
      }`}
    >
      <div className="flex items-center gap-3 px-5 py-5 md:px-6 md:py-7">
        <div className="grid size-10 shrink-0 place-items-center rounded-xl border border-sky-400/30 bg-sky-400/10 text-sm font-bold tracking-tight text-sky-300">
          JT
        </div>
        <div>
          <Link to="/" className="font-semibold tracking-tight text-white">
            Job Search Tracker
          </Link>
        </div>
      </div>

      <nav
        aria-label="Primary navigation"
        className="flex flex-col gap-1 px-3 md:px-4"
      >
        {navigation.map((item, index) =>
          item.to ? (
            <NavLink
              key={item.label}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `sidebar-nav-item ${
                  isActive
                    ? "sidebar-nav-item-active"
                    : "text-slate-400 transition-colors hover:bg-slate-900 hover:text-slate-100"
                }`
              }
            >
              <span className="font-mono text-[10px] tracking-wider opacity-60">
                {String(index + 1).padStart(2, "0")}
              </span>
              <span>{item.label}</span>
            </NavLink>
          ) : (
            <div
              key={item.label}
              aria-disabled="true"
              className="sidebar-nav-item sidebar-nav-item-disabled"
            >
              <span className="font-mono text-[10px] tracking-wider opacity-60">
                {String(index + 1).padStart(2, "0")}
              </span>
              <span>{item.label}</span>
              <span className="ml-auto hidden text-[10px] font-normal uppercase tracking-wider text-slate-600 md:inline">
                Soon
              </span>
            </div>
          ),
        )}
      </nav>

      {profiles.length > 0 && (
        <nav
          aria-label="Profiles navigation"
          className="relative mr-3 ml-7 space-y-1 pl-3 before:absolute before:top-0 before:bottom-[18px] before:left-0 before:border-l before:border-slate-700/80 md:mr-4 md:ml-8"
        >
          {profiles.map((profile) => (
            <NavLink
              key={profile.id}
              to={`/profiles/${profile.id}`}
              className={({ isActive }) =>
                `relative flex min-w-0 items-center rounded-lg px-3 py-2 text-sm font-medium transition-colors before:absolute before:top-1/2 before:-left-3 before:w-3 before:border-t before:border-slate-700/80 ${
                  isActive
                    ? "bg-sky-400/10 text-sky-300 before:border-sky-400/50"
                    : "text-slate-400 hover:bg-slate-900 hover:text-slate-100"
                }`
              }
            >
              <span className="truncate">{profile.name}</span>
            </NavLink>
          ))}
        </nav>
      )}

      <div className="mt-auto hidden border-t border-slate-800 p-5 md:block">
        <div className="flex items-center gap-2 text-xs text-slate-500">
          <span className="size-1.5 rounded-full bg-emerald-400" />
          Local workspace
        </div>
      </div>
    </aside>
  )
}
