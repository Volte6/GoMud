
.DEFAULT_GOAL := build

VERSION ?= $(shell git rev-parse HEAD)
BIN ?= go-mud-server
DOCKER_COMPOSE := docker-compose -f provisioning/docker-compose.yml

export GOFLAGS := -mod=mod
export GOSUMDB := off

## Build Targets

.PHONY: docker_build 
docker_build: 
	TAG=$(VERSION) $(DOCKER_COMPOSE) build server

DOCKER_CMD ?= bash

.PHONY: console
console:
	@docker run --rm -it --init \
			-v "$(PWD)":/src \
			-w /src \
			golang:1.21.3-alpine3.18 \
			$(DOCKER_CMD)

docker-%:
	@$(MAKE) console DOCKER_CMD="make $(patsubst docker-%,%,$@)"

# Clean both development and production containers
.PHONY: clean
clean:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	docker system prune -a

## Run Targets

.PHONY: run 
run: ### Build and run server.
	$(DOCKER_COMPOSE) up --build --remove-orphans server

.PHONY: client
client: ### Build and run client terminal client
	$(DOCKER_COMPOSE) run --rm terminal telnet go-mud-server 33333



.PHONY: image_tag
image_tag:
	@echo $(VERSION)

.PHONY: port
port:
	@$(eval PORT := $(shell $(DOCKER_COMPOSE) port server 8080))
	@echo $(PORT)

.PHONY: shell
shell:
	@$(eval CONTAINER_NAME := $(shell docker ps --filter="name=mud" --format '{{.Names}}' ))
	docker exec -it $(CONTAINER_NAME) /bin/sh 

#
#
# Local code run/test
#
#
.PHONY: validate
validate: fmtcheck vet

.PHONY: build_rpi
build_rpi: ### Build a binary for a raspberry pi
	env GOOS=linux GOARCH=arm GOARM=5 go build -o $(BIN)-rpi

.PHONY: build_only
build: validate build_only  ### Validate the code and build the binary.

.PHONY: build_only
build_only:
	CGO_ENABLED=0 go build -trimpath -a -o $(BIN) 

# Go targets

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: fmtcheck
fmtcheck:
	@set -e; \
	unformatted=$$(go fmt ./...); \
	if [ ! -z "$$unformatted" ]; then \
		echo Fixed inconsistent format in some files.; \
		echo $$unformatted; \
		exit 1; \
	fi

.PHONY: mod
mod:
	@go mod vendor
	@go mod tidy
	@go mod verify


.PHONY: vet
vet:
	@go vet

.PHONY: set_gopath
set_gopath:
ifeq ($(OS),Windows_NT)
	setx PATH "$(PATH);mytest" -m
else
	export GOPATH=$GOPATH:$(pwd)
endif

.PHONY: view_pprof_mem
view_pprof_mem:
	go tool pprof -http=:8989 source/_datafiles/profiles/mem.pprof


#
# Help target - greps and formats special comments to form a "help" command for makefiles
#
## Help
.PHONY: help
help:                 ### List makefile targets.
# 1. grep for any lines starting with "##" or containing "\s###\s"
# 2. Align targets/comments with string padding
# 3. Wrap lines starting with "##" in ANSI escape codes (color) as headers
# 4. Wrap lines containing "\s###\s" in ANSI escape codes (color) as commands
# 5. Add new lines before any that aren't prefixed with space (Headers)
	@grep -hE "^##\s|\s###\s" $(MAKEFILE_LIST) \
		| awk -F'[[:space:]]###[[:space:]]' '{printf "%-36s### %s\n", substr($$1,1,35), $$2}' \
		| sed -E "s/^## ([^#]*)#*/`printf "\033[90;3m"`\1`printf "\033[0m"`/" \
		| sed "s/\(.*\):\(.*\)###\(.*\)$$/  `printf "\033[93m"`\1:`printf "\033[36m"`\2`printf "\033[97m"`-\3`printf "\033[0m"`/" \
    	| sed "/^[^[:blank:]]/{x;p;x;}"
	@printf "\n"

