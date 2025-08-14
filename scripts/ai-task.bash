#!/bin/bash
# Start Claude Code with task agent pre-loaded

TASK_DESCRIPTION="$1"

if [ -z "$TASK_DESCRIPTION" ]; then
    echo "Usage: $0 '<task description>'"
    echo "Example: $0 'Implement Phase 1 from user authentication plan'"
    exit 1
fi

echo "Starting Claude Code task session for: $TASK_DESCRIPTION"
claude -p "task: $TASK_DESCRIPTION"