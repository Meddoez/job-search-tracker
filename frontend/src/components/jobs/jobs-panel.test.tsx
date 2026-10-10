import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"

import type { Job, SearchJobsResult } from "../../services/jobs"
import { JobsPanel } from "./jobs-panel"

const { searchJobsMock, saveJobsMock, getProfilesMock } = vi.hoisted(() => ({
  searchJobsMock: vi.fn<(query: string) => Promise<SearchJobsResult>>(),
  saveJobsMock: vi.fn(),
  getProfilesMock: vi.fn(),
}))

vi.mock("../../services/jobs", () => ({
  searchJobs: searchJobsMock,
  saveJobs: saveJobsMock,
}))

vi.mock("../../services/profile", () => ({
  getProfiles: getProfilesMock,
}))

const jobs: Job[] = [
  {
    source: "jobtech",
    external_id: "job-1",
    title: "Go Developer",
    company: "Example AB",
    location: "Stockholm",
    url: "https://example.com/jobs/1",
    description: "Build Go services",
    skills: ["git"],
    date_posted: null,
    is_remote: false,
  },
]

beforeEach(() => {
  searchJobsMock.mockReset()
  saveJobsMock.mockReset()
  getProfilesMock.mockReset()
  getProfilesMock.mockResolvedValue([
    { id: 12, name: "Backend roles", skills: ["go"] },
    { id: 27, name: "Other roles", skills: [] },
  ])
})

describe("JobsPanel", () => {
  it("starts with a search prompt", async () => {
    render(<JobsPanel />)

    expect(screen.getByText("Search for jobs")).toBeInTheDocument()
    expect(searchJobsMock).not.toHaveBeenCalled()
    await screen.findByRole("option", { name: "Backend roles" })
  })

  it("rejects a whitespace-only search query", async () => {
    const user = userEvent.setup()
    render(<JobsPanel />)

    await user.type(
      screen.getByRole("textbox", { name: "Role or keyword" }),
      "   ",
    )
    await user.click(screen.getByRole("button", { name: "Search jobs" }))

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Enter a role or keyword before searching jobs.",
    )
    expect(searchJobsMock).not.toHaveBeenCalled()
  })

  it("renders transient jobs returned by a search", async () => {
    const user = userEvent.setup()
    searchJobsMock.mockResolvedValue({ count: 1, jobs })
    render(<JobsPanel />)

    await user.type(
      screen.getByRole("textbox", { name: "Role or keyword" }),
      "  go developer  ",
    )
    await user.click(screen.getByRole("button", { name: "Search jobs" }))

    expect(
      await screen.findByText('Found 1 job for "go developer".'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole("heading", { name: "Go Developer" }),
    ).toBeInTheDocument()
    expect(screen.getByText("Example AB")).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "View posting" })).toHaveAttribute(
      "href",
      "https://example.com/jobs/1",
    )
    expect(searchJobsMock).toHaveBeenCalledWith("go developer")
  })

  it("shows an empty state when a search has no matches", async () => {
    const user = userEvent.setup()
    searchJobsMock.mockResolvedValue({ count: 0, jobs: [] })
    render(<JobsPanel />)

    await user.type(
      screen.getByRole("textbox", { name: "Role or keyword" }),
      "rust",
    )
    await user.click(screen.getByRole("button", { name: "Search jobs" }))

    expect(await screen.findByText("No matching jobs")).toBeInTheDocument()
  })

  it("keeps existing results when a later search fails", async () => {
    const user = userEvent.setup()
    searchJobsMock
      .mockResolvedValueOnce({ count: 1, jobs })
      .mockRejectedValueOnce(new Error("Job search service unavailable"))
    render(<JobsPanel />)

    const input = screen.getByRole("textbox", { name: "Role or keyword" })
    await user.type(input, "go")
    await user.click(screen.getByRole("button", { name: "Search jobs" }))
    await screen.findByRole("heading", { name: "Go Developer" })

    await user.clear(input)
    await user.type(input, "python")
    await user.click(screen.getByRole("button", { name: "Search jobs" }))

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Job search service unavailable",
    )
    expect(
      screen.getByRole("heading", { name: "Go Developer" }),
    ).toBeInTheDocument()
  })

  it("requires a profile and saves to the selected profile, including after switching", async () => {
    const user = userEvent.setup()
    searchJobsMock.mockResolvedValue({ count: 1, jobs })
    saveJobsMock.mockResolvedValue({})
    render(<JobsPanel />)

    await user.type(
      screen.getByRole("textbox", { name: "Role or keyword" }),
      "go",
    )
    await user.click(screen.getByRole("button", { name: "Search jobs" }))
    const saveButton = await screen.findByRole("button", {
      name: "Save posting",
    })
    expect(saveButton).toBeDisabled()
    expect(saveJobsMock).not.toHaveBeenCalled()

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Save jobs to profile" }),
      "12",
    )
    await user.click(saveButton)
    expect(saveJobsMock).toHaveBeenCalledWith("12", jobs[0])
    expect(
      await screen.findByText('Saved "Go Developer" to Backend roles.'),
    ).toBeInTheDocument()

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Save jobs to profile" }),
      "27",
    )
    await user.click(saveButton)
    expect(saveJobsMock).toHaveBeenLastCalledWith("27", jobs[0])
    expect(
      await screen.findByText('Saved "Go Developer" to Other roles.'),
    ).toBeInTheDocument()
  })

  it("reports save failures and allows retrying", async () => {
    const user = userEvent.setup()
    searchJobsMock.mockResolvedValue({ count: 1, jobs })
    saveJobsMock
      .mockRejectedValueOnce(new Error("Failed to save profile job"))
      .mockResolvedValueOnce({})
    render(<JobsPanel />)

    await screen.findByRole("option", { name: "Backend roles" })
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Save jobs to profile" }),
      "12",
    )
    await user.type(
      screen.getByRole("textbox", { name: "Role or keyword" }),
      "go",
    )
    await user.click(screen.getByRole("button", { name: "Search jobs" }))
    await user.click(
      await screen.findByRole("button", { name: "Save posting" }),
    )
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Failed to save profile job",
    )
    expect(
      screen.queryByText('Saved "Go Developer" to Backend roles.'),
    ).not.toBeInTheDocument()

    await user.click(screen.getByRole("button", { name: "Save posting" }))
    expect(
      await screen.findByText('Saved "Go Developer" to Backend roles.'),
    ).toBeInTheDocument()
    expect(screen.queryByRole("alert")).not.toBeInTheDocument()
  })

  it("offers profile creation when no profiles exist, without blocking searches", async () => {
    const user = userEvent.setup()
    getProfilesMock.mockResolvedValue([])
    searchJobsMock.mockResolvedValue({ count: 1, jobs })
    render(<JobsPanel />)

    expect(
      await screen.findByRole("link", { name: "Create a profile" }),
    ).toHaveAttribute("href", "/profile")
    expect(screen.getByRole("combobox")).toBeDisabled()
    await user.type(
      screen.getByRole("textbox", { name: "Role or keyword" }),
      "go",
    )
    await user.click(screen.getByRole("button", { name: "Search jobs" }))
    expect(
      await screen.findByRole("button", { name: "Save posting" }),
    ).toBeDisabled()
    expect(
      screen.getByRole("heading", { name: "Go Developer" }),
    ).toBeInTheDocument()
  })

  it("shows profile-loading failures", async () => {
    getProfilesMock.mockRejectedValue(new Error("Could not get profiles"))
    render(<JobsPanel />)

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Could not get profiles",
    )
    expect(screen.getByRole("combobox")).toBeDisabled()
  })
})
