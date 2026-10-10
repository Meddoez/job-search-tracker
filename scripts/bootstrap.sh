#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
VENV_DIR="$ROOT_DIR/.venv"
PYTHON_BIN=${PYTHON_BIN:-python3}

if [ ! -d "$VENV_DIR" ]; then
  "$PYTHON_BIN" -m venv "$VENV_DIR"
fi

"$VENV_DIR/bin/python" -m pip install --upgrade pip
"$VENV_DIR/bin/python" -m pip install -r "$ROOT_DIR/requirements.txt"
"$VENV_DIR/bin/python" -m pip install --editable "$ROOT_DIR"

cat <<EOF
Virtual environment ready at: $VENV_DIR
Activate it with:
  . "$VENV_DIR/bin/activate"
EOF
