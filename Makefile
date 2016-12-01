.PHONY: pb data build

pb:
	for f in ./protos/**/*.proto; do \
		protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. $$f; \
		protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. $$f; \
		protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --swagger_out=logtostderr=true:. $$f; \
		echo compiled: $$f; \
	done

lint:
	./bin/lint.sh

build:
	./bin/build.sh

data:
	go-bindata -o data/bindata.go -pkg data data/*.json

run:
	docker-compose build
	docker-compose up
