all: check

check: vet test

vet:
	go vet $(CURDIR)/...

test:
	go test $(CURDIR)/...

.PHONY: all build check vet test
