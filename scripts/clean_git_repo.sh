#!/bin/bash

# Check if both tag and repo are provided
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Please provide a tag to keep and the repository name. Usage: $0 <tag-to-keep> <repository-name>"
    exit 1
fi

TAG_TO_KEEP=$1
GITHUB_REPO=$2 # Repository name in 'username/repo' format

# Function to delete tag
delete_tag() {
    local tag=$1
    git tag -d "$tag"               # Delete tag locally
    git push origin --delete "$tag" # Delete tag remotely
}

# Function to delete GitHub release
delete_release() {
    local tag=$1
    # Get release ID by tag
    local release_id=$(curl -s \
        "https://api.github.com/repos/$GITHUB_REPO/releases/tags/$tag" |
        jq -r '.id')

    if [ "$release_id" != "null" ]; then
        response=$(curl -s -X DELETE \
            "https://api.github.com/repos/$GITHUB_REPO/releases/$release_id")
        echo "Deleting release $release_id: $response"
    else
        echo "No release found for tag $tag"
    fi
}

echo "All draft releases deleted."
# Get all tags, exclude the one to keep, and delete the rest
git fetch --tags
git tag | grep -v "$TAG_TO_KEEP" | while read -r tag; do
    delete_release "$tag"
    delete_tag "$tag"
done

echo "All tags and corresponding releases deleted except for $TAG_TO_KEEP in repository $GITHUB_REPO"
