#!/usr/bin/env bash
# Usage
#   bash scripts/release.sh v0.3.2

set -eu
set -o pipefail

REMOTE=https://github.com/suzuki-shunsuke/run-ci

ee() {
  echo "+ $*"
  eval "$@"
}

BRANCH="$(git branch | grep "^\* " | sed -e "s/^\* \(.*\)/\1/")"
if [ "$BRANCH" != "master" ]; then
  read -r -p "The current branch isn't master but $BRANCH. Are you ok? (y/n)" YN
  if [ "${YN}" != "y" ]; then
    echo "cancel to release"
    exit 0
  fi
fi

TAG="$1"
echo "TAG: $TAG"
VERSION="${TAG#v}"

if [ "$TAG" = "$VERSION" ]; then
  echo "the tag must start with 'v'" >&2
  exit 1
fi

cd "$(dirname "$0")/.."

VERSION_FILE=pkg/constant/version.go
echo "create $VERSION_FILE"
cat << EOS > "$VERSION_FILE"
package constant

// Don't edit this file.
// This file is generated by the release command.

// Version is the run-ci's version.
const Version = "$VERSION"
EOS

ee git add "$VERSION_FILE"
echo "+ git commit -m \"build: update version to $TAG\""
git commit -m "build: update version to $TAG"
ee git tag "$TAG"
ee git push "$REMOTE" "$BRANCH"
ee git push "$REMOTE" "$TAG"
