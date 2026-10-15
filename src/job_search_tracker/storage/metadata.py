"""Canonical SQLAlchemy metadata for the job tracking schema."""

from __future__ import annotations

from sqlalchemy import (
    Boolean,
    Column,
    DateTime,
    ForeignKey,
    Index,
    Integer,
    MetaData,
    String,
    Table,
    Text,
    UniqueConstraint,
    func,
    CheckConstraint,
)

metadata = MetaData()

jobs = Table(
    "jobs",
    metadata,
    Column("id", Integer, primary_key=True),
    Column("source", String, nullable=False),
    Column("external_id", String),
    Column("title", String, nullable=False),
    Column("company", String),
    Column("location", String),
    Column("url", String),
    Column("description", Text),
    Column("date_posted", String),
    Column("is_remote", Boolean),
    Column("created_at", DateTime(timezone=True), nullable=False, server_default=func.now()),
)

Index(
    "idx_jobs_source_external_id",
    jobs.c.source,
    jobs.c.external_id,
    unique=True,
    postgresql_where=jobs.c.external_id.is_not(None),
)

job_skills = Table(
    # Position preserves matcher order for stable stored output.
    "job_skills",
    metadata,
    Column("job_id", Integer, ForeignKey("jobs.id", ondelete="CASCADE"), nullable=False),
    Column("skill", String, nullable=False),
    Column("position", Integer, nullable=False),
    Column("skill_type", String, nullable=False, server_default="mentioned"),
    Column("match_source", String, nullable=False, server_default="job"),
    UniqueConstraint("job_id", "skill", name="uq_job_skills_job_id_skill"),
)

Index("idx_job_skills_skill", job_skills.c.skill)

profiles = Table(
    "profiles",
    metadata,
    Column("id", Integer, primary_key=True),
    Column("name", Text, nullable=False),
    Column("created_at", DateTime(timezone=True), nullable=False, server_default=func.now()),
    Column("updated_at", DateTime(timezone=True), nullable=False, server_default=func.now()),
    CheckConstraint("btrim(name) <> ''", name="ck_profiles_name_not_blank"),
)

profile_skills = Table(
    "profile_skills",
    metadata,
    Column("profile_id", Integer, ForeignKey("profiles.id", ondelete="CASCADE"), primary_key=True),
    Column("skill", String, primary_key=True),
    CheckConstraint("btrim(skill) <> ''", name="ck_profile_skills_skill_not_blank"),
)

profile_jobs = Table(
    "profile_jobs",
    metadata,
    Column("profile_id", Integer, ForeignKey("profiles.id", ondelete="CASCADE"), primary_key=True),
    Column("job_id", Integer, ForeignKey("jobs.id", ondelete="RESTRICT"), primary_key=True),
    Column("status", String, nullable=False, server_default="saved"),
    Column("saved_at", DateTime(timezone=True), nullable=False, server_default=func.now()),
    Column("updated_at", DateTime(timezone=True), nullable=False, server_default=func.now()),
    Column("notes", Text, nullable=True),
    CheckConstraint(
        "status IN ('saved', 'applied', 'interviewing', 'offer', 'rejected', 'withdrawn')",
        name="ck_profile_jobs_status",
    ),
)
