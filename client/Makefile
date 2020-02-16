POETRY = poetry
PYTHON = $(POETRY) run python
PROTOC = $(PYTHON) -m grpc_tools.protoc

.DEFAULT: prepare

.PHONY: prepare
prepare: proto

.PHONY: test
test: prepare
	$(PYTHON) -m pytest

.PHONY: clean
clean: cleanproto

PROTO_GEN_PATH = erebus/client

GRPC_PROTOS = \
	../shared/proto/client_controller.proto
PROTOS = \
	../shared/proto/sim.proto \
	../shared/proto/session.proto \
	../shared/proto/types.proto

define protorule
PROTO_GEN_SRC += $(PROTO_GEN_PATH)/$1_pb2.py

$(PROTO_GEN_PATH)/$1_pb2.py: $2
	$(PROTOC) -I ../shared/proto --python_out=$(PROTO_GEN_PATH)/ $2
	sed -i -E -e 's/import (.+_pb2.*)/from . import \1/' $(PROTO_GEN_PATH)/$1_pb2.py
endef
define grpc_protorule
PROTO_GEN_SRC += $(PROTO_GEN_PATH)/$1_pb2.py
PROTO_GEN_SRC += $(PROTO_GEN_PATH)/$1_pb2_grpc.py

$(PROTO_GEN_PATH)/$1_pb2%py $(PROTO_GEN_PATH)/$1_pb2_grpc%py: $2
	$(PROTOC) -I ../shared/proto --python_out=$(PROTO_GEN_PATH)/ --grpc_python_out=$(PROTO_GEN_PATH)/ $2
	sed -i -E -e 's/import (.+_pb2.*)/from . import \1/' $(PROTO_GEN_PATH)/$1_pb2.py $(PROTO_GEN_PATH)/$1_pb2_grpc.py
endef

$(foreach pb,$(PROTOS),$(eval $(call \
	protorule,$(basename $(notdir $(pb))),$(pb))))
$(foreach pb,$(GRPC_PROTOS),$(eval $(call \
	grpc_protorule,$(basename $(notdir $(pb))),$(pb))))

.PHONY: proto
proto: $(PROTO_GEN_SRC)

.PHONY: cleanproto
cleanproto:
	rm -rf erebus/client/*_pb2*.py
