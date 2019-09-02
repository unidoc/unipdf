#!/usr/bin/env bash
repo_path="$GOPATH/src/github.com/unidoc/unipdf-examples"
git clone https://github.com/unidoc/unipdf-examples.git $repo_path
cd $repo_path

branch_name="v3"
if [[ `git ls-remote origin "$TRAVIS_BRANCH"` ]]; then
    branch_name="$TRAVIS_BRANCH"
fi

git checkout $branch_name
./build_examples.sh
