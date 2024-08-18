export SHELL:=/bin/bash
export DEBUG:=false

help: ## Prints help for targets with comments
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

###########
## D E V ##
###########

generate: ## Generate code (e.g. mocks)
	@scripts/tools.sh generate

install: ## Install all required dependencies
	@scripts/tools.sh install

cleanup: ## Cleanup project
	@scripts/tools.sh cleanup

update: ## Update dependencies
	@scripts/tools.sh update

fmt: ## Code formatting
	@scripts/tools.sh fmt

statan: ## Static Analysis
	@scripts/tools.sh statan

tag: ## Create a Git tag from the branch name
	@scripts/tools.sh tag

test: generate ## Run all tests
	@scripts/tools.sh test && scripts/tools.sh cleanup

test-benchmark: ## Run all benchmark tests
	@scripts/tools.sh benchmark