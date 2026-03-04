#!/usr/bin/env bash
set -euo pipefail

CONFIG_FILE="$GITHUB_ACTION_PATH/config/board.json"

PROJECT_ID=$(jq -r '.projectId' "$CONFIG_FILE")
STATUS_FIELD_ID=$(jq -r '.statusFieldId' "$CONFIG_FILE")

if [ "$PROJECT_ID" = "null" ] || [ -z "$PROJECT_ID" ]; then
  echo "" >&2
  echo "::error::Missing projectId in board config" >&2
  exit 1
fi

OPTION_ID=$(jq -r --arg status "$TARGET_STATUS" '.options[$status] // empty' "$CONFIG_FILE")

if [ -z "$OPTION_ID" ]; then
  echo "" >&2
  echo "::error::Could not find option ID for status '$TARGET_STATUS' in board config ($CONFIG_FILE)." >&2
  echo "Available options:" >&2
  jq -r '.options | keys[] | "- \(. | tostring)"' "$CONFIG_FILE" >&2
  exit 1
fi

echo "Loaded config for board: $PROJECT_ID"
echo "Status field ID:  $STATUS_FIELD_ID"
echo "Option ID '$TARGET_STATUS': $OPTION_ID"

{
  echo "project_id=$PROJECT_ID"
  echo "status_field_id=$STATUS_FIELD_ID"
  echo "option_id=$OPTION_ID"
} >> "$GITHUB_OUTPUT"
