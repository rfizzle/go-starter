#!/bin/bash
# Start Claude Code with planning agent pre-loaded

FEATURE_NAME="$1"

if [ -z "$FEATURE_NAME" ]; then
    echo "Usage: $0 <feature-name>"
    echo "Example: $0 user-authentication"
    exit 1
fi

echo "Starting Claude Code planning session for: $FEATURE_NAME"
claude -p "plan: $FEATURE_NAME"