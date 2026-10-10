import type { SubmitEventHandler } from "react"

import { FormField } from "../ui/form-field"

type JobsSearchFormProps = {
  query: string
  isBusy: boolean
  isSearching: boolean
  onQueryChange: (query: string) => void
  onSubmit: SubmitEventHandler<HTMLFormElement>
}

export function JobsSearchForm({
  query,
  isBusy,
  isSearching,
  onQueryChange,
  onSubmit,
}: JobsSearchFormProps) {
  return (
    <div className="border-b border-slate-800 bg-gradient-to-br from-sky-400/10 via-slate-950 to-slate-950 p-5 sm:p-7">
      <div className="flex items-start gap-4">
        <div className="grid size-11 shrink-0 place-items-center rounded-xl border border-sky-400/20 bg-sky-400/10 text-sky-300">
          <svg
            aria-hidden="true"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.75"
            className="size-5"
          >
            <circle cx="11" cy="11" r="7" />
            <path strokeLinecap="round" d="m16 16 4 4" />
          </svg>
        </div>
        <div>
          <h2
            id="jobs-panel-title"
            className="text-lg font-semibold text-white"
          >
            Search the job market
          </h2>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-slate-400">
            Search JobTech for a role or keyword. Results stay temporary until
            you choose to save them to a profile.
          </p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="mt-6">
        <FormField
          label="Role or keyword"
          htmlFor="job-query"
        >
          <div className="flex flex-col gap-3 sm:flex-row">
            <input
              id="job-query"
              name="query"
              value={query}
              onChange={(event) => onQueryChange(event.target.value)}
              disabled={isBusy}
              required
              autoComplete="off"
              placeholder="e.g. junior backend developer"
              className="form-input min-w-0 flex-1"
            />
            <button type="submit" disabled={isBusy} className="button-primary">
              {isSearching && (
                <span className="size-4 animate-spin rounded-full border-2 border-slate-950/30 border-t-slate-950" />
              )}
              {isSearching ? "Searching jobs..." : "Search jobs"}
            </button>
          </div>
        </FormField>
      </form>
    </div>
  )
}
