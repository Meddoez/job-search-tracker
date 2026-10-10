"""Create the initial PostgreSQL schema."""

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision = "0001_initial_jobs_schema"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "jobs",
        sa.Column("id", sa.Integer(), primary_key=True),
        sa.Column("source", sa.String(), nullable=False),
        sa.Column("external_id", sa.String(), nullable=True),
        sa.Column("title", sa.String(), nullable=False),
        sa.Column("company", sa.String(), nullable=True),
        sa.Column("location", sa.String(), nullable=True),
        sa.Column("url", sa.String(), nullable=True),
        sa.Column("description", sa.Text(), nullable=True),
        sa.Column("date_posted", sa.String(), nullable=True),
        sa.Column("is_remote", sa.Boolean(), nullable=True),
        sa.Column("raw_data", postgresql.JSONB(), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
    )
    op.create_index(
        "idx_jobs_source_external_id",
        "jobs",
        ["source", "external_id"],
        unique=True,
        postgresql_where=sa.text("external_id IS NOT NULL"),
    )
    op.create_table(
        "job_skills",
        sa.Column("job_id", sa.Integer(), nullable=False),
        sa.Column("skill", sa.String(), nullable=False),
        sa.Column("position", sa.Integer(), nullable=False),
        sa.Column("skill_type", sa.String(), server_default="mentioned", nullable=False),
        sa.Column("match_source", sa.String(), server_default="job", nullable=False),
        sa.ForeignKeyConstraint(["job_id"], ["jobs.id"], ondelete="CASCADE"),
        sa.UniqueConstraint("job_id", "skill", name="uq_job_skills_job_id_skill"),
    )
    op.create_index("idx_job_skills_skill", "job_skills", ["skill"], unique=False)
    op.create_table(
        "refresh_metadata",
        sa.Column("id", sa.Integer(), primary_key=True),
        sa.Column("query", sa.String(), nullable=False),
        sa.Column("refreshed_at", sa.DateTime(timezone=True), nullable=False),
        sa.CheckConstraint("id = 1", name="ck_refresh_metadata_singleton"),
    )


def downgrade() -> None:
    op.drop_table("refresh_metadata")
    op.drop_index("idx_job_skills_skill", table_name="job_skills")
    op.drop_table("job_skills")
    op.drop_index("idx_jobs_source_external_id", table_name="jobs")
    op.drop_table("jobs")
