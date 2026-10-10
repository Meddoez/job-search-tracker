import type { ReactNode } from "react"

type PanelProps = {
  children: ReactNode
}

export function Panel({ children }: PanelProps) {
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-700/80 bg-slate-950 shadow-2xl shadow-slate-950/30">
      {children}
    </div>
  )
}
