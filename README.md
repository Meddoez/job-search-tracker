# Job Search & Application Tracker

I'm building this to make job searching easier to keep track of. Finding a job posting is only the start, you also need to remember where you applied, which applications are moving forward, and what happened along the way.

The idea is to keep that in one place: jobs you're interested in, where you are in the application process, and notes you can come back to. After an interview or a rejection, I want somewhere to write down what went well, what I struggled with, and what I could improve before the next application.

The skill analysis side supports that too. Seeing which skills appear in job postings can help you figure out what to work on next.

## What works today

- Search job postings through JobTech, the Swedish public employment service's job API.
- Create named profiles with your skills.
- Save jobs to a profile so you can come back to them later.
- Track each job as **saved**, **applied**, **interviewing**, **offer**, **rejected**, or **withdrawn**.
- Filter saved jobs by their application status.
- Extract technical skills from job descriptions using dictionary-based matching.

This is a work in progress. Job saving and status tracking are implemented. Job-specific notes and reflections are the next part of the idea.

## Run locally

### Prerequisites

- Docker with Docker Compose
- Make
- Go 1.27 or newer (`make up` runs `go mod tidy` locally)
- Git

### Start the app

Clone the repository and enter the project directory:

```bash
git clone https://github.com/Meddoez/job-search-tracker.git
cd job-search-tracker
```

Copy the example environment configuration:

```bash
cp .env.example .env
```

Set `POSTGRES_PASSWORD` in `.env` and use the same password in `DATABASE_URL` and `TEST_DATABASE_URL`. The database hostname in both URLs should stay as `db` when running with Compose.

Start the application:

```bash
make up
```

Compose builds the services, starts PostgreSQL, and runs database migrations before starting the Go API. The first build may take a few minutes.

Open **http://localhost:5173**. The Go API is available at **http://localhost:8080**, and PostgreSQL is exposed on port **55432** by default. You can change these using `FRONTEND_PORT`, `API_PORT`, and `POSTGRES_PORT` in `.env`.

### Try it out

1. Create a profile and add your skills.
2. Go to the Jobs page and search for a role or keyword.
3. Choose a profile and save a posting to it.
4. Open that profile to view your saved jobs, update their statuses, and filter by stage.

Search terms can be separated by commas or spaces, and every term is required. Common English and Swedish role terms are treated as equivalents. For example, `developer, C++, junior` requires `C++`, `junior`, and one of `developer`, `utvecklare`, or `systemutvecklare`.

Search results are temporary. Jobs are stored in PostgreSQL when you save them to a profile.

### Stop or reset

Stop the services while keeping your saved data:

```bash
make down
```

Stop the services and delete the database volume, including profiles and saved jobs:

```bash
make reset
```

## How it's built

| Part | Technology | Responsibility |
| --- | --- | --- |
| Frontend | React, TypeScript, Vite, Tailwind CSS | Job searches, profiles, and application tracking |
| API | Go, Gin | Profile management, saved jobs, and status updates |
| Search service | Python, FastAPI | Fetching JobTech postings, normalizing results, and extracting skills |
| Database | PostgreSQL, Alembic | Persistent profiles and saved jobs, with versioned schema migrations |
| Local environment | Docker Compose | Running the services together |

The frontend sends searches through the Go API to the Python service. The Python service queries JobTech and returns postings in a shared format with extracted skills. The Go API handles saving jobs and tracking their status in PostgreSQL.

Skill extraction currently uses a dictionary-based matcher. Comparing job requirements against profile skills is an area I want to build on.

## Tests

For the full suite, you also need Node.js 24 with npm and frontend dependencies installed:

```bash
npm --prefix frontend ci
make test
```

This runs Go tests with the race detector, frontend tests and a production build, then database migrations and Python tests against a dedicated `job_search_tracker_test` database. The test database is recreated each time; the application database is not modified by test cleanup.

The example environment file includes `TEST_DATABASE_URL`. Keep its credentials consistent with the application database and its database name as `job_search_tracker_test`.

For local Python development, install Python 3.11 or newer and set up the virtual environment:

```bash
./scripts/bootstrap.sh
.venv/bin/python -m pytest -q
```

PostgreSQL-backed Python tests require `TEST_DATABASE_URL`; `make test` configures this inside Compose.

## What's next
- Add notes to each job for interview questions, feedback, and follow-ups.
- Record what went well, what I struggled with, and what to practice before the next interview.
- Compare job requirements with profile skills to spot gaps.
- Highlight skills that keep appearing across saved jobs, so it's easier to decide what to learn next.
