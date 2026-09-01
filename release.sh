#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if we're on main branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$BRANCH" != "main" ]; then
    echo -e "${YELLOW}Warning: You're on branch '$BRANCH', not 'main'${NC}"
    read -p "Continue anyway? (y/N) " confirm
    if [ "$confirm" != "y" ]; then
        exit 1
    fi
fi

# Check for uncommitted changes
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo -e "${RED}Error: You have uncommitted changes${NC}"
    exit 1
fi

# Pull latest
echo "Pulling latest changes..."
git pull --rebase

# Show existing tags
echo -e "\n${GREEN}Existing tags:${NC}"
git tag -l 'v*' | tail -5

# Prompt for new tag
echo -e "\nEnter new tag (e.g., v1.2.3):"
read -r tag

# Validate semver format
if ! [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    echo -e "${RED}Error: Invalid tag format. Expected format: v1.2.3 or v1.2.3-beta.1${NC}"
    exit 1
fi

# Check if tag exists locally
if git rev-parse "$tag" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag '$tag' already exists locally${NC}"
    exit 1
fi

# Check if tag exists on remote
if git ls-remote --tags origin "$tag" | grep -q "$tag"; then
    echo -e "${RED}Error: Tag '$tag' already exists on remote${NC}"
    exit 1
fi

# Create and push tag
echo -e "\nCreating tag '$tag'..."
git tag -a "$tag" -m "Release $tag"

echo -e "${GREEN}Tag '$tag' created.${NC}"
read -p "Push to origin? (y/N) " push

if [ "$push" = "y" ]; then
    git push origin "$tag"
    echo -e "${GREEN}Done! Release will be built at: https://github.com/biisal/bai/releases/tag/$tag${NC}"
else
    echo -e "${YELLOW}Tag created but not pushed. Run 'git push origin $tag' when ready.${NC}"
fi
