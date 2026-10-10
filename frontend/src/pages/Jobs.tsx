import { JobsPanel } from "../components/jobs/jobs-panel"
import { PageHeader } from "../components/ui/page-header"

export function Jobs() {
  return (
    <>
      <PageHeader
        eyebrow="Workspace"
        title="Find jobs"
        description="Search current JobTech listings and save the jobs you want to follow up on."
      />
      <JobsPanel />
    </>
  )
}
