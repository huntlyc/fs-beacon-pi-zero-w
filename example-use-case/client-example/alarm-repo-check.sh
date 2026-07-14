#!/usr/bin/env bash

# EDIT THIS!  This is the repo path
REPO_DIR="/path/to/your/repo"

set -euo pipefail

cd "$REPO_DIR"

git fetch --quiet

# Determine if the current branch is behind its upstream
read -r LOCAL REMOTE < <(
    git rev-list --left-right --count HEAD...@{u}
)

# REMOTE is the number of commits we're behind, spin that many times
if (( REMOTE > 0 )); then
    echo "Upstream has $REMOTE new commit(s). Triggering beacon..."
    curl -fsS "http://beacon:1337/spin/$REMOTE/" >/dev/null
else
    echo "Repository is up to date."
fi
