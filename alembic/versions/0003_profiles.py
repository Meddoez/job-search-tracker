"""Create profiles and their skills."""

from alembic import op
import sqlalchemy as sa

revision = "0003_profiles"
down_revision = "0002_source_refresh_metadata"
branch_labels = None
depends_on = None


def upgrade() -> None:
    """Create profile tables with timestamp defaults and skill constraints."""
    op.create_table(
        "profiles",
        sa.Column("id", sa.Integer(), primary_key=True),
        sa.Column("name", sa.Text(), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.Column("updated_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.CheckConstraint("btrim(name) <> ''", name="ck_profiles_name_not_blank"),
    )
    op.create_table(
        "profile_skills",
        sa.Column("profile_id", sa.Integer(), primary_key=True),
        sa.Column("skill", sa.String(), primary_key=True),
        sa.ForeignKeyConstraint(["profile_id"], ["profiles.id"], ondelete="CASCADE"),
        sa.CheckConstraint("btrim(skill) <> ''", name="ck_profile_skills_skill_not_blank"),
    )


def downgrade() -> None:
    """Drop profile skills before the profiles they reference."""
    op.drop_table("profile_skills")
    op.drop_table("profiles")
