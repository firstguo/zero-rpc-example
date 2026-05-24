MODULE = zero-rpc-example
VERSION = v1
PROTO_ROOT = buf_proto_example/apis
DOMAIN = tripo

.PHONY: proto
proto:
	cd buf_proto_example && buf generate

.PHONY: gen-user
gen-user: proto
	@PROTO=$(PROTO_ROOT)/$(DOMAIN)/user/$(VERSION)/user.proto; \
	PB_PKG=$(MODULE)/buf_proto_example/gen/go/$(DOMAIN)/user/$(VERSION); \
	goctl rpc protoc $$PROTO \
		--go_out=. \
		--go-grpc_out=. \
		--zrpc_out=services/user \
		--go_opt=module=$(MODULE) \
		--go-grpc_opt=module=$(MODULE) \
		--go_opt=M$$PROTO=$$PB_PKG \
		--go-grpc_opt=M$$PROTO=$$PB_PKG
	@find services/user -name '*.go' -exec sed -i '' \
		-e 's|"$(MODULE)/buf_proto_example/gen/go/$(DOMAIN)/user/$(VERSION)"|user "$(MODULE)/buf_proto_example/gen/go/$(DOMAIN)/user/$(VERSION)"|g' \
		-e 's|v1\.|user.|g' \
		-e 's|user user "|user "|g' \
		{} +

.PHONY: gen-user-auth
gen-user-auth: proto
	@PROTO=$(PROTO_ROOT)/$(DOMAIN)/user_auth/$(VERSION)/user_auth.proto; \
	PB_PKG=$(MODULE)/buf_proto_example/gen/go/$(DOMAIN)/user_auth/$(VERSION); \
	goctl rpc protoc $$PROTO \
		--go_out=. \
		--go-grpc_out=. \
		--zrpc_out=services/user-auth \
		--go_opt=module=$(MODULE) \
		--go-grpc_opt=module=$(MODULE) \
		--go_opt=M$$PROTO=$$PB_PKG \
		--go-grpc_opt=M$$PROTO=$$PB_PKG
	@find services/user-auth -name '*.go' -exec sed -i '' \
		-e 's|"$(MODULE)/buf_proto_example/gen/go/$(DOMAIN)/user_auth/$(VERSION)"|user_auth "$(MODULE)/buf_proto_example/gen/go/$(DOMAIN)/user_auth/$(VERSION)"|g' \
		-e 's|v1\.|user_auth.|g' \
		-e 's|user_auth user_auth "|user_auth "|g' \
		{} +

.PHONY: gen
gen: gen-user gen-user-auth

.PHONY: build-user
build-user:
	cd services/user && go build -o ../../bin/user-service .

.PHONY: build-user-auth
build-user-auth:
	cd services/user-auth && go build -o ../../bin/user-auth-service .

.PHONY: build
build: build-user build-user-auth

.PHONY: clean
clean:
	cd buf_proto_example && $(MAKE) clean
	rm -rf services/user services/user-auth bin
