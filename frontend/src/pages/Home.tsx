import { Link } from "react-router"

const steps = [
  {
    title: "Create a profile",
    description:
      "Create a profile to keep your saved jobs and applications together.",
  },
  {
    title: "Find and save jobs",
    description:
      "Search JobTech for a role or keyword, then save listings to your profile.",
  },
  {
    title: "Track your applications",
    description:
      "Update each job's status as you apply, interview, and hear back.",
  },
]

export function Home() {
  return (
    <>
      <header className="border-b border-slate-800 pb-7">
        <p className="page-eyebrow">Workspace</p>
        <h1 className="page-title">Job Search &amp; Application Tracker</h1>
        <p className="page-description">
          Find jobs, save the ones you're interested in, and keep track of where
          you are in each application process.
        </p>
      </header>

      <section className="mt-8" aria-labelledby="getting-started-title">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
              First steps
            </p>
            <h2
              id="getting-started-title"
              className="mt-1 text-xl font-semibold text-white"
            >
              Getting started
            </h2>
          </div>
          <Link to="/jobs" className="button-primary self-start sm:self-auto">
            Open Jobs
          </Link>
        </div>

        <ol className="mt-5 grid gap-3 md:grid-cols-3">
          {steps.map((step, index) => (
            <li
              key={step.title}
              className="rounded-xl border border-slate-800 bg-slate-950 p-5"
            >
              <span className="font-mono text-xs text-sky-400">
                {String(index + 1).padStart(2, "0")}
              </span>
              <h3 className="mt-4 font-semibold text-slate-100">
                {step.title}
              </h3>
              <p className="mt-2 text-sm leading-6 text-slate-400">
                {step.description}
              </p>
            </li>
          ))}
        </ol>
      </section>
    </>
  )
}
