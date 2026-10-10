"""Tests for skill matching."""

import pytest

from job_search_tracker.extraction.matcher import extract_skills, extract_skills_from_texts


def test_extract_skills_returns_all_mentioned_skills():
    text = "Python, Django, Git, Docker"

    assert extract_skills(text) == ["python", "git", "docker", "django"]


def test_extract_skills_from_texts_combines_sources():
    assert extract_skills_from_texts(("Build APIs", "Use Docker and AWS")) == [
        "apis",
        "docker",
        "aws",
    ]


def test_extract_skills_detects_specific_employer_tools():
    text = "We use AWS, Docker, Kubernetes, PostgreSQL, MySQL, MongoDB and Redis."

    assert extract_skills(text) == [
        "docker",
        "kubernetes",
        "postgresql",
        "mysql",
        "mongodb",
        "redis",
        "aws",
    ]


def test_extract_skills_handles_jobtech_taxonomy_labels():
    text = "Python, programmeringsspråk. Amazon Web Services (AWS), molnplattform."

    assert extract_skills(text) == ["python", "aws"]


def test_extract_skills_handles_symbol_heavy_skills():
    text = "Build services with Node.js, C# and .NET."

    assert extract_skills(text) == ["c#", ".net", "node.js"]


def test_extract_skills_does_not_match_terms_inside_other_words():
    text = "The latest JavaScript shellfish testing framework uses Node.js."

    assert extract_skills(text) == ["node.js", "javascript", "testing"]


def test_extract_skills_handles_data_tools_seen_in_jobtech_postings():
    text = "Use SQL Server, Power BI, Pandas, NumPy, Matplotlib and SQLAlchemy."

    assert extract_skills(text) == [
        "sql",
        "sql server",
        "pandas",
        "numpy",
        "matplotlib",
        "sqlalchemy",
        "power bi",
    ]


@pytest.mark.parametrize("text", ["", "   ", "Enjoy working with a friendly team."])
def test_extract_skills_returns_empty_when_no_skills_are_present(text):
    assert extract_skills(text) == []


def test_extract_skills_deduplicates_repeated_mentions_and_aliases():
    assert extract_skills_from_texts(("Postgres and PostgreSQL", "POSTGRES and Python, python")) == [
        "python",
        "postgresql",
    ]


def test_extract_skills_is_case_insensitive():
    assert extract_skills("pYtHoN, DOCKER, PostgreSQL") == ["python", "docker", "postgresql"]


@pytest.mark.parametrize("alias, canonical", [("cpp", "c++"), ("k8s", "kubernetes"), ("postgres", "postgresql")])
def test_extract_skills_returns_canonical_names_for_aliases(alias, canonical):
    assert extract_skills(alias) == [canonical]
