# docker build -t logsight/filebeat .
FROM golang:1.17.0 as build
RUN apt-get update && apt-get install -y curl bash mercurial gcc g++ docker musl-dev gcc-aarch64-linux-gnu
WORKDIR /build
ENV GO111MODULE=on

# Copy go.mod first and download dependencies, to enable the Docker build cache
COPY go.mod .
RUN go mod download

# Copy rest of the source code and build
# Delete go.sum files and clean go.mod files form local 'replace' directives
COPY . .
RUN go build -ldflags="-w -s" -o "build/filebeat" "./filebeat"

FROM golang:1.17.0
WORKDIR /
COPY --from=build /build/build/filebeat /
ENTRYPOINT ["/filebeat", "-e", "--strict.perms=false"]
