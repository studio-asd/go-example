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

.PHONY: genall
genall:
	@$(MAKE) $(gen_dirs) command=generate

.PHONY: upall
upall:
	@$(MAKE) $(gen_dirs) command=up

.PHONY: downall
downall:
	@$(DB_SCRIPT) down

.PHONY: dbcomposeup
dbcomposeup:
	@docker compose -f $(REPOSITORY_ROOT)/database/docker-compose.yaml up -d --remove-orphans

.PHONY: dbcomposedown
dbcomposedown:
	@docker compose -f $(REPOSITORY_ROOT)/database/docker-compose.yaml down -v
