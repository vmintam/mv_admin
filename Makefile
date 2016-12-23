.PHONY: pb

pb:
	for f in ./protos/**/*.proto; do \
		protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:. $$f; \
		protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. $$f; \
		protoc -I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --swagger_out=logtostderr=true:. $$f; \
		echo compiled: $$f; \
	done

audios:
	cd services.db.audios/docker && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services.db.audios ../ && docker build -t services.db.audios .

videos:
	cd services.db.videos/docker && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services.db.videos ../ && docker build -t services.db.videos .

comments:
	cd services.db.comments/docker && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services.db.comments ../ && docker build -t services.db.comments .

users:
	cd services.db.users/docker && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services.db.users ../ && docker build -t services.db.users .


run:
	docker-compose build
	docker-compose up
