# docker build -t logsight/logsight-filebeat .
FROM golang:1.17.0-alpine as build
RUN apk --no-cache add curl bash git mercurial gcc g++ docker musl-dev glibc-static.x86_64
WORKDIR /build
ENV GO111MODULE=on
ENV GOARCH=arm64

# Copy go.mod first and download dependencies, to enable the Docker build cache
COPY go.mod .
RUN go mod download

# Copy rest of the source code and build
# Delete go.sum files and clean go.mod files form local 'replace' directives
COPY . .
RUN go build -o "build/filebeat" "./filebeat"

FROM golang:1.17.0-alpine
WORKDIR /
COPY --from=build /build/build/filebeat /
ENTRYPOINT ["/filebeat", "-e", "--strict.perms=false"]
