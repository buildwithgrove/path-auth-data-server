###############################
### Generation Make Targets ###
###############################

.PHONY: sqlc_generate
sqlc_generate: ## Generate SQLC code from postgres/sqlc/*.sql files
	sqlc generate -f ./postgres/sqlc/sqlc.yaml