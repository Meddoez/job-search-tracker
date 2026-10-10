"""Remove fields used only by the retired persistent refresh workflow."""

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql


revision = "0005_transient_search"
down_revision = "0004_unique_profile_names"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.drop_table("refresh_metadata")
    op.drop_column("jobs", "raw_data")


def downgrade() -> None:
    op.add_column("jobs", sa.Column("raw_data", postgresql.JSONB(), nullable=True))
    op.create_table(
        "refresh_metadata",
        sa.Column("id", sa.Integer(), primary_key=True),
        sa.Column("source", sa.String(), nullable=False),
        sa.Column("query", sa.String(), nullable=False),
        sa.Column("refreshed_at", sa.DateTime(timezone=True), nullable=False),
        sa.UniqueConstraint("source", name="uq_refresh_metadata_source"),
    )
