.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# UNIVERSE ###

## build/up:
.PHONY: up/build
up/build: dl/build mosaic/build
	@echo "Stopping docker images (if running)..."
	docker-compose down
	@echo "Starting docker images..."
	docker-compose up --build -d
	@echo "Docker images started!"

## down:
.PHONY: down
down:
	@echo "Stopping docker images..."
	docker-compose down
	@echo "Done"


#BROKER-SERVICE ###

## broker/build: Build the broker service binary
.PHONY: broker/build
broker/build:
	@echo "Building broker service..."
	cd ./broker-service && env GOOS=linux go build -o ./build/linux/brokerApp ./cmd/api
	@echo "Done!"

## broker/build/up
.PHONY: broker/build/up
broker/build/up: broker/build
	@echo "Start"
	docker-compose up -d --build broker-service
	@echo "Docker image started!"

## broker/down
.PHONY:broker/down
broker/down:
	@echo "Stopping broker service docker container..."
	docker-compose down broker-service
	@echo "Done"


# DOWNLOADER-SERVICE ###

## dl/build: Build downloader service binary
.PHONY: dl/build
dl/build:
	@echo "Building downloader service..."
	cd ./downloader-service && env GOOS=linux go build -o ./build/linux/dlApp ./cmd/api
	@echo "Done!"

## dl/build/up: Stop downloader docker container, build binary, build image and start the container
.PHONY: dl/build/up
dl/build/up: dl/down dl/build
	@echo "Starting fresh downloader docker image..."
	docker-compose up --build -d downloader-service
	@echo "Docker image started!"

## dl/down: Stop downloader docker container
.PHONY: dl/down
dl/down:
	@echo "Stopping downloader docker container..."
	docker-compose down downloader-service
	@echo "Done"



# MOSAIC-SERVICE ###

## mosaic/build: Build the mosaic service binary
.PHONY: mosaic/build
mosaic/build:
	@echo "Building mosaic service binary..."
	cd ./mosaic-service && env GOOS=linux go build -o ./build/linux/mosaicApp ./cmd/api
	@echo "Done!"

## mosaic/build/up: Stop mosaic container, build binary, build image and start the container
.PHONY: mosaic/build/up
mosaic/build/up: mosaic/down mosaic/build
	@echo "Starting new mosaic docker image..."
	docker-compose up --build -d mosaic-service
	@echo "Docker image started!"

## mosaic/down: Stop mosaic service docker container
.PHONY: mosaic/down
mosaic/down:
	@echo "Stopping mosaic service docker container..."
	docker-compose down mosaic-service
	@echo "Done"


#REDIS ###
## redis: Start redis docker container
.PHONY: redis
redis:
	@echo "Starting redis docker image..."
	docker-compose up -d redis
	@echo "Started docker image!"


#NATS ###
## nats:
.PHONY: nats
nats:
	@echo "Starting redis docker image..."
	docker-compose up -d nats
	@echo "Started docker image!"



# TEST ENV ###
## redis/test: Start redis test docker container
.PHONY: redis/test
redis/test:
	@echo "Starting redis test docker container"
	docker run -d --rm --name redis-test -p 6378:6379 redis:8.2-alpine
	@echo "Started redis test container!"

## redis/test: Start redis test docker container
.PHONY: redis/test/down
redis/test/down:
	@echo "Stopping redis test container"
	docker stop redis-test
	@echo "Done"
