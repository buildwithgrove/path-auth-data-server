########################
### Makefile Helpers ###
########################

.PHONY: list
list: ## List all make targets
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help: ## Prints all the targets in all the Makefiles
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-60s\033[0m %s\n", $$1, $$2}'

####################
### Test Targets ###
####################

.PHONY: test_all
test_all: ## Runs all unit tests
	go test ./... -count=1

.PHONY: test_unit
test_unit: ## Runs unit tests
	go test ./... -short -count=1

####################
### Mock Targets ###
####################

.PHONY: gen_mocks
gen_mocks: ## Generates mocks for testing
	mockgen -source=./grpc/data_source.go -destination=./grpc/data_source_mock_test.go -package=grpc
