import { useEffect, useState } from "react"
import { Form, Link, useParams } from "react-router"

import { getProfile, type Profile } from "../services/profile"
import { getSavedJobs, type SavedProfileJob } from "../services/jobs"
import { updateSavedJob } from "../services/jobs"

const statusViews = [
  "saved",
  "applied",
  "interviewing",
  "offer",
  "rejected",
  "withdrawn",
] as const

type JobStatus = (typeof statusViews)[number]
type JobView = "all" | JobStatus

export type SavedJobUpdate = {
  status?: JobStatus
  notes?: string
}

function SavedJobsSkeleton() {
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

export function ProfileDetailsPage() {
  const { id } = useParams()
  const profileId = Number(id)
  const [profile, setProfile] = useState<Profile | null>(null)
  const [profileError, setProfileError] = useState(false)
  const [jobs, setSavedJobs] = useState<SavedProfileJob[] | null>(null)
  const [jobsError, setJobsError] = useState(false)
  const [selectedView, setSelectedView] = useState<JobView>("all")

  useEffect(() => {
    let active = true
    if (!Number.isInteger(profileId) || profileId <= 0) {
      return
    }

    getProfile(profileId)
      .then((result) => {
        if (active) {
          setProfile(result)
          setProfileError(false)
        }
      })
      .catch(() => {
        if (active) setProfileError(true)
      })

    getSavedJobs(profileId)
      .then((result) => {
        if (active) {
          setSavedJobs(result)
          setJobsError(false)
        }
      })
      .catch(() => {
        if (active) setJobsError(true)
      })

    return () => {
      active = false
    }
  }, [profileId])

  async function handleStatusChange(jobId: number, update: SavedJobUpdate) {
    try {
      await updateSavedJob(profileId, jobId, update)
      setSavedJobs(
        (prevJobs) =>
          prevJobs?.map((job) =>
            job.job_id === jobId && update.status !== undefined
              ? { ...job, status: update.status }
              : job,
          ) ?? null,
      )
    } catch (error) {
      console.error("Error updating saved job:", error)
    }
  }

  async function HandleNotesSave(jobId: number, update: SavedJobUpdate) {
    try {
      await updateSavedJob(profileId, jobId, update)



    } catch (error) {
      console.error("Error updating saved notes:", error)
    }
  }

  const currentProfile =
    !profileError && profile?.id === profileId ? profile : null
  const unavailable =
    profileError || !Number.isInteger(profileId) || profileId <= 0

  const visibleJobs =
    jobs?.filter(
      (item) => selectedView === "all" || item.status === selectedView,
    ) ?? []

  return (
    <>
      <header className="border-b border-slate-800 pb-7">
        <p className="page-eyebrow">Profile workspace</p>
        <h1 className="page-title">
          {currentProfile?.name ??
            (unavailable ? "Profile unavailable" : "Loading profile...")}
        </h1>
        <p className="page-description">
          {unavailable
            ? "This profile could not be loaded."
            : "Keep track of the opportunities you save for this profile."}
        </p>
        {currentProfile && currentProfile.skills.length > 0 && (
          <div
            className="mt-5 flex flex-wrap items-center gap-2"
            aria-label="Profile skills"
          >
            {currentProfile.skills.map((skill) => (
              <span
                key={skill}
                className="rounded-md border border-slate-700 bg-slate-800/60 px-2.5 py-1 text-xs text-slate-300"
              >
                {skill}
              </span>
            ))}
          </div>
        )}
      </header>

      {currentProfile && !unavailable && (
        <section className="mt-8" aria-labelledby="profile-jobs-title">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                Opportunities
              </p>
              <h2
                id="profile-jobs-title"
                className="mt-1 text-xl font-semibold text-white"
              >
                Your jobs
              </h2>
            </div>
            <p className="text-sm text-slate-500">
              Track each role from save to outcome.
            </p>
          </div>

          <div className="mt-6 overflow-x-auto border-b border-slate-800">
            <div
              className="flex w-max min-w-full items-center gap-1"
              aria-label="Job status views"
            >
              <button
                type="button"
                onClick={() => setSelectedView("all")}
                aria-pressed={selectedView === "all"}
                className={`whitespace-nowrap border-b-2 px-4 py-3 text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 ${
                  selectedView === "all"
                    ? "border-sky-400 font-semibold text-sky-300"
                    : "border-transparent text-slate-500 hover:text-slate-200"
                }`}
              >
                All jobs
              </button>
              {statusViews.map((status) => (
                <button
                  key={status}
                  type="button"
                  onClick={() => setSelectedView(status)}
                  aria-pressed={selectedView === status}
                  className={`whitespace-nowrap border-b-2 px-4 py-3 text-sm font-medium capitalize transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400 ${
                    selectedView === status
                      ? "border-sky-400 font-semibold text-sky-300"
                      : "border-transparent text-slate-500 hover:text-slate-200"
                  }`}
                >
                  {status}
                </button>
              ))}
            </div>
          </div>

          <div className="mt-4">
            {jobsError ? (
              <p role="alert" className="text-sm text-rose-300">
                Could not load saved jobs.
              </p>
            ) : jobs === null ? (
              <SavedJobsSkeleton />
            ) : jobs.length === 0 ? (
              <div className="mt-6 rounded-2xl border border-dashed border-slate-700 bg-slate-950/70 px-6 py-14 text-center sm:py-20">
                <div
                  className="mx-auto mb-5 grid size-12 place-items-center rounded-xl border border-sky-400/20 bg-sky-400/10 text-xl text-sky-300"
                  aria-hidden="true"
                >
                  +
                </div>
                <h3 className="text-lg font-semibold text-white">
                  No saved jobs yet
                </h3>
                <p className="mx-auto mt-2 max-w-sm text-sm leading-6 text-slate-400">
                  Jobs you save for this profile will appear here, ready to
                  organize by application status.
                </p>
                <Link to="/jobs" className="button-primary mt-6">
                  Browse jobs
                </Link>
              </div>
            ) : visibleJobs.length === 0 ? (
              <div className="mt-6 rounded-2xl border border-dashed border-slate-700 bg-slate-950/70 px-6 py-14 text-center sm:py-20">
                <h3 className="text-lg font-semibold text-white">
                  No jobs with this status
                </h3>
                <p className="mt-2 text-sm text-slate-400">
                  Choose another status to see your jobs.
                </p>
              </div>
            ) : (
              visibleJobs.map((savedJob) => (
                <div
                  key={savedJob.job_id}
                  className="rounded-xl border border-slate-800 bg-slate-900/60 p-5"
                >
                  <h3 className="font-semibold text-white">
                    {savedJob.job.title}
                  </h3>
                  {savedJob.job.company && (
                    <p className="mt-1 text-sm text-slate-400">
                      {savedJob.job.company}
                    </p>
                  )}
                  {savedJob.job.url ? (
                    <a
                      href={savedJob.job.url}
                      target="_blank"
                      rel="noreferrer"
                      className="mt-3 ml-auto flex w-fit items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-xs font-semibold text-slate-300 transition group-hover:border-sky-400/40 group-hover:text-sky-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400"
                    >
                      View posting
                    </a>
                  ) : (
                    <span className="shrink-0 text-xs text-slate-600">
                      Link unavailable
                    </span>
                  )}
                  <select
                    value={savedJob.status}
                    onChange={(e) =>
                      handleStatusChange(
                        savedJob.job_id, { 
                        status: e.target.value as JobStatus
                      })
                    }
                    className="mt-3 ml-auto flex w-fit items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-xs font-semibold text-slate-300 transition group-hover:border-sky-400/40 group-hover:text-sky-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400"
                  >
                    {statusViews.map((status) => (
                      <option key={status} value={status}>
                        {status}
                      </option>
                    ))}
                  </select>
                  <form className="mt-3 ml-auto flex w-fit items-center gap-2 rounded-lg border border-slate-700 px-3 py-2 text-xs font-semibold text-slate-300 transition group-hover:border-sky-400/40 group-hover:text-sky-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-400" 
                  >
                    <label htmlFor="notes" className="font-semibold">Notes:</label>
                    <textarea className="bg-white text-black rounded-lg border-slate-700 focus-visible:outline-offset-2 flex"
                      id="notes" 
                      name="notes" 
                      maxLength={100}
                      placeholder="Enter notes here..."
                    ></textarea>
                  </form>
                </div>
              ))
            )}
          </div>
        </section>
      )}
    </>
  )
}
