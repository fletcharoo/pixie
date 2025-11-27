# Makefile

.PHONY: help test

default: help

help: ## Show this help.
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

test: ## Run all tests.
	go test -count 1 ./...

test/update: ## Run all tests and update snapshots.
	UPDATE_SNAPS=true go test -count 1 ./... 