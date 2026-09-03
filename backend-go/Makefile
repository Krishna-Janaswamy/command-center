.PHONY: build local deploy

build:
	GOOS=linux GOARCH=amd64 go build -o bin/auth ./cmd/auth

local:
	# Run local HTTP server (not Lambda)
	FRONTEND_ORIGIN=http://localhost:8080 JWT_SECRET=change-this-secret go run ./cmd/local

deploy:
	# Deploy using Serverless Framework
	npx serverless deploy --stage dev
