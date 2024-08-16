#!/bin/bash

# Note: the set [-/+] x is purely there to turn on and off outputting of the commands being executed.
if [ "${DEBUG}" = "true" ]; then
  set -x
fi

case "$1" in

"install")
  command -v mockgen >/dev/null || go install go.uber.org/mock/mockgen@latest
  command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
  command -v goimports >/dev/null || go install golang.org/x/tools/cmd/goimports@latest
  command -v gofumpt >/dev/null || go install mvdan.cc/gofumpt@latest
  command -v kind >/dev/null || go install sigs.k8s.io/kind@latest
  ;;

"cleanup")
  rm -Rf tmp var .ssh
  ;;

"fmt")
  gofumpt -l -w .
  go mod tidy -v
  goimports -w .
  golangci-lint run --fix
  ;;

"generate")
  go generate -v ./...
  ;;

"update")
  go get -u ./...
  ;;

"statan")
  golangci-lint run
  ;;

"tag")
  # Get the current branch name
  branch="$(git rev-parse --abbrev-ref HEAD)"
  # Define the latest commit hash
  hash="$(git rev-parse --short HEAD)"

  # Create tag using branch name and commit hash
  tag="$branch-$hash"
  echo "Creating tag $tag"
  git tag "$tag" && git push origin "$tag"
  ;;

"test")
  mkdir -p tmp
  go test $(go list ./... | grep -v /testing/) -race -failfast -cover -coverprofile=coverage.out ./...
  test_exit_code=$?
  if [ $test_exit_code -ne 0 ]; then
    echo "Tests failed."
    exit 1
  fi
  cat coverage.out | grep -v "/mocks" | grep -v "/cmd" | grep -v "/testing" > coverage.redacted.out
  go tool cover -func coverage.redacted.out
  echo "Current test coverage at $(go tool cover -func coverage.redacted.out | tail -n 1 | awk '{print $3}' | cut -d "." -f 1) %"
  ;;

"benchmark")
  go test -bench=.  ./...
  ;;

"version")
  git describe --tags "$(git rev-list --tags --max-count=1)"
  ;;

*)
  echo "error: incorrect '$1' command..."
  ;;

esac
