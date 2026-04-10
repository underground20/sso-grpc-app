generate-proto:
	protoc -I contract ./contract/sso/sso.proto --go_out=./contract/gen --go_opt=paths=source_relative --go-grpc_out=./contract/gen --go-grpc_opt=paths=source_relative