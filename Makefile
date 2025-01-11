include database/Makefile

.PHONY: dbup
dbup:
	@$(MAKE) upall

.PHONY: dbdown
dbdown:
	@make downall

.PHONY: test
test:
	@make dbdown
	@make dbup
	@cd services && go test -v -race ./...
	@make dbdown
	@cd internal && go test -v -race ./...
