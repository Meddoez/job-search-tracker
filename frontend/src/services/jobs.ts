import { SavedJobUpdate } from "../pages/ProfileDetails"

export type Job = {
  source: string
  external_id: string | null
  title: string
  company: string | null
  location: string | null
  url: string | null
  description: string | null
  skills: string[]
  date_posted: string | null
  is_remote: boolean | null
}

export type SavedProfileJob = {
  job_id: number
  job: Job
  status: string
  notes: string | null
  saved_at: string
  updated_at: string
}

export type SearchJobsResult = {
  count: number
  jobs: Job[]
}

export type SaveJobResult = {
  profile_id: number
  job_id: number
  status: string
  saved_at: string
  updated_at: string
}

async function getErrorMessage(response: Response, fallback: string) {
  try {
    const data: unknown = await response.json()
    if (typeof data === "object" && data !== null) {
      if ("detail" in data && typeof data.detail === "string") {
        return data.detail
      }
      if ("message" in data && typeof data.message === "string") {
        return data.message
      }
    }
  } catch {
    // Fall back to the response status when the body is not JSON.
  }

  return `${fallback} (${response.status})`
}

export async function searchJobs(query: string): Promise<SearchJobsResult> {
  const response = await fetch("/api/jobs/search", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query, limit: 50 }),
  })

  if (!response.ok) {
    throw new Error(await getErrorMessage(response, "Could not search jobs"))
  }

  return response.json()
}

// We need to show if we refresh and press savejob again that this job is already saved or something similar
export async function saveJobs(
  profile_id: string,
  job: Job,
): Promise<SaveJobResult> {
  const response = await fetch(`/api/profiles/${profile_id}/jobs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(job),
  })

  if (!response.ok) {
    console.log("Could not save job")
    throw new Error(await getErrorMessage(response, "Could not save job"))
  }

  const data: { profile_job: SaveJobResult } = await response.json()
  console.log("Saved job")
  return data.profile_job
}

export async function getSavedJobs(
  profile_id: number,
): Promise<SavedProfileJob[]> {
  const response = await fetch(`/api/profiles/${profile_id}/jobs`)
  if (!response.ok) {
    console.log("Could not get saved Jobs")
    throw new Error(await getErrorMessage(response, "Could not get saved Jobs"))
  }

  const data: { profile_job: SavedProfileJob[] } = await response.json()
  console.log("Saved jobs fetched")

  return data.profile_job
}

export async function updateSavedJob(
  profile_id: number,
  job_id: number,
  update: SavedJobUpdate,
): Promise<SavedProfileJob> {
  const response = await fetch(`/api/profiles/${profile_id}/jobs/${job_id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(update),
  })

  if (!response.ok) {
    console.log("Could not update profile job")
    throw new Error(
      await getErrorMessage(response, "Could not update profile job"),
    )
  }
  const data: { profile_job: SavedProfileJob } = await response.json()
  console.log("Profile job updated")
  return data.profile_job
}
