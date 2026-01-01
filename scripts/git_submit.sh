#!/bin/bash

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Error handling function
handle_error() {
    echo -e "${RED}Error: $1${NC}"
    if [ "$2" == "rollback_stage" ]; then
        echo -e "${YELLOW}Rolling back staged changes...${NC}"
        git reset HEAD .
    fi
    echo "Aborting operation."
    exit 1
}

echo -e "${BLUE}=== Git Automation Submit Script ===${NC}"

# 1. Check git status
echo -e "\n${BLUE}[Step 1] Checking status...${NC}"
if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1; then
    handle_error "Not a git repository."
fi

STATUS_OUTPUT=$(git status)
echo -e "${YELLOW}----------------------------------------${NC}"
echo "$STATUS_OUTPUT"
echo -e "${YELLOW}----------------------------------------${NC}"

# Check if there are changes
if [[ "$STATUS_OUTPUT" == *"nothing to commit, working tree clean"* ]]; then
    echo -e "${GREEN}No changes to commit.${NC}"
    exit 0
fi

# 2. Input commit message
echo -e "\n${BLUE}[Step 2] Enter commit details${NC}"
echo "Please describe your changes (e.g., 'feat: add login page' or 'fix: resolve null pointer')."
read -p "Commit Message: " USER_MSG

if [ -z "$USER_MSG" ]; then
    handle_error "Commit message cannot be empty."
fi

# 3. Combine output
FULL_MSG="$USER_MSG

Git Status Snapshot:
$STATUS_OUTPUT"

# 4. Add files
echo -e "\n${BLUE}[Step 3] Staging files...${NC}"
git add .
if [ $? -ne 0 ]; then
    handle_error "Failed to stage files."
fi

# 5. Commit
echo -e "\n${BLUE}[Step 4] Committing...${NC}"
# Use a temp file to safely pass multiline message
MSG_FILE=$(mktemp)
echo "$FULL_MSG" > "$MSG_FILE"

git commit -F "$MSG_FILE"
COMMIT_EXIT=$?
rm "$MSG_FILE"

if [ $COMMIT_EXIT -ne 0 ]; then
    handle_error "Commit failed." "rollback_stage"
fi

echo -e "${GREEN}Commit successful!${NC}"

# 6. Push
echo -e "\n${BLUE}[Step 5] Push to remote?${NC}"
read -p "Execute 'git push'? (y/n): " PUSH_CONFIRM

if [[ "$PUSH_CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Pushing changes..."
    git push
    if [ $? -ne 0 ]; then
        echo -e "${RED}Push failed!${NC}"
        echo "Your changes are committed locally. You can try pushing manually later."
        exit 1
    fi
    echo -e "${GREEN}Successfully pushed to remote!${NC}"
else
    echo -e "${YELLOW}Skipped push. Changes are saved locally.${NC}"
fi
