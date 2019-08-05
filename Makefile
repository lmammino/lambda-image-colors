.PHONY: test
test: lint unit build integration

.PHONY: lint
lint:
	GO111MODULE=on go vet -mod=vendor ./...
	golint --set_exit_status cmd/...

.PHONY: unit
unit:
	GO111MODULE=on go test -mod=vendor -cover -v -short ./...

.PHONY: build
build:
	go mod tidy
	go mod vendor
	mkdir -p build
	GOOS=linux go build -mod=vendor -o build/image-colors ./cmd/image-colors-lambda
	cd build && zip image-colors.zip ./image-colors
	echo "build/image-colors.zip created"

.PHONY: integration
integration:
	tests/integration/run.sh

.PHONY: deploy
deploy:
	cd stack && \
	terraform plan -out plan && \
	terraform apply plan && \
	rm plan
