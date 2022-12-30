GOCC := go

BINS := ./bin
VERSION := $(shell git describe --always --long)


all: build

deps:
	go mod tidy

meta-car:
	rm -f ${BINS}/meta-car
	$(GOCC) build -o ${BINS}/meta-car ./cmd/meta-car


build: deps meta-car

install:
	install -C ${BINS}/meta-car /usr/local/bin/meta-car

clean:
	rm -rf ${BINS}/*
	$(GOCC) clean

.PHONY: deps