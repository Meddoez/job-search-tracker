type JobsFeedbackProps = {
  isSearching: boolean
  error: string | null
  notice: string | null
}

export function JobsFeedback({
  isSearching,
  error,
  notice,
}: JobsFeedbackProps) {
  return (
    <>
      {isSearching && (
        <div
          role="status"
          className="flex items-center gap-3 border-b border-slate-800 bg-sky-400/5 px-5 py-4 text-sm text-sky-200 sm:px-7"
        >
          <span className="size-4 shrink-0 animate-spin rounded-full border-2 border-sky-300/30 border-t-sky-300" />
          Searching JobTech...
        </div>
      )}

      {error && (
        <div className="border-b border-slate-800 px-5 py-4 sm:px-7">
          <div
            role="alert"
            className="flex items-start gap-3 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm leading-6 text-rose-200"
          >
            <span aria-hidden="true" className="font-semibold">
              !
            </span>
            <p>{error}</p>
          </div>
        </div>
      )}

      {notice && (
        <p
          role="status"
          className="border-b border-slate-800 bg-emerald-400/5 px-5 py-4 text-sm text-emerald-300 sm:px-7"
        >
          {notice}
        </p>
      )}
    </>
  )
}
