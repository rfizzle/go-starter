#!/bin/bash
# Start Claude Code with review agent pre-loaded

REVIEW_REQUEST="${1:-Check recent changes}"

echo "Starting Claude Code review session: $REVIEW_REQUEST"
claude -p "review: $REVIEW_REQUEST"