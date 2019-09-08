#!/usr/bin/env bash
repo_path="$HOME/projects/unipdf-examples"
git clone https://github.com/unidoc/unipdf-examples.git $repo_path
cd $repo_path

branch_name="v3"
if [[ `git ls-remote origin "$TRAVIS_BRANCH"` ]]; then
    branch_name="$TRAVIS_BRANCH"
fi

git checkout $branch_name
echo "replace github.com/unidoc/unipdf/v3 => $TRAVIS_BUILD_DIR" >> go.mod
./build_examples.sh
