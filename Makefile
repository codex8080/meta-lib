GOCC := go

BINS := ./bin
VERSION := $(shell git describe --always --long)


all: build

deps:
	go mod tidy

car:
	rm -f ${BINS}/car
	$(GOCC) build -o ${BINS}/car ./cmd/car


build: deps

install:
	install -C ${BINS}/car /usr/local/bin/car

clean:
	rm -rf ${BINS}/*
	$(GOCC) clean

.PHONY: deps