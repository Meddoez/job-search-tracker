import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { MemoryRouter, Route, Routes } from "react-router"
import { beforeEach, expect, it, vi } from "vitest"

import type { SavedProfileJob } from "../services/jobs"
import type { Profile } from "../services/profile"
import { ProfileDetailsPage } from "./ProfileDetails"

const { getProfileMock, getSavedJobsMock } = vi.hoisted(() => ({
  getProfileMock: vi.fn<(id: number) => Promise<Profile>>(),
  getSavedJobsMock: vi.fn<(id: number) => Promise<SavedProfileJob[]>>(),
}))

vi.mock("../services/profile", () => ({ getProfile: getProfileMock }))
vi.mock("../services/jobs", () => ({ getSavedJobs: getSavedJobsMock }))

beforeEach(() => {
  getProfileMock.mockReset()
  getSavedJobsMock.mockReset()
  getSavedJobsMock.mockResolvedValue([])
})

it("shows the profile and an empty jobs workspace with backend status views", async () => {
  getProfileMock.mockResolvedValue({
    id: 7,
    name: "Backend Developer",
    skills: ["go"],
  })

  render(
    <MemoryRouter initialEntries={["/profiles/7"]}>
      <Routes>
        <Route path="/profiles/:id" element={<ProfileDetailsPage />} />
      </Routes>
    </MemoryRouter>,
  )

  expect(
    await screen.findByRole("heading", { name: "Backend Developer" }),
  ).toBeInTheDocument()
  expect(getProfileMock).toHaveBeenCalledWith(7)
  expect(screen.getByText("go")).toBeInTheDocument()
  for (const status of [
    "All jobs",
    "saved",
    "applied",
    "interviewing",
    "offer",
    "rejected",
    "withdrawn",
  ]) {
    expect(screen.getByRole("button", { name: status })).toBeInTheDocument()
  }
  expect(
    await screen.findByRole("heading", { name: "No saved jobs yet" }),
  ).toBeInTheDocument()
  expect(getSavedJobsMock).toHaveBeenCalledWith(7)
  expect(screen.getByRole("link", { name: "Browse jobs" })).toHaveAttribute(
    "href",
    "/jobs",
  )
})

it("filters fetched jobs by status and returns to all jobs", async () => {
  const user = userEvent.setup()
  getProfileMock.mockResolvedValue({
    id: 7,
    name: "Backend Developer",
    skills: [],
  })
  const job = (id: number, title: string, status: string): SavedProfileJob => ({
    job_id: id,
    job: {
      source: "jobtech",
      external_id: String(id),
      title,
      company: null,
      location: null,
      url: null,
      description: null,
      skills: [],
      date_posted: null,
      is_remote: null,
    },
    status,
    saved_at: "2026-10-10T12:00:00Z",
    updated_at: "2026-10-10T12:00:00Z",
  })
  getSavedJobsMock.mockResolvedValue([
    job(1, "Go Developer", "saved"),
    job(2, "React Developer", "applied"),
  ])

  render(
    <MemoryRouter initialEntries={["/profiles/7"]}>
      <Routes>
        <Route path="/profiles/:id" element={<ProfileDetailsPage />} />
      </Routes>
    </MemoryRouter>,
  )

  expect(
    await screen.findByRole("heading", { name: "Go Developer" }),
  ).toBeInTheDocument()
  expect(
    screen.getByRole("heading", { name: "React Developer" }),
  ).toBeInTheDocument()

  await user.click(screen.getByRole("button", { name: /^saved$/ }))
  expect(
    screen.getByRole("heading", { name: "Go Developer" }),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole("heading", { name: "React Developer" }),
  ).not.toBeInTheDocument()
  expect(screen.getByRole("button", { name: /^saved$/ })).toHaveAttribute(
    "aria-pressed",
    "true",
  )

  await user.click(screen.getByRole("button", { name: /^interviewing$/ }))
  expect(
    screen.getByRole("heading", { name: "No jobs with this status" }),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole("heading", { name: "No saved jobs yet" }),
  ).not.toBeInTheDocument()

  await user.click(screen.getByRole("button", { name: "All jobs" }))
  expect(
    screen.getByRole("heading", { name: "Go Developer" }),
  ).toBeInTheDocument()
  expect(
    screen.getByRole("heading", { name: "React Developer" }),
  ).toBeInTheDocument()
  expect(getSavedJobsMock).toHaveBeenCalledTimes(1)
})

it("does not show the jobs empty state if the profile cannot be loaded", async () => {
  getProfileMock.mockRejectedValue(new Error("Profile not found"))

  render(
    <MemoryRouter initialEntries={["/profiles/7"]}>
      <Routes>
        <Route path="/profiles/:id" element={<ProfileDetailsPage />} />
      </Routes>
    </MemoryRouter>,
  )

  expect(
    await screen.findByRole("heading", { name: "Profile unavailable" }),
  ).toBeInTheDocument()
  expect(
    screen.queryByRole("heading", { name: "No saved jobs yet" }),
  ).not.toBeInTheDocument()
})
