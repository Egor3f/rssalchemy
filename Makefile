PROTOBUF_TAGGER_PATH := ${GOPATH}/pkg/mod/github.com/srikrsna/protoc-gen-gotag@v1.0.2

all:

js_proto:
	protoc -I=. -I=${PROTOBUF_TAGGER_PATH} --ts_out=./frontend/wizard-vue/src/urlmaker ./proto/specs.proto
	# Remove unneeded code left from GO plugin (and corresponsing unused import)
	sed -i -E '/import.+tagger\/tagger/d' ./frontend/wizard-vue/src/urlmaker/proto/specs.ts
	rm -rf ./frontend/wizard-vue/src/urlmaker/google
	rm -rf ./frontend/wizard-vue/src/urlmaker/tagger

go_proto:
	protoc -I=. -I=${PROTOBUF_TAGGER_PATH} --go_out=. ./proto/specs.proto
	protoc -I=. -I=${PROTOBUF_TAGGER_PATH} --gotag_out=. ./proto/specs.proto

proto: js_proto go_proto

update_adblock:
	wget -O internal/extractors/pwextractor/blocklists/easylist.txt https://easylist.to/easylist/easylist.txt
	wget -O internal/extractors/pwextractor/blocklists/easyprivacy.txt https://easylist.to/easylist/easyprivacy.txt
