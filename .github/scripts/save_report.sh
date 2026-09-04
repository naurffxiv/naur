#!/bin/bash
set -euo pipefail

# Usage: ./save_report.sh SERVICE ORDER CHECK_NAME STATUS [LOG_FILE] [MESSAGE]
SERVICE=$1
ORDER=$2
CHECK_NAME=$3
STATUS=$4
LOG_FILE=${5:-}
MESSAGE=${6:-}

if [[ -z "$SERVICE" || -z "$ORDER" || -z "$CHECK_NAME" || -z "$STATUS" ]]; then
    echo "Usage: $0 SERVICE ORDER CHECK_NAME STATUS [LOG_FILE] [MESSAGE]" >&2
    exit 1
fi

if [[ ! "$STATUS" =~ ^(success|failure|warning|skipped)$ ]]; then
    echo "Error: STATUS must be 'success', 'failure', 'warning', or 'skipped'" >&2
    exit 1
fi

# Create output directory if it doesn't exist
OUTPUT_DIR="${OUTPUT_DIR:-.}"
mkdir -p "$OUTPUT_DIR"

SAFE_CHECK="${CHECK_NAME//[^a-zA-Z0-9]/_}"
JSON_OUTPUT="$OUTPUT_DIR/report-${SERVICE}-${ORDER}-${SAFE_CHECK}.json"
LOG_LINES=${LOG_LINES:-100}

# Generate JSON Report
LOG_CONTENT=""
if [[ "$STATUS" != "success" && -f "$LOG_FILE" ]]; then
    LOG_CONTENT=$(tail -n "$LOG_LINES" "$LOG_FILE")
fi

jq -n \
  --arg service "$SERVICE" \
  --arg order "$ORDER" \
  --arg check_name "$CHECK_NAME" \
  --arg status "$STATUS" \
  --arg message "$MESSAGE" \
  --arg log_content "$LOG_CONTENT" \
  '{service: $service, order: ($order | tonumber), check_name: $check_name, status: $status, message: $message, log_content: $log_content}' > "$JSON_OUTPUT"

echo "Report saved: $JSON_OUTPUT"
