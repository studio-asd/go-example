REPOSITORY_ROOT := $(shell git rev-parse --show-toplevel)
DB_SCRIPT := $(REPOSITORY_ROOT)/database/all.bash
RIVER_QUEUE := github.com/riverqueue/river@v0.16.0

.PHONY:
dbinstall:
	@echo "Installing database tools"
ifeq (,$(shell which migrate))
# Note that this will install the migrate CLI on a specific version but 'migrate -version' will always returns 'dev'.
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3
endif
ifeq (,$(shell which sqlc))
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
endif

.PHONY: upall
upall:
ifndef ($(GITHUB_ACTIONS))
		make dbcomposeup
endif
	@$(MAKE) $(gen_dirs) command=up

.PHONY: downall
downall:
	@$(DB_SCRIPT) down

.PHONY: dbcomposeup
dbcomposeup:
	@docker compose -f $(REPOSITORY_ROOT)/database/docker-compose.yaml up -d --remove-orphans
	@sleep 5

.PHONY: dbcomposedown
dbcomposedown:
	@docker compose -f $(REPOSITORY_ROOT)/database/docker-compose.yaml down -v

gen_dirs := \
	go-example/ledger \
	user-data/user \
	user-data/rbac

$(gen_dirs):
ifeq (generate,$(command))
	@$(DB_SCRIPT) $(command) $@
endif
ifeq (up,$(command))
	@$(DB_SCRIPT) $(command) $@
endif

.PHONY: dbgenall
dbgenall :
ifndef ($(GITHUB_ACTIONS))
	make dbcomposeup
endif
	-$(MAKE) $(gen_dirs) command=generate
ifndef ($(GITHUB_ACTIONS))
	make dbcomposedown
endif
