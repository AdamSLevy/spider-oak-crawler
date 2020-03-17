# Let go build do the work of figuring out updated dependencies.
.PHONY: crawl crawld generate

all: crawl crawld

crawl:
	go build ./cmd/crawl

crawld:
	go build ./cmd/crawld

generate:
	go generate ./internal/crawl
