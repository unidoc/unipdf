#!/usr/bin/env bash
repo_path="$GOPATH/src/github.com/unidoc"
mkdir -p $repo_path
cd $repo_path
git clone https://github.com/unidoc/unipdf-examples.git
cd unipdf-examples

branch_name="v3"
if [ `git branch --list "$TRAVIS_BRANCH"` ]
then
    branch_name="$TRAVIS_BRANCH"
fi

git checkout $branch_name
./build_examples.sh
