#!/bin/bash

# --- Initialize Feature Agent ---
# Creates a new, numbered feature directory and a corresponding git branch.

set -e

# 1. VALIDATE INPUT
if [ -z "$1" ]; then
    echo "Error: Feature title is required."
    echo "Usage: $0 \"Your New Feature Title\""
    exit 1
fi

FEATURE_TITLE="$1"

# 2. DETERMINE THE NEXT FEATURE NUMBER
# Find the highest existing number in the 'features' directory.
LATEST_NUM=$(ls -d features/*/ 2>/dev/null | sed 's|features/||' | cut -d'-' -f1 | sort -n | tail -n 1)

# If no features exist, start at 0. Otherwise, use the latest number.
if [ -z "$LATEST_NUM" ]; then
    LATEST_NUM=0
fi

NEW_NUM=$((LATEST_NUM + 1))
# Format the number to be three digits with leading zeros (e.g., 001, 012).
FORMATTED_NUM=$(printf "%03d" $NEW_NUM)

# 3. CREATE THE DIRECTORY NAME (SLUG)
# Convert title to lowercase, replace spaces with hyphens, and remove special chars.
SLUG=$(echo "$FEATURE_TITLE" | tr '[:upper:]' '[:lower:]' | tr -s ' ' '-' | sed 's/[^a-z0-9-]//g')
DIR_NAME="${FORMATTED_NUM}-${SLUG}"
FEATURE_PATH="docs/features/${DIR_NAME}"

# 4. CREATE THE FEATURE DIRECTORY AND FILE
if [ -d "$FEATURE_PATH" ]; then
    echo "Error: Feature directory already exists at ${FEATURE_PATH}"
    exit 1
fi

echo "Creating new feature directory: ${FEATURE_PATH}"
mkdir -p "$FEATURE_PATH"

# Copy the template and rename it to request.md
cp feature_template.md "${FEATURE_PATH}/request.md"

# Replace the placeholder title in the new request.md file.
sed -i.bak "s/\[Feature Title\]/$FEATURE_TITLE/g" "${FEATURE_PATH}/request.md"
rm "${FEATURE_PATH}/request.md.bak" # Clean up the backup file

# 5. CREATE AND CHECKOUT GIT BRANCH
BRANCH_NAME="feature/${DIR_NAME}"
echo "Switching to new git branch: ${BRANCH_NAME}"
git checkout -b "$BRANCH_NAME"

echo ""
echo "✅ Feature created and checked out successfully!"
echo "   Path: ${FEATURE_PATH}/request.md"
echo "   Branch: $(git rev-parse --abbrev-ref HEAD)"
echo "   Next step: Fill out the details in the request.md file, then commit the new feature request."