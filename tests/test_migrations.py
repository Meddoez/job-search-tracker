"""Integration tests for the PostgreSQL Alembic schema."""

from pathlib import Path

import pytest
from alembic.config import Config
from alembic.script import ScriptDirectory
from sqlalchemy import create_engine, inspect, text

pytestmark = pytest.mark.integration


def test_database_is_at_expected_alembic_revision(clean_database):
    repo_root = Path(__file__).resolve().parents[1]
    config = Config(str(repo_root / "alembic.ini"))
    config.set_main_option("script_location", str(repo_root / "alembic"))
    expected_revisions = set(ScriptDirectory.from_config(config).get_heads())
    engine = create_engine(clean_database)
    try:
        with engine.connect() as connection:
            revisions = set(connection.execute(text("SELECT version_num FROM alembic_version")).scalars())
        assert revisions == expected_revisions
    finally:
        engine.dispose()


def test_postgresql_schema_has_expected_constraints(clean_database):
    engine = create_engine(clean_database)
    inspector = inspect(engine)

    assert {column["name"] for column in inspector.get_columns("jobs")} >= {
        "id",
        "source",
        "external_id",
        "is_remote",
    }
    assert "raw_data" not in {column["name"] for column in inspector.get_columns("jobs")}
    assert {index["name"] for index in inspector.get_indexes("jobs")} >= {
        "idx_jobs_source_external_id",
    }
    assert {constraint["name"] for constraint in inspector.get_unique_constraints("job_skills")} >= {
        "uq_job_skills_job_id_skill",
    }
    assert not inspector.has_table("refresh_metadata")
    assert {index["name"] for index in inspector.get_indexes("profiles")} >= {
        "uq_profiles_name_normalized",
    }
    assert inspector.get_foreign_keys("job_skills")[0]["options"]["ondelete"] == "CASCADE"
    assert inspector.get_pk_constraint("profile_jobs")["constrained_columns"] == [
        "profile_id",
        "job_id",
    ]
    profile_job_foreign_keys = {
        foreign_key["constrained_columns"][0]: foreign_key["options"]["ondelete"]
        for foreign_key in inspector.get_foreign_keys("profile_jobs")
    }
    assert profile_job_foreign_keys == {"profile_id": "CASCADE", "job_id": "RESTRICT"}
    assert {constraint["name"] for constraint in inspector.get_check_constraints("profile_jobs")} >= {
        "ck_profile_jobs_status",
    }
    engine.dispose()
