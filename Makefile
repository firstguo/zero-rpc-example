MODULE = zero-rpc-example
VERSION = v1
PROTO_ROOT = buf_proto_example/proto
DOMAIN = example
SERVICE_PATH = base/svr/user
PROTO_FILE = $(PROTO_ROOT)/$(DOMAIN)/$(SERVICE_PATH)/$(VERSION)/user.proto

PB_OUT = .
GRPC_OUT = .
ZRPC_OUT = .

PB_PKG = $(MODULE)/buf_proto_example/gen/go/${DOMAIN}/$(SERVICE_PATH)/$(VERSION)
PB_ALIAS = $(notdir $(patsubst %/,%,$(dir $(PB_PKG))))

.PHONY: gen
gen:
	goctl rpc protoc $(PROTO_FILE) \
		--go_out=$(PB_OUT) \
		--go-grpc_out=$(GRPC_OUT) \
		--zrpc_out=$(ZRPC_OUT) \
		--go_opt=module=$(MODULE) \
		--go-grpc_opt=module=$(MODULE) \
		--go_opt=M$(PROTO_FILE)=$(PB_PKG) \
		--go-grpc_opt=M$(PROTO_FILE)=$(PB_PKG)
	find . -path ./buf_proto_example -prune -o -name '*.go' -print | xargs sed -i '' \
		-e 's|	"$(PB_PKG)"|	$(PB_ALIAS) "$(PB_PKG)"|g' \
		-e 's|$(VERSION)\.|$(PB_ALIAS).|g'
