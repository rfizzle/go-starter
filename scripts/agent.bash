#!/bin/bash

# --- Environment-Based Agent Orchestrator ---
# This script loads file contents into environment variables, uses envsubst
# to create a final prompt, and then runs the specified AI tool.

set -e

# 1. VALIDATE ARGUMENTS
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <agent_name> <feature_name>"
    exit 1
fi

AGENT=$1
FEATURE_NAME=$2

# 2. DEFINE PATHS
AGENT_TEMPLATE_FILE="docs/ai/agents/${AGENT}.md.tpl"
FEATURE_DIR="docs/features/${FEATURE_NAME}"
FINAL_PROMPT_FILE=$(mktemp) # Create a temporary file for the final prompt

# Ensure the temporary file is cleaned up on exit
trap 'rm -f "$FINAL_PROMPT_FILE"' EXIT

# Check if essential files exist
if [ ! -f "$AGENT_TEMPLATE_FILE" ]; then
    echo "Error: Agent template file not found at ${AGENT_TEMPLATE_FILE}"
    exit 1
fi
if [ ! -d "$FEATURE_DIR" ]; then
    echo "Error: Feature directory not found at ${FEATURE_DIR}"
    exit 1
fi

echo "--- Orchestrator Initialized (EnvSubst Mode) ---"
echo "Agent:   ${AGENT}"
echo "Feature: ${FEATURE_NAME}"
echo "------------------------------------------------"

# 3. LOAD FILE CONTENTS INTO ENVIRONMENT VARIABLES
echo "Loading context into environment variables..."

# Helper function to safely load file content into an environment variable.
# If the file doesn't exist, the variable will be empty.
load_file_content() {
    local var_name="$1"
    local file_path="$2"
    local content=""

    if [ -f "$file_path" ]; then
        # Read file content. Using < operator is slightly more efficient than cat.
        content=$(<"$file_path")
    else
        # Print an informational message but continue, leaving the content empty.
        echo "Info: Optional context file not found, leaving ${var_name} empty: ${file_path}"
    fi
    # Export the variable with the content (or empty string).
    export "$var_name"="$content"
}

# Load project-level context files
load_file_content "PROJECT_BRIEF_CONTENT" "docs/project_brief.md"
load_file_content "CODING_STANDARDS_CONTENT" "docs/coding_standards.md"
load_file_content "TECH_STACK_CONTENT" "docs/tech_stack.md"
load_file_content "PACKAGE_PATTERN_CONTENT" "docs/package_pattern.md"

# Load feature-level context files
load_file_content "FEATURE_CONTENT" "${FEATURE_DIR}/feature.md"
load_file_content "FEATURE_REQUEST_CONTENT" "${FEATURE_DIR}/request.md"
load_file_content "FEATURE_ARCHITECTURE_CONTENT" "${FEATURE_DIR}/architecture.md"
load_file_content "FEATURE_PLAN_CONTENT" "${FEATURE_DIR}/plan.md"
load_file_content "FEATURE_DEVELOPMENT_LOG" "${FEATURE_DIR}/development-log.md"
load_file_content "FEATURE_REVIEW_CONTENT" "${FEATURE_DIR}/review.md"
load_file_content "FEATURE_REVIEW_ACTIONS_CONTENT" "${FEATURE_DIR}/review-actions.yaml"

# 4. CREATE FINAL PROMPT USING ENVSUBST
echo "Substituting environment variables into agent template..."
# Get a list of all exported VAR_CONTENT variables for envsubst
# This tells envsubst which variables to replace.
VARS_TO_SUBSTITUTE=$(env | grep '_CONTENT=' | sed 's/=.*//' | sed 's/^/$/')
envsubst "$VARS_TO_SUBSTITUTE" < "$AGENT_TEMPLATE_FILE" > "$FINAL_PROMPT_FILE"
echo "Final prompt created at: ${FINAL_PROMPT_FILE}"

# 5. EXECUTE TOOL-SPECIFIC LOGIC
case "$AGENT" in
  architect)
    aider --model "gpt-5" --reasoning-effort high --message-file "$FINAL_PROMPT_FILE"
    ;;
  plan)
    claude < "$FINAL_PROMPT_FILE"
    ;;
  code)
    claude < "$FINAL_PROMPT_FILE"
    ;;
  test)
    aider --model "o3-pro" --reasoning-effort high --message-file "$FINAL_PROMPT_FILE"
    ;;
  review)
    claude < "$FINAL_PROMPT_FILE"
    ;;
  *)
    echo "Error: Unknown agent '$AGENT'."
    exit 1
    ;;
esac

echo "--- Orchestrator Finished ---"