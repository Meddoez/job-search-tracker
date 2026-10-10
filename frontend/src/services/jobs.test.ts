import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { saveJobs, searchJobs, type Job, type SearchJobsResult } from "./jobs"

const jobs: Job[] = [
  {
    source: "jobtech",
    external_id: "job-1",
    title: "Backend Developer",
    company: "Example AB",
    location: "Stockholm",
    url: "https://example.com/jobs/1",
    description: "Build APIs",
    skills: ["python"],
    date_posted: "2026-09-26T10:00:00",
    is_remote: false,
  },
]

const result: SearchJobsResult = { count: 1, jobs }
const fetchMock = vi.fn<typeof fetch>()

describe("saveJobs", () => {
  it("posts to the chosen profile and reads the backend's profile_job response", async () => {
    const saved = {
      profile_id: 12,
      job_id: 3,
      status: "saved",
      saved_at: "2026-10-10T10:00:00Z",
      updated_at: "2026-10-10T10:00:00Z",
    }
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ profile_job: saved }), { status: 201 }),
    )

    await expect(saveJobs("12", jobs[0])).resolves.toEqual(saved)
    expect(fetchMock).toHaveBeenCalledWith("/api/profiles/12/jobs", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(jobs[0]),
    })
  })
})

beforeEach(() => {
  fetchMock.mockReset()
  vi.stubGlobal("fetch", fetchMock)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe("searchJobs", () => {
  it("posts the query and search limit to the API", async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify(result), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    )

    await expect(searchJobs("backend developer")).resolves.toEqual(result)
    expect(fetchMock).toHaveBeenCalledWith("/api/jobs/search", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query: "backend developer", limit: 50 }),
    })
  })

  it("surfaces the API error detail", async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ detail: "Query is not supported" }), {
        status: 422,
        headers: { "Content-Type": "application/json" },
      }),
    )

    await expect(searchJobs("backend")).rejects.toThrow(
      "Query is not supported",
    )
  })

  it("falls back to the response status for a non-JSON error", async () => {
    fetchMock.mockResolvedValue(
      new Response("service unavailable", { status: 503 }),
    )

    await expect(searchJobs("backend")).rejects.toThrow(
      "Could not search jobs (503)",
    )
  })
})
