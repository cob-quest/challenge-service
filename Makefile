IMAGE_TAG=registry.gitlab.com/cs302-2023/g3-team8/project/assignment-service/main:latest
.DEFAULT_GOAL := run

# running
run:
	go run main.go

# linting
lint:
	golangci-lint run --enable-all

test:
	go test -v

# building
build:
	go build -o main main.go 

# clean
clean:
	rm ./main

# builds a docker image
docker-build:
	docker build -t ${IMAGE_TAG} -f ./docker/Dockerfile.dev .

# builds a docker image and loads into minikube
minikube: docker-build
	minikube image rm ${IMAGE_TAG} 2> /dev/null
	minikube image load ${IMAGE_TAG}

