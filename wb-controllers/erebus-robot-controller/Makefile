PYTHON = python
PROTOC = $(PYTHON) -m grpc_tools.protoc

.DEFAULT_GOAL := prepare

.SECONDEXPANSION:

# General targets

.PHONY: prepare
prepare: $$(PROTO_GEN_SRC)

.PHONY: proto
proto: $$(PROTO_GEN_SRC)

.PHONY: clean
clean: cleanproto

# Plumbing / dependencies

PROTO_GEN_PATH = .
PROTOS_PATH = ../../shared/proto

GRPC_PROTOS = \
	$(PROTOS_PATH)/wb_controller.proto
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
