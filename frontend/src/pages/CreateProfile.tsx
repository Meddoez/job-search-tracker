import { ProfileForm } from "../components/profile/profile-form"
import { PageHeader } from "../components/ui/page-header"
import { Profile } from "../services/profile"

type ProfilePageProps = {
  onProfileCreated: (profile: Profile) => void
}

export function CreateProfilePage({ onProfileCreated }: ProfilePageProps) {
  return (
    <>
      <PageHeader
        eyebrow="Workspace"
        title="Your profile"
        description="Create a profile to organize your saved jobs and track your applications."
      />
      <ProfileForm inputName="" inputSkills={[]} onChange={() => {}} />
    </>
  )
}
