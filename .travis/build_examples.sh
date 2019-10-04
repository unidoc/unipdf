#!/usr/bin/env bash
repo_path="$HOME/projects/unipdf-examples"
git clone https://github.com/unidoc/unipdf-examples.git $repo_path
cd $repo_path

build_branch="$TRAVIS_BRANCH"
if [[ -n "$TRAVIS_PULL_REQUEST_BRANCH" ]]; then
    build_branch="$TRAVIS_PULL_REQUEST_BRANCH"
fi
if [[ -z "$build_branch" ]]; then
    build_branch="development"
fi

target_branch="development"
if [[ `git ls-remote origin "$build_branch"` ]]; then
    target_branch="$build_branch"
fi

git checkout $target_branch
echo "replace github.com/unidoc/unipdf/v3 => $TRAVIS_BUILD_DIR" >> go.mod
./build_examples.sh
