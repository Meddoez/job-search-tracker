import { useEffect, useState, type SubmitEvent } from "react"

import { searchJobs, type Job } from "../../services/jobs"
import { Panel } from "../ui/panel"
import { JobsFeedback } from "./jobs-feedback"
import { JobsList } from "./jobs-list"
import { JobsSearchForm } from "./jobs-search-form"
import { saveJobs } from "../../services/jobs"
import { getProfiles, type Profile } from "../../services/profile"

type RequestState = "idle" | "searching"

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback
}

export function JobsPanel() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [query, setQuery] = useState("")
  const [requestState, setRequestState] = useState<RequestState>("idle")
  const [hasSearched, setHasSearched] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const isBusy = requestState !== "idle"
  const [profiles, setProfiles] = useState<Profile[] | null>(null)
  const [profileError, setProfileError] = useState<string | null>(null)
  const [selectedProfileId, setSelectedProfileId] = useState("")
  const [savingJobId, setSavingJobId] = useState<string | null>(null)

  useEffect(() => {
    let active = true
    getProfiles()
      .then((result) => {
        if (active) setProfiles(result)
      })
      .catch((error) => {
        if (active)
          setProfileError(errorMessage(error, "Could not load profiles"))
      })
    return () => {
      active = false
    }
  }, [])

  const handleSubmit = async (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmedQuery = query.trim()

    if (!trimmedQuery) {
      setError("Enter a role or keyword before searching jobs.")
      return
    }

    setRequestState("searching")
    setError(null)
    setNotice(null)

    try {
      const result = await searchJobs(trimmedQuery)
      setJobs(result.jobs)
      setHasSearched(true)
      setNotice(
        `Found ${result.count} ${result.count === 1 ? "job" : "jobs"} for "${trimmedQuery}".`,
      )
    } catch (error) {
      setError(errorMessage(error, "Could not search jobs"))
    } finally {
      setRequestState("idle")
    }
  }
  const handleSaveJob = async (job: Job) => {
    const profile = profiles?.find(
      (item) => String(item.id) === selectedProfileId,
    )
    if (!job.external_id || !profile || savingJobId) return
    setSavingJobId(job.external_id)
    setError(null)
    setNotice(null)
    try {
      await saveJobs(selectedProfileId, job)
      setNotice(`Saved "${job.title}" to ${profile.name}.`)
    } catch (error) {
      setError(errorMessage(error, "Could not save job"))
    } finally {
      setSavingJobId(null)
    }
  }

  return (
    <section
      className="mt-8"
      aria-labelledby="jobs-panel-title"
      aria-busy={isBusy}
    >
      <Panel>
        <JobsSearchForm
          query={query}
          isBusy={isBusy}
          isSearching={requestState === "searching"}
          onQueryChange={setQuery}
          onSubmit={handleSubmit}
        />
        <JobsFeedback
          isSearching={requestState === "searching"}
          error={error}
          notice={notice}
        />
      </Panel>

      <div className="mt-6 rounded-xl border border-slate-800 bg-slate-950/60 p-5">
        <label
          htmlFor="save-profile"
          className="block text-sm font-medium text-slate-200"
        >
          Save jobs to profile
        </label>
        <select
          id="save-profile"
          value={selectedProfileId}
          onChange={(event) => {
            setSelectedProfileId(event.target.value)
            setNotice(null)
          }}
          disabled={!profiles?.length || savingJobId !== null}
          className="mt-2 w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-slate-200 disabled:opacity-50"
        >
          <option value="">Choose a profile</option>
          {profiles?.map((profile) => (
            <option key={profile.id} value={profile.id}>
              {profile.name}
            </option>
          ))}
        </select>
        {profileError ? (
          <p role="alert" className="mt-2 text-sm text-rose-300">
            {profileError}
          </p>
        ) : profiles === null ? (
          <p role="status" className="mt-2 text-sm text-slate-400">
            Loading profiles...
          </p>
        ) : profiles.length === 0 ? (
          <p className="mt-2 text-sm text-slate-400">
            <a href="/profile" className="text-sky-300 underline">
              Create a profile
            </a>{" "}
            to save jobs. You can still search and browse postings.
          </p>
        ) : (
          <p className="mt-2 text-sm text-slate-400">
            Choose where to save your jobs before saving a posting.
          </p>
        )}
      </div>

      <JobsList
        jobs={jobs}
        isSearching={requestState === "searching"}
        hasSearched={hasSearched}
        onSaveJob={handleSaveJob}
        canSave={Boolean(selectedProfileId) && !profileError}
        savingJobId={savingJobId}
      />
    </section>
  )
}
