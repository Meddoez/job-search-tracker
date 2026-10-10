"""Create profile-owned saved jobs."""

from alembic import op
import sqlalchemy as sa


revision = "0006_profile_jobs"
down_revision = "0005_transient_search"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "profile_jobs",
        sa.Column("profile_id", sa.Integer(), primary_key=True),
        sa.Column("job_id", sa.Integer(), primary_key=True),
        sa.Column("status", sa.String(), server_default="saved", nullable=False),
        sa.Column("saved_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.ForeignKeyConstraint(["profile_id"], ["profiles.id"], ondelete="CASCADE"),
        sa.ForeignKeyConstraint(["job_id"], ["jobs.id"], ondelete="RESTRICT"),
        sa.CheckConstraint(
            "status IN ('saved', 'applied', 'interviewing', 'offer', 'rejected', 'withdrawn')",
            name="ck_profile_jobs_status",
        ),
    )


def downgrade() -> None:
    op.drop_table("profile_jobs")
