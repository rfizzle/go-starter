#!/bin/bash

# Claude Agent Setup Script
# Creates symbolic links from docs/ai/prompts to ~/.claude/agents/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROMPTS_DIR="$PROJECT_ROOT/docs/ai/prompts"
CLAUDE_AGENTS_DIR="$PROJECT_ROOT/.claude/agents"

echo "Setting up Claude agents..."

# Create Claude agents directory if it doesn't exist
if [ ! -d "$CLAUDE_AGENTS_DIR" ]; then
    echo "Creating Claude agents directory: $CLAUDE_AGENTS_DIR"
    mkdir -p "$CLAUDE_AGENTS_DIR"
fi

# Check if prompts directory exists
if [ ! -d "$PROMPTS_DIR" ]; then
    echo "Error: Prompts directory not found: $PROMPTS_DIR"
    exit 1
fi

# Remove existing agent links if they exist
echo "Removing existing agent links..."
rm -f "$CLAUDE_AGENTS_DIR"/*.md

# Create symbolic links for individual agent files
echo "Creating agent links..."
for prompt_file in "$PROMPTS_DIR"/*.md; do
    if [ -f "$prompt_file" ]; then
        agent_name=$(basename "$prompt_file")
        agent_link="$CLAUDE_AGENTS_DIR/$agent_name"
        echo "  Linking $agent_name..."
        ln -sf "$prompt_file" "$agent_link"
    fi
done

echo "Setup complete! Available agents:"
find "$CLAUDE_AGENTS_DIR" -name "*.md" -type f -exec basename {} .md \; 2>/dev/null | sed 's/^/  /' || echo "  (none found)"