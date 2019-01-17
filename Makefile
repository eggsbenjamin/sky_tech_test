.SILENT:

all: docker_build

build:
	printf "\n\tBuilding binary\n\n"
	mkdir -p ./bin && GOOS=linux go build -o ./bin/main ./*.go

docker_build: build
	printf "\n\tBuilding container\n\n"
	docker build -t eggsbenjamin/order_process_service .
	docker push eggsbenjamin/order_process_service

