include database/Makefile
include proto/Makefile

.PHONY: dbup
dbup:
	@$(MAKE) upall

.PHONY: dbdown
dbdown:
	@make downall

.PHONY: gendb
gendb:
	cd database && make dbgenall

.PHONY: genproto
genproto:
	cd proto && make protogenall

.PHONY: genall
genall: genproto gendb

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
