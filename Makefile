include database/Makefile

.PHONY: dbup
dbup:
	@$(MAKE) upall

.PHONY: dbdown
dbdown:
	@make downall

.PHONY: composeup
composeup: dbcomposeup

.PHONY: composedown
composedown: dbcomposedown

.PHONY: test
test:
	@make dbdown
	@make dbup
	@cd services && PGTEST_SKIP_PREPARE=1 go test -v -race ./...
	@make dbdown
	@cd internal && go test -v -race ./...
