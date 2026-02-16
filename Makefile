.PHONY: papiauth

papiauth:
	cd Api_Gateway && protoc --go_out=. --go-grpc_out=. ./pkg/proto/auth_subscription.proto
