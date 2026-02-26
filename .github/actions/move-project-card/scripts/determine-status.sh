#!/usr/bin/env bash

EVENT_NAME=$1
REVIEW_STATE=$2
IS_DRAFT=$3
IS_MERGED=$4

if [ "$EVENT_NAME" = "pull_request" ] && [ "$IS_MERGED" = "true" ]; then
  target="Done"
  reason="PR merged"
elif [ "$EVENT_NAME" = "pull_request_review" ] && [ "$REVIEW_STATE" = "approved" ]; then
  target="QA Review"
  reason="PR approved"
elif [ "$EVENT_NAME" = "pull_request_review" ] && [ "$REVIEW_STATE" = "changes_requested" ]; then
  target="Change requested"
  reason="Changes requested"
elif [ "$IS_DRAFT" = "true" ]; then
  target="In progress"
  reason="draft PR"
else
  target="In review"
  reason="ready-for-review PR"
fi

echo "Target status: $target ($reason)" >&2
echo "status=$target"
