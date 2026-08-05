#!/usr/bin/env bash

BEACON_URL="http://beacon.local:1337"

set -euo pipefail

REPOS=(
    # name|dir|color
    "one|/home/you/Projects/repo1|green"
    "two|/home/you/Projects/repo2|red"
)

for REPO_STR in "${REPOS[@]}"; do
    # Parse the repo string into vars
    IFS='|' read -r NAME DIR COLOUR <<< "$REPO_STR"

    echo "Checking ${NAME}..."

    if [[ ! -d "$DIR" ]]; then
        continue
    fi

    (
        cd "$DIR"
        git fetch --quiet

        # safety check
        if ! git rev-parse --abbrev-ref @{u} >/dev/null 2>&1; then
            exit 0 # Exits the subshell, not the whole script
        fi

        # Determine if the current branch is behind its upstream
        read -r LOCAL REMOTE < <(git rev-list --left-right --count HEAD...@{u})

        if (( REMOTE > 0 )); then
            curl -X POST \
                 -fsS \
                 -H "Content-Type: application/json" \
                 -d "{\"msg\": \"${NAME}\n${REMOTE} incoming!\", \"color\": \"${COLOUR}\"}" \
                 "${BEACON_URL}/spin/${REMOTE}/" > /dev/null
            sleep 10
        fi
    )
done
