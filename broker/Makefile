
GO = go
PROTOC = protoc
GOX = gox

CROSS_ARCH = amd64
CROSS_OS = linux darwin windows

GO111MODULE := on
export GO111MODULE

.DEFAULT: build

.PHONY: build
build: builddeps
	$(GO) build

.PHONY: crossbuild
crossbuild: builddeps
	$(GOX) -arch '$(CROSS_ARCH)' -os '$(CROSS_OS)'

.PHONY: builddeps
builddeps: proto

.PHONY: clean
clean: cleanproto
	rm -f broker $(foreach os,$(CROSS_OS),$(foreach arch,$(CROSS_ARCH),broker_$(os)_$(arch)))

PROTOS = \
	../shared/proto/client_controller.proto \
	../shared/proto/wb_controller.proto \
	../shared/proto/control.proto \
	../shared/proto/sim.proto \
	../shared/proto/session.proto \
	../shared/proto/types.proto

define protorule
PROTO_GEN_SRC += gen/$1.pb.go

gen/$1.pb.go: $2
	@mkdir -p gen
	$(PROTOC) -I ../shared/proto $2 --go_out=plugins=grpc:gen/

endef

$(foreach pb,$(PROTOS),$(eval $(call \
	protorule,$(basename $(notdir $(pb))),$(pb))))

.PHONY: proto
proto: $(PROTO_GEN_SRC)

.PHONY: cleanproto
cleanproto:
	rm -rf gen/*
