generate:
	protoc \
		-I api/v1/ \
		-I=$(GOPATH)/src \
		-I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/ \
		-I $(GOPATH)/src/github.com/gogo/googleapis/ \
		--gogo_out=plugins=grpc,\
Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api:\
$(PWD)/api/v1/ \
		api/v1/*.proto
install:
	go get \
		github.com/gogo/protobuf/protoc-gen-gogo \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger