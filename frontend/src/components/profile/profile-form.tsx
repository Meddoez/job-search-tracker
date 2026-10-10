import { FormField } from "../ui/form-field"
import { Panel } from "../ui/panel"
import { createProfile } from "../../services/profile"
import { useState, type ChangeEvent } from "react"

interface ProfileCreateFormProps {
  inputName: string
  inputSkills: string[]
  onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

export function ProfileForm({
  inputName,
  inputSkills,
}: ProfileCreateFormProps) {
  const [name, setName] = useState(inputName)
  const [skills, setSkills] = useState(inputSkills.join(","))

  return (
    <section className="mt-8" aria-labelledby="profile-form-title">
      <Panel>
        <div className="bg-gradient-to-br from-sky-400/10 via-slate-950 to-slate-950 p-5 sm:p-7">
          <div className="flex items-start gap-4">
            <div className="grid size-11 shrink-0 place-items-center rounded-xl border border-sky-400/20 bg-sky-400/10 text-sky-300">
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.75"
                className="size-5"
              >
                <circle cx="12" cy="8" r="3.5" />
                <path strokeLinecap="round" d="M5 20c.6-4 3-6 7-6s6.4 2 7 6" />
              </svg>
            </div>
            <div>
              <h2
                id="profile-form-title"
                className="text-lg font-semibold text-white"
              >
                Create new profile
              </h2>
              <p className="mt-1 max-w-2xl text-sm leading-6 text-slate-400">
                Give your profile a name and list the skills you already have.
              </p>
            </div>
          </div>

          <div className="mt-6 grid gap-5">
            <FormField label="Profile name" htmlFor="profile-name">
              <input
                id="profile-name"
                name="name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                type="text"
                autoComplete="off"
                placeholder="e.g. Backend developer"
                className="form-input w-full"
              />
            </FormField>

            <FormField
              label="Skills"
              htmlFor="profile-skills"
              hint="Separate each skill with a comma."
            >
              <input
                id="profile-skills"
                name="skills"
                value={skills}
                onChange={(event) => setSkills(event.target.value)}
                type="text"
                autoComplete="off"
                placeholder="e.g. Python, PostgreSQL, Docker"
                className="form-input w-full"
              />
            </FormField>
            <button
              onClick={() => {
                const typedSkills = skills
                  .split(",")
                  .map((skill) => skill.trim())
                  .filter(Boolean)
                createProfile(name, typedSkills)
              }}
              type="button"
              className="button-primary justify-self-start"
            >
              Create profile
            </button>
          </div>
        </div>
      </Panel>
    </section>
  )
}
