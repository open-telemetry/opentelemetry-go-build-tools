#!/bin/bash -ex

git config user.name "$GITHUB_ACTOR"
git config user.email "$GITHUB_ACTOR@users.noreply.github.com"

# shellcheck disable=SC2006
PR_NAME=dependabot-prs/$(date +'%Y-%m-%dT%H%M%S')
git checkout -b "$PR_NAME"

IFS=$'\n'
requests=$( gh pr list --search "author:app/dependabot" --json title --jq '.[].title' | sort )
message=""
dirs=$(find . -type f -name "go.mod" -exec dirname {} \; | sort )

for line in $requests; do
    if [[ $line != Bump* ]]; then
        continue
    fi

    module=$(echo "$line" | cut -f 2 -d " ")
    version=$(echo "$line" | cut -f 6 -d " ")
    
    topdir=$(pwd)
    for dir in $dirs; do
        echo "checking $dir"
        cd "$dir" && if grep -q "$module " go.mod; then go get "$module"@v"$version"; fi
        cd $topdir
    done
    message+=$line
    message+=$'\n'
done

make tidy lint

git add --all
git commit -m "dependabot updates $(date)
$message"
git push origin "$PR_NAME"

gh pr create --title "[chore] dependabot updates $(date)" --body "$message" -l "dependencies"
