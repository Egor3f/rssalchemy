PROTOBUF_TAGGER_PATH := ${GOPATH}/pkg/mod/github.com/srikrsna/protoc-gen-gotag@v1.0.2

all:

js_proto:
	protoc -I=. -I=${PROTOBUF_TAGGER_PATH} --ts_out=./frontend/wizard-vue/src/urlmaker ./proto/specs.proto
	sed -i '1 i //@ts-nocheck' ./frontend/wizard-vue/src/urlmaker/google/protobuf/descriptor.ts

go_proto:
	protoc -I=. -I=${PROTOBUF_TAGGER_PATH} --go_out=. ./proto/specs.proto
	protoc -I=. -I=${PROTOBUF_TAGGER_PATH} --gotag_out=. ./proto/specs.proto

proto: js_proto go_proto
