import type { Job } from "../../services/jobs"

type JobsListProps = {
  jobs: Job[]
  isSearching: boolean
  hasSearched: boolean
  onSaveJob: (job: Job) => void
  canSave: boolean
  savingJobId: string | null
}

function JobsSkeleton() {
  return (
    <div className="grid gap-3" aria-hidden="true">
      {[1, 2, 3].map((item) => (
        <div
          key={item}
          className="animate-pulse rounded-xl border border-slate-800 bg-slate-900/60 p-5"
        >
          <div className="h-4 w-2/3 rounded bg-slate-800" />
          <div className="mt-3 h-3 w-1/4 rounded bg-slate-800/70" />
        </div>
      ))}
    </div>
  )
}

export function JobsList({
  jobs,
  isSearching,
  hasSearched,
  onSaveJob,
  canSave,
  savingJobId,
}: JobsListProps) {
  return (
    <>
      <div className="mt-8 flex items-end justify-between gap-4 border-b border-slate-800 pb-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
            JobTech
          </p>
          <h2 className="mt-1 text-xl font-semibold text-white">
            Search results
            {hasSearched && (
              <span className="ml-2 text-base font-normal text-slate-500">
                {jobs.length}
              </span>
            )}
          </h2>
        </div>
      </div>

      <div className="mt-4">
        {isSearching && !hasSearched ? (
          <JobsSkeleton />
        ) : jobs.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-slate-700 bg-slate-900/40 px-6 py-12 text-center">
            <p className="font-medium text-slate-200">
              {hasSearched ? "No matching jobs" : "Search for jobs"}
            </p>
            <p className="mx-auto mt-2 max-w-md text-sm leading-6 text-slate-500">
              {hasSearched
                ? "Try a different role or keyword."
                : "Enter a role or keyword above to find JobTech postings."}
            </p>
          </div>
        ) : jobs.length > 0 ? (
          <div
            className={`grid gap-3 transition-opacity ${isSearching ? "opacity-60" : "opacity-100"}`}
          >
            {jobs.map((job, index) => (
              <article
                key={`${job.source}:${job.external_id ?? `${job.title}:${index}`}`}
                className="job-card group flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="min-w-0">
                  <h3 className="font-medium leading-6 text-slate-100">
                    {job.title}
                  </h3>
                  <p className="mt-1 text-xs text-slate-500">
                    {job.company || "Unknown company"}
                  </p>
                </div>
                <div className="flex flex-wrap items-center gap-2 self-start sm:self-auto">
                  {job.url ? (
                    <a
                      href={job.url}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex shrink-0 items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-xs font-semibold text-slate-300 transition group-hover:border-sky-400/40 group-hover:text-sky-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400"
                    >
                      View posting
                    </a>
                  ) : (
                    <span className="shrink-0 text-xs text-slate-600">
                      Link unavailable
                    </span>
                  )}
                  <button
                    type="button"
                    onClick={() => onSaveJob(job)}
                    disabled={
                      !canSave || !job.external_id || savingJobId !== null
                    }
                    className="button-secondary group-hover:border-sky-400/40 group-hover:text-sky-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {savingJobId !== null && savingJobId === job.external_id
                      ? "Saving..."
                      : "Save posting"}
                  </button>
                </div>
              </article>
            ))}
          </div>
        ) : null}
      </div>
    </>
  )
}
