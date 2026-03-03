#!/usr/bin/env bash
set -euo pipefail

RESPONSE=$(gh api graphql \
  -f query="$(cat "$GITHUB_ACTION_PATH/queries/resolve-issue.graphql")" \
  -f owner="$REPO_OWNER" \
  -f repo="$REPO_NAME" \
  -F number="$TICKET")

ISSUE_ID=$(echo "$RESPONSE" | jq -r '.data.repository.issue.id')

if [ "$ISSUE_ID" = "null" ] || [ -z "$ISSUE_ID" ]; then
  echo "Issue #$TICKET not found in repository. Skipping."
  echo "found=false" >> "$GITHUB_OUTPUT"
  exit 0
fi

echo "Issue node ID: $ISSUE_ID"
{
  echo "found=true"
  echo "issue_id=$ISSUE_ID"
} >> "$GITHUB_OUTPUT"
