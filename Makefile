.PHONY: build lint upgrade help
all: build

# $Env:GOOS = "e" $Env:GOOS = "darwin"
# export GOOS=linux
build:
	CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -o ./bin/learnmark ./cmd/learnmark/

lint:
	golangci-lint run --verbose --timeout 50m

upgrade:
	go get -t -u ./...

run:
	./bin/learnmark server

help:
	@echo "make build: compile packages and dependencies"
	@echo "make lint: golangci-lint"
	@echo "make upgrade: upgrade deps"

.PHONY: api_gen api_dep_install api_clean
api_dep_install:
	# go env -w GOPROXY=https://goproxy.cn,direct
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/jstemmer/go-junit-report@latest
	go install github.com/mwitkow/go-proto-validators/protoc-gen-govalidators@latest
	go install github.com/rakyll/statik@latest

api_gen:
	protoc -I . -I third_party \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		--grpc-gateway_out=paths=source_relative:. \
		--openapiv2_out=logtostderr=true:. \
		--openapiv2_opt allow_merge=true \
		--openapiv2_opt output_format=json \
		--openapiv2_opt merge_file_name="learnmark." \
		api/general/v1/common.proto \
		api/general/v1/learnmark.proto \
		api/learnmark/v1/learnmark.proto \
		api/user/v1/user.proto
	cp -R *.swagger.json swagger-ui/learnmark.swagger.json
	rm *.swagger.json

api_clean:
	rm -f api/*/*/*.pb.go api/*/*/*.pb.gw.go api/*/*/*.swagger.json api/*/*/*.pb.validate.go
	rm -rf dist/sdk/*
	rm -rf docs/swagger-ui/*.swagger.json