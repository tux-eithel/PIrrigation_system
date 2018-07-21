.PONY: help rsync build deploy

.DEFAULT_GOAL := help


help:
	@echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

build: ## Build the hello app for ARM
	GOARCH=arm GOOS=linux go build -o robot *.go

rsync: ## Rsync hello exec to remote server
	@printf "Rsync file to remote server\n"
	@rsync -hv --progress robot raspy3b:~

deploy: build rsync

