"""Internal HTTP API for transient JobTech searches."""

from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel, Field, field_validator

from job_search_tracker.ingestion.jobtech import fetch_jobs
from job_search_tracker.models import Job


app = FastAPI()


class SearchRequest(BaseModel):
    query: str
    limit: int = Field(default=50, gt=0, le=100)

    @field_validator("query")
    @classmethod
    def normalize_query(cls, value: str) -> str:
        query = value.strip()
        if not query:
            raise ValueError("query must not be blank")
        return query


class JobResponse(BaseModel):
    source: str
    external_id: str | None
    title: str
    company: str | None
    location: str | None
    url: str | None
    description: str | None
    skills: list[str]
    date_posted: str | None
    is_remote: bool | None

    @classmethod
    def from_job(cls, job: Job) -> "JobResponse":
        return cls(
            source=job.source,
            external_id=job.external_id,
            title=job.title,
            company=job.company,
            location=job.location,
            url=job.url,
            description=job.description,
            skills=list(job.skills),
            date_posted=job.date_posted,
            is_remote=job.is_remote,
        )


class SearchResponse(BaseModel):
    count: int
    jobs: list[JobResponse]


@app.get("/health")
def health() -> dict[str, str]:
    """Return readiness for container health checks."""
    return {"status": "ok"}


@app.post("/search", response_model=SearchResponse)
def search(request: SearchRequest) -> SearchResponse:
    """Return normalized JobTech postings without persisting them."""
    try:
        jobs = fetch_jobs(query=request.query, limit=request.limit)
    except ConnectionError as exc:
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail="JobTech request failed",
        ) from exc
    except ValueError as exc:
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail="JobTech returned an invalid response",
        ) from exc

    response_jobs = [JobResponse.from_job(job) for job in jobs]
    return SearchResponse(count=len(response_jobs), jobs=response_jobs)
