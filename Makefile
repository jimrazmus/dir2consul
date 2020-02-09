VERSION=`git rev-parse HEAD`
BUILD=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.PHONY: build
build:
	@export DOCKER_CONTENT_TRUST=1 && docker build -f Dockerfile -t dir2consul .

.PHONY: run
run:
	@docker run -p 80:8181 --env-file=.env dir2consul:latest

.PHONY: rc
rc:
	@docker run -p 80:8181 --env-file=.env jimrazmus/dir2consul:rc