REPOSITORY_ROOT := $(shell git rev-parse --show-toplevel)
DB_SCRIPT := $(REPOSITORY_ROOT)/database/all.bash

gen_dirs := \
	go-example/ledger \
	go-example/wallet

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
	@$(MAKE) $(gen_dirs) command=generate
ifndef ($(GITHUB_ACTIONS))
	make dbcomposedown
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
	@sleep 2

.PHONY: dbcomposedown
dbcomposedown:
	@docker compose -f $(REPOSITORY_ROOT)/database/docker-compose.yaml down -v
