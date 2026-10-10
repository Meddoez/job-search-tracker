"""Integration tests for profile-owned saved-job constraints."""

import pytest
from sqlalchemy import create_engine, delete, func, insert, select
from sqlalchemy.exc import IntegrityError

from job_search_tracker.storage.metadata import jobs, profile_jobs, profiles


pytestmark = pytest.mark.integration


def test_profile_jobs_enforce_ownership_and_delete_rules(clean_database):
    engine = create_engine(clean_database)
    with engine.begin() as connection:
        first_profile_id = connection.execute(
            insert(profiles).values(name="Backend").returning(profiles.c.id)
        ).scalar_one()
        second_profile_id = connection.execute(
            insert(profiles).values(name="Platform").returning(profiles.c.id)
        ).scalar_one()
        job_id = connection.execute(
            insert(jobs)
            .values(source="jobtech", external_id="job-1", title="Backend Developer")
            .returning(jobs.c.id)
        ).scalar_one()
        connection.execute(insert(profile_jobs).values(profile_id=first_profile_id, job_id=job_id))
        connection.execute(insert(profile_jobs).values(profile_id=second_profile_id, job_id=job_id))

    with pytest.raises(IntegrityError):
        with engine.begin() as connection:
            connection.execute(insert(profile_jobs).values(profile_id=first_profile_id, job_id=job_id))

    with engine.begin() as connection:
        connection.execute(delete(profiles).where(profiles.c.id == first_profile_id))
        relationship_count = connection.execute(
            select(func.count()).select_from(profile_jobs)
        ).scalar_one()
    assert relationship_count == 1

    with pytest.raises(IntegrityError):
        with engine.begin() as connection:
            connection.execute(delete(jobs).where(jobs.c.id == job_id))

    engine.dispose()


def test_profile_jobs_reject_invalid_status(clean_database):
    engine = create_engine(clean_database)
    with engine.begin() as connection:
        profile_id = connection.execute(
            insert(profiles).values(name="Backend").returning(profiles.c.id)
        ).scalar_one()
        job_id = connection.execute(
            insert(jobs)
            .values(source="jobtech", external_id="job-1", title="Backend Developer")
            .returning(jobs.c.id)
        ).scalar_one()

    with pytest.raises(IntegrityError):
        with engine.begin() as connection:
            connection.execute(
                insert(profile_jobs).values(
                    profile_id=profile_id,
                    job_id=job_id,
                    status="unknown",
                )
            )

    engine.dispose()
