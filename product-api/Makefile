check-install:
	which swagger || go get -u github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger: check-install
	swagger generate spec -o ./swagger.yaml --scan-models