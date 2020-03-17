# Let go build do the work of figuring out updated dependencies.
.PHONY: crawl crawld generate

all: crawl crawld

crawl: generate
	go build ./cmd/crawl

crawld: generate
	go build ./cmd/crawld

generate:
	go generate ./internal/crawl
