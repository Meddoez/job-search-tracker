export type Profile = {
  id: number
  name: string
  skills: string[]
}

// Should not use effect, but maybe the display of profiles when entering the profiles page?
// add a useState for error, seterror ?
export async function createProfile(
  name: string,
  skills: string[],
): Promise<Profile> {
  const response = await fetch("/api/profiles", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, skills }),
  })
  if (!response.ok) {
    console.log("Could not create profile")
    throw new Error("Could not create profile")
  }
  const data: { profile: Profile } = await response.json()
  console.log("Profile created")
  return data.profile
}

export async function getProfile(id: number): Promise<Profile> {
  const response = await fetch(`/api/profiles/${id}`)
  if (!response.ok) {
    console.log("Could not get profiles")
    throw new Error("Could not get profiles")
  }
  const data: { profile: Profile } = await response.json()
  console.log("Profiles fetched")
  return data.profile
}

// We need to deduplicate thse get functions
export async function getProfiles(): Promise<Profile[]> {
  const response = await fetch(`/api/profiles`)
  if (!response.ok) {
    console.log("Could not get profiles")
    throw new Error("Could not get profiles")
  }
  const data: { profiles: Profile[] } = await response.json()
  console.log("Profiles fetched")
  return data.profiles
}
