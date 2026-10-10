import type { ReactNode } from "react"

type FormFieldProps = {
  label: string
  htmlFor: string
  hint?: string
  children: ReactNode
}

export function FormField({ label, htmlFor, hint, children }: FormFieldProps) {
  return (
    <div>
      <label htmlFor={htmlFor} className="text-sm font-medium text-slate-200">
        {label}
      </label>
      <div className="mt-2">{children}</div>
      {hint && <p className="mt-2 text-xs leading-5 text-slate-500">{hint}</p>}
    </div>
  )
}
