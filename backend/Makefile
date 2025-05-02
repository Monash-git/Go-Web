.PHONY: docker
docker:
	@rm go-web || true
	@go mod tidy
	@GOOS=linux GOARCH=amd64 go build -tags=k8s -o go-web .
	@docker rmi -f Mark/go-web:v0.0.1
	@docker build -t Mark/go-web:v0.0.1 .