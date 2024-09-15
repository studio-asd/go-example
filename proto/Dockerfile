FROM public.ecr.aws/docker/library/golang:1.23.1

WORKDIR /generate

RUN apt update \
    && apt install -y protobuf-compiler \
    && apt clean

RUN go mod init github.com/albertwidi/go-example && \
    go mod tidy && \
    go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/golang/protobuf/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

ENTRYPOINT ["bash"]
