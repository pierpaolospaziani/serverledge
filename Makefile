BIN=bin

all: serverledge executor serverledge-cli lb

serverledge:
	go build -o $(BIN)/$@ cmd/$@/main.go

lb:
	CGO_ENABLED=0 go build -o $(BIN)/$@ cmd/$@/main.go

serverledge-cli:
	CGO_ENABLED=0 go build -o $(BIN)/$@ cmd/cli/main.go

executor:
	CGO_ENABLED=0 go build -o $(BIN)/$@ cmd/$@/executor.go

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/scheduling/protobuf/solver.proto

race:
	go build -race -o $(BIN)/serverledge cmd/serverledge/main.go

DOCKERHUB_USER=ferrarally
images:  image-python310 image-nodejs17ng image-base
image-python310:
	docker build -t $(DOCKERHUB_USER)/serverledge-python310 -f images/python310/Dockerfile .
image-base:
	docker build -t $(DOCKERHUB_USER)/serverledge-base -f images/base-alpine/Dockerfile .
image-nodejs17ng:
	docker build -t $(DOCKERHUB_USER)/serverledge-nodejs17ng -f images/nodejs17ng/Dockerfile .

images-multi-arch:  image-python310-multi-arch image-nodejs17ng-multi-arch image-base-multi-arch
PLATFORMS=linux/amd64,linux/arm64,linux/arm/v7
image-python310-multi-arch:
	docker buildx build --platform $(PLATFORMS) -t $(DOCKERHUB_USER)/serverledge-python310 -f images/python310/Dockerfile --push .
image-base-multi-arch:
	docker buildx build --platform $(PLATFORMS) -t $(DOCKERHUB_USER)/serverledge-base -f images/base-alpine/Dockerfile --push .
image-nodejs17ng-multi-arch:
	docker buildx build --platform $(PLATFORMS) -t $(DOCKERHUB_USER)/serverledge-nodejs17ng -f images/nodejs17ng/Dockerfile --push .

push-images:
	docker push $(DOCKERHUB_USER)/serverledge-python310
	docker push $(DOCKERHUB_USER)/serverledge-base
	docker push $(DOCKERHUB_USER)/serverledge-nodejs17ng

test:
	go test -v ./...

.PHONY: serverledge serverledge-cli lb executor test images

	
