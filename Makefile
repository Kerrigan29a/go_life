MODULE = $(shell grep module go.mod | cut -d ' ' -f 2)
VERSION = $(shell grep "Version" version.go | cut -d '"' -f 2)

test:
	go test

release: test
	git tag -a $(VERSION) -m "Releasing version $(VERSION)"
	git push origin HEAD
	git push origin tag $(VERSION)
	curl https://proxy.golang.org/$(MODULE)/@v/$(VERSION).info