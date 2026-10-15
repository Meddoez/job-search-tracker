"""Add notes column to profile_jobs table."""

from alembic import op
import sqlalchemy as sa

revision = "0007_notes"
down_revision = "0006_profile_jobs"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.add_column("profile_jobs", sa.Column("notes", sa.Text(), nullable=True))

def downgrade() -> None:
    op.drop_column("profile_jobs", "notes")
