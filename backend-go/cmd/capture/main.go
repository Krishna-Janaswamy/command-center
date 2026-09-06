package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/example/service-virtualization-go/internal/virtualization"
)

func main() {
	store := virtualization.NewConfiguredStore(context.Background())
	mux := http.NewServeMux()
	mux.Handle("/", &virtualization.CaptureHandler{Store: store, Blobs: virtualization.NewConfiguredBlobStore(context.Background())})
	lambda.Start(httpadapter.New(mux).ProxyWithContext)
}
