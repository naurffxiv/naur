#!/usr/bin/env bash
set -euo pipefail

gh api graphql \
  -f query="$(cat "$GITHUB_ACTION_PATH/mutations/update-project-item-field.graphql")" \
  -f projectId="$PROJECT_ID" \
  -f itemId="$ITEM_ID" \
  -f fieldId="$FIELD_ID" \
  -f optionId="$OPTION_ID" || {
  echo "" >&2
  echo "::error::API call failed while moving ticket. Check permissions or board connection." >&2
  exit 1
}

echo "Ticket #$TICKET moved to '$TARGET_STATUS'"
