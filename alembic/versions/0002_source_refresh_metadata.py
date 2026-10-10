"""Track refresh metadata independently for each source."""

from alembic import op
import sqlalchemy as sa

revision = "0002_source_refresh_metadata"
down_revision = "0001_initial_jobs_schema"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.add_column("refresh_metadata", sa.Column("source", sa.String(), nullable=True))
    op.execute("UPDATE refresh_metadata SET source = 'unknown' WHERE source IS NULL")
    op.alter_column("refresh_metadata", "source", nullable=False)
    op.drop_constraint("ck_refresh_metadata_singleton", "refresh_metadata", type_="check")
    op.create_unique_constraint("uq_refresh_metadata_source", "refresh_metadata", ["source"])


def downgrade() -> None:
    op.drop_constraint("uq_refresh_metadata_source", "refresh_metadata", type_="unique")
    op.execute(
        "DELETE FROM refresh_metadata "
        "WHERE id <> (SELECT MIN(id) FROM refresh_metadata)"
    )
    op.create_check_constraint(
        "ck_refresh_metadata_singleton",
        "refresh_metadata",
        "id = 1",
    )
    op.drop_column("refresh_metadata", "source")
