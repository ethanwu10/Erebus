PYTHON = python
PROTOC = $(PYTHON) -m grpc_tools.protoc

.DEFAULT_GOAL := build

.SECONDEXPANSION:

# General targets

.PHONY: prepare
prepare: $$(PROTO_GEN_SRC)

.PHONY: build
build: erebus-supervisor-controller.py $$(PROTO_GEN_SRC) requirements.txt
	@mkdir -p dist/erebus-supervisor-controller
	$(MAKE) -C ../../broker crossbuild
	cp $? \
		../../broker/broker_linux_amd64 \
		../../broker/broker_darwin_amd64 \
		../../broker/broker_windows_amd64.exe \
		dist/erebus-supervisor-controller

.PHONY: proto
proto: $$(PROTO_GEN_SRC)

.PHONY: clean
clean: cleandist cleanproto

# Plumbing / dependencies

PROTO_GEN_PATH = .
PROTOS_PATH = ../../shared/proto

GRPC_PROTOS = \
	$(PROTOS_PATH)/control.proto
PROTOS = \
	$(PROTOS_PATH)/sim.proto \
	$(PROTOS_PATH)/session.proto \
	$(PROTOS_PATH)/types.proto

define protorule
PROTO_GEN_SRC += $(PROTO_GEN_PATH)/$1_pb2.py

$(PROTO_GEN_PATH)/$1_pb2.py: $2
	$(PROTOC) -I $(PROTOS_PATH) --python_out=$(PROTO_GEN_PATH)/ $2
	#sed -i -E -e 's/import (.+_pb2.*)/from . import \1/' $(PROTO_GEN_PATH)/$1_pb2.py
endef
define grpc_protorule
PROTO_GEN_SRC += $(PROTO_GEN_PATH)/$1_pb2.py
PROTO_GEN_SRC += $(PROTO_GEN_PATH)/$1_pb2_grpc.py

$(PROTO_GEN_PATH)/$1_pb2%py $(PROTO_GEN_PATH)/$1_pb2_grpc%py: $2
	$(PROTOC) -I $(PROTOS_PATH) --python_out=$(PROTO_GEN_PATH)/ --grpc_python_out=$(PROTO_GEN_PATH)/ $2
	#sed -i -E -e 's/import (.+_pb2.*)/from . import \1/' $(PROTO_GEN_PATH)/$1_pb2.py $(PROTO_GEN_PATH)/$1_pb2_grpc.py
endef

$(foreach pb,$(PROTOS),$(eval $(call \
	protorule,$(basename $(notdir $(pb))),$(pb))))
$(foreach pb,$(GRPC_PROTOS),$(eval $(call \
	grpc_protorule,$(basename $(notdir $(pb))),$(pb))))

.PHONY: cleanproto
cleanproto:
	rm -rf $(PROTO_GEN_PATH)/*_pb2*.py
