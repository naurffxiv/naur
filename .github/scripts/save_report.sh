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

SAFE_CHECK=$(echo "$CHECK_NAME" | sed 's/[^a-zA-Z0-9]/_/g')
ROW_OUTPUT="summary-${SERVICE}-${ORDER}-${SAFE_CHECK}.md"
LOG_OUTPUT="logs-${SERVICE}-${ORDER}-${SAFE_CHECK}.md"
JSON_OUTPUT="report-${SERVICE}-${ORDER}-${SAFE_CHECK}.json"
LOG_LINES=${LOG_LINES:-100}

# Generate Markdown Row
case "$STATUS" in
    success) icon="🟢 Passed"; [[ -n "$MESSAGE" ]] && icon="$icon : $MESSAGE" ;;
    warning) icon="⚠️ Warning" ;;
    skipped) icon="🚫 Skipped" ;;
    *)       icon="🔴 Failed" ;;
esac
echo "| **$CHECK_NAME** | $icon |" > "$ROW_OUTPUT"

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
  --arg log_file "$LOG_FILE" \
  --arg message "$MESSAGE" \
  --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  --arg log_content "$LOG_CONTENT" \
  '{service: $service, order: ($order | tonumber), check_name: $check_name, status: $status, log_file: $log_file, message: $message, timestamp: $timestamp, log_content: $log_content}' > "$JSON_OUTPUT"

# Generate Collapsible Logs (if failed)
if [[ "$STATUS" != "success" ]]; then
    if [[ -s "$LOG_FILE" ]]; then
        # Log file exists and has content
        {
            echo "<details><summary><strong>$CHECK_NAME Output</strong></summary>"
            echo -e "\n\`\`\`text"
            echo "$LOG_CONTENT"
            echo "\`\`\`\n"
            echo "</details>"
        } > "$LOG_OUTPUT"
    elif [[ -f "$LOG_FILE" ]]; then
        # Log file exists but is empty
        {
            echo "<details><summary><strong>$CHECK_NAME Output</strong></summary>"
            echo -e "\n\`\`\`text"
            echo "[Log file is empty]"
            echo "\`\`\`\n"
            echo "</details>"
        } > "$LOG_OUTPUT"
    else
        # Log file doesn't exist
        {
            echo "<details><summary><strong>$CHECK_NAME Output</strong></summary>"
            echo -e "\n\`\`\`text"
            echo "[Log file not found: $LOG_FILE]"
            echo "\`\`\`\n"
            echo "</details>"
        } > "$LOG_OUTPUT"
    fi
fi
