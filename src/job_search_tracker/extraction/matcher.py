"""Dictionary-based technical skill matching for job text."""

from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Iterable


@dataclass(frozen=True, slots=True)
class SkillPattern:
    """Canonical skill name and the aliases used to detect it."""

    canonical: str
    aliases: tuple[str, ...]


SKILL_PATTERNS: tuple[SkillPattern, ...] = (
    SkillPattern("python", ("python", "python 3", "python, programmeringsspråk", "py")),
    SkillPattern("c++", ("c++", "cpp", "c plus plus")),
    SkillPattern("c#", ("c#", "c sharp")),
    SkillPattern(".net", (".net", "dotnet", "asp.net")),
    SkillPattern("git", ("git", "git workflow", "git workflows")),
    SkillPattern("jira", ("jira",)),
    SkillPattern("object-oriented programming", ("oop", "object oriented programming", "object oriented design")),
    SkillPattern("apis", ("api", "apis", "rest api", "rest apis")),
    SkillPattern("containers", ("container", "containers")),
    SkillPattern("docker", ("docker",)),
    SkillPattern("kubernetes", ("kubernetes", "k8s")),
    SkillPattern("databases", ("database", "databases")),
    SkillPattern("sql", ("sql",)),
    SkillPattern("sql server", ("sql server", "microsoft sql server")),
    SkillPattern("postgresql", ("postgresql", "postgres", "postgresql database")),
    SkillPattern("mysql", ("mysql",)),
    SkillPattern("mongodb", ("mongodb", "mongo")),
    SkillPattern("redis", ("redis",)),
    SkillPattern("cloud", ("cloud",)),
    SkillPattern("aws", ("aws", "amazon web services", "amazon web services (aws)", "aws, molnplattform")),
    SkillPattern("azure", ("azure", "microsoft azure", "azure, molnplattform")),
    SkillPattern("gcp", ("gcp", "google cloud", "google cloud platform")),
    SkillPattern("linux / command line", ("linux", "command line", "shell", "terminal", "bash")),
    SkillPattern("java", ("java", "java, programmeringsspråk")),
    SkillPattern("django", ("django",)),
    SkillPattern("flask", ("flask",)),
    SkillPattern("fastapi", ("fastapi",)),
    SkillPattern("pandas", ("pandas",)),
    SkillPattern("numpy", ("numpy",)),
    SkillPattern("matplotlib", ("matplotlib",)),
    SkillPattern("sqlalchemy", ("sqlalchemy",)),
    SkillPattern("power bi", ("power bi", "powerbi")),
    SkillPattern("node.js", ("node.js", "nodejs", "node")),
    SkillPattern("javascript", ("javascript", "javascript, programmeringsspråk", "js")),
    SkillPattern("typescript", ("typescript", "typescript, programmeringsspråk", "ts")),
    SkillPattern("react", ("react", "reactjs", "react.js")),
    SkillPattern("angular", ("angular",)),
    SkillPattern("vue", ("vue", "vue.js", "vuejs")),
    SkillPattern("spring", ("spring", "spring boot")),
    SkillPattern("testing", ("testing", "test", "pytest", "unit test", "unit testing")),
    SkillPattern("jenkins", ("jenkins",)),
    SkillPattern("terraform", ("terraform",)),
    SkillPattern("ansible", ("ansible",)),
    SkillPattern("devops", ("devops", "dev ops")),
    SkillPattern("data engineering", ("data engineering", "data engineer")),
    SkillPattern("machine learning", ("machine learning", "ml")),
    SkillPattern("openai apis", ("openai api", "openai apis", "openai")),
    SkillPattern("ci/cd", ("ci/cd", "cicd", "continuous integration", "continuous delivery", "continuous deployment")),
)

def _compile_alias(alias: str) -> re.Pattern[str]:
    """Compile an alias pattern that matches whole words or explicit symbols."""
    alias = alias.strip().lower()
    if alias == "js":
        return re.compile(r"(?<![\w.])js(?!\w)")

    escaped = re.escape(alias)
    return re.compile(rf"(?<!\w){escaped}(?!\w)")


def _matches_any(text: str, patterns: tuple[re.Pattern[str], ...]) -> bool:
    return any(pattern.search(text) is not None for pattern in patterns)


_PATTERN_MAP = {
    pattern.canonical: tuple(_compile_alias(alias) for alias in pattern.aliases)
    for pattern in SKILL_PATTERNS
}


def extract_skills(
    text: str,
) -> list[str]:
    """Extract canonical skills with deterministic dictionary matching.

    Skills are returned in dictionary order, not match position order, so repeated
    runs over the same text produce stable output.
    """
    normalized_text = text.lower()
    found: list[str] = []
    seen: set[str] = set()

    for pattern in SKILL_PATTERNS:
        if _matches_any(normalized_text, _PATTERN_MAP[pattern.canonical]):
            if pattern.canonical not in seen:
                found.append(pattern.canonical)
                seen.add(pattern.canonical)

    return found


def extract_skills_from_texts(
    texts: Iterable[str],
) -> list[str]:
    """Extract canonical skills after combining multiple text sources."""
    combined_text = "\n".join(text for text in texts if text)
    return extract_skills(combined_text)
