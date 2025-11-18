
ENV = $(shell go env GOPATH)
GO_VERSION = $(shell go version)
BINARY_NAME=mscproject.out

lint:
	go vet ./...
	golangci-lint run
  
mscproject-build:
	echo "building mscproject"
	go build -o bin/${BINARY_NAME} app/mscproject/main.go

run-mscproject:
	echo "running mscproject"
	./bin/${BINARY_NAME}

run-back-mscproject:
	echo "running mscproject in background"
	nohup ./bin/${BINARY_NAME} &

stop-back-mscproject:
	echo "stopping mscproject in background"
	killall ${BINARY_NAME}

container-build:
	echo "building mscproject container"
	sudo docker build --tag docker-mscproject .
	
config-up:
	echo "starting up configs"
	sudo docker-compose up -d

config-down:
	echo "shuting down configs"
	sudo docker-compose down