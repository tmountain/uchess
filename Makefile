TXTMAN=txt2man
GO_BIN ?= "go"

man: docs/uchess.txt
	$(TXTMAN) -s1 -p -P uchess -t uchess docs/uchess.txt > docs/uchess.man

install: tidy
	cd ./cmd/uchess && $(GO_BIN) install
	make tidy

tidy:
	$(GO_BIN) mod tidy -v

build: tidy
	pkger -o ./cmd/uchess
	cd ./cmd/uchess && $(GO_BIN) build -v .
	make tidy

release:
	rm -rf dist && goreleaser

release_test:
	rm -rf dist && goreleaser --snapshot --skip-publish --rm-dist
