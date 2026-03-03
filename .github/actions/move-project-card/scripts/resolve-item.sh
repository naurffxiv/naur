#!/usr/bin/env bash
set -euo pipefail

RESPONSE=$(gh api graphql \
  -f query="$(cat "$GITHUB_ACTION_PATH/queries/resolve-project-items.graphql")" \
  -f issueId="$ISSUE_ID")

ITEM_ID=$(echo "$RESPONSE" | jq -r --arg proj "$PROJECT_ID" '
  .data.node.projectItems.nodes[]
  | select(.project.id == $proj)
  | .id
' | head -n 1)

if [ -z "$ITEM_ID" ] || [ "$ITEM_ID" = "null" ]; then
  echo "::warning::Issue is not yet added to Project #$PROJECT_NUMBER. Adding it now..."

  # Manually handle race conditions: if another job added the item simultaneously,
  # GitHub returns an error. We catch it and retry the lookup.
  ADD_RESPONSE=$(gh api graphql \
    -f query="$(cat "$GITHUB_ACTION_PATH/mutations/add-project-item.graphql")" \
    -f projectId="$PROJECT_ID" \
    -f contentId="$ISSUE_ID" 2>&1) || {
      if echo "$ADD_RESPONSE" | grep -q "already"; then
        echo "::notice::Item was added by concurrent process. Retrying lookup..."
        sleep 2
        exec "$0"
      fi
      echo "" >&2
      echo "::error::Failed to add issue: $(echo "$ADD_RESPONSE" | head -n 1)" >&2
      exit 1
    }

  ITEM_ID=$(echo "$ADD_RESPONSE" | jq -r '.data.addProjectV2ItemById.item.id')

  if [ -z "$ITEM_ID" ] || [ "$ITEM_ID" = "null" ]; then
    echo "::error::Failed to add issue to Project #$PROJECT_NUMBER."
    exit 1
  fi
fi

echo "Project item ID: $ITEM_ID"
echo "item_id=$ITEM_ID" >> "$GITHUB_OUTPUT"
