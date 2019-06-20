OK_COLOR=\033[32;01m
NO_COLOR=\033[0m

build:
	@echo "$(OK_COLOR)==> Compiling binary$(NO_COLOR)"
	go build -a -o ./bin/prism -ldflags="-s -w -h -X" ./cmd/

test:
	go test

install:
	go mod verify && go mod download
#	INSTALL VIPS

docker-build:
	@echo "$(OK_COLOR)==> Building Docker image$(NO_COLOR)"
	docker build --tag=prism .

docker-run:
	@echo "$(OK_COLOR)==> Pushing Docker image v$(VERSION) $(NO_COLOR)"
	docker-compose up -d

docker-down:
	@echo "$(OK_COLOR)==> Pushing Docker image v$(VERSION) $(NO_COLOR)"
	docker-compose down
