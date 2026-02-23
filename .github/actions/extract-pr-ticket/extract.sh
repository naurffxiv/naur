#!/usr/bin/env bash
set -euo pipefail

# Dependabot PRs never contain ticket references per team convention
if [[ "$ACTOR" == "dependabot[bot]" || "$PR_USER" == "dependabot[bot]" ]]; then
  echo "Dependabot PR. Skipping."
  {
    echo "is_dependabot=true"
    echo "found=false"
  } >> "$GITHUB_OUTPUT"
  exit 0
fi

echo "is_dependabot=false" >> "$GITHUB_OUTPUT"

if [[ -z "${EVENT_PATH:-}" || ! -f "$EVENT_PATH" ]]; then
  echo "Event payload not found. Skipping."
  echo "found=false" >> "$GITHUB_OUTPUT"
  exit 0
fi

echo "Reading PR content from $EVENT_PATH"
PR_BODY=$(jq -r .pull_request.body "$EVENT_PATH" || echo "")
PR_TITLE=$(jq -r .pull_request.title "$EVENT_PATH" || echo "")
CONTENT=$(printf "%s\n%s" "$PR_TITLE" "$PR_BODY")
echo "Content length: ${#CONTENT} characters"

REGEX='^[[:space:]]*(ticket[: ]+#?([0-9]+|NA)|(fix(es|ed|ing)?|clos(e[ds]?|ing)|resolv(e[ds]?|ing)|ref(s)?)[[:space:]]+#?([0-9]+|NA)|chore\(deps\)|chore\(release\)|\bNA\b)'

echo "Checking content against regex..."
if ! echo "$CONTENT" | grep -qiE "$REGEX"; then
  echo "No ticket reference found in PR title or body. Skipping."
  echo "found=false" >> "$GITHUB_OUTPUT"
  exit 0
fi

echo "Match found. Extracting specific ticket..."
MATCH=$(echo "$CONTENT" | grep -oiE "$REGEX" | head -n 1 || true)

TICKET_NUMBER=$(echo "$MATCH" | grep -oE '[0-9]+' | head -n 1 || true)

if [ -z "$TICKET_NUMBER" ]; then
  echo "Ticket reference is NA or non-numeric. Skipping board sync."
  {
    echo "found=false"
    echo "match=$MATCH"
  } >> "$GITHUB_OUTPUT"
  exit 0
fi

echo "Ticket number found: #$TICKET_NUMBER"
{
  echo "found=true"
  echo "ticket=$TICKET_NUMBER"
  echo "match=$MATCH"
} >> "$GITHUB_OUTPUT"
