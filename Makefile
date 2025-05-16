include database/Makefile
include proto/Makefile

.PHONY: install
install:
	$(MAKE) dbinstall
	@echo "Installing tools"
ifeq (,$(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.0.2
endif

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
	make dbdown
	make dbup
	cd services && PGTEST_SKIP_PREPARE=1 go test -v -race ./... -run=TestTransact
	make dbdown
	cd internal && go test -v -race ./...

.PHONY: lint
lint: install
	golangci-lint run --verbose \
		--config=./.config/golangcilint.yaml \
		--print-resources-usage
