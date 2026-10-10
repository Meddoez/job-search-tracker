"""Require profile names to be unique after normalization."""

from alembic import op
import sqlalchemy as sa

revision = "0004_unique_profile_names"
down_revision = "0003_profiles"
branch_labels = None
depends_on = None


def upgrade() -> None:
    """Prevent profile names that differ only by case or surrounding whitespace."""
    op.create_index(
        "uq_profiles_name_normalized",
        "profiles",
        [sa.text("lower(btrim(name))")],
        unique=True,
    )


def downgrade() -> None:
    """Allow duplicate normalized profile names again."""
    op.drop_index("uq_profiles_name_normalized", table_name="profiles")
