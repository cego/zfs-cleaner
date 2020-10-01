VERSION = $(shell git describe --tags --dirty)

.PHONY: zfs-cleaner

all: zfs-cleaner

deb: zfs-cleaner
	dpkg-buildpackage -uc -us

test:
	go test ./... -race -cover

lint:
	golangci-lint run

zfs-cleaner:
	go build \
		-ldflags '-X main.version=$(VERSION)' \
		.

clean:
	rm -rf debian/.debhelper
	rm -f zfs-cleaner
