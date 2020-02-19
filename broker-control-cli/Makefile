
GO = go
PROTOC = protoc

GO111MODULE := on
export GO111MODULE

.DEFAULT_GOAL := build

.SECONDEXPANSION:

# General targets

.PHONY: build
build: $$(BUILDDEPS)
	$(GO) build

.PHONY: proto
proto: $$(PROTO_GEN_SRC)

# Plumbing / dependencies

BUILDDEPS = $(PROTO_GEN_SRC)

PROTOS = \
	../shared/proto/control.proto \
	../shared/proto/sim.proto \
	../shared/proto/types.proto

define protorule
PROTO_GEN_SRC += gen/$1.pb.go

gen/$1.pb.go: $2
	@mkdir -p gen
	$(PROTOC) -I ../shared/proto $2 --go_out=plugins=grpc:gen/

endef

$(foreach pb,$(PROTOS),$(eval $(call \
	protorule,$(basename $(notdir $(pb))),$(pb))))

.PHONY: clean
clean:
	rm -r gen/*
