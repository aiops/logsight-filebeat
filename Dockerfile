# docker build -t logsight/logsight-filebeat .
FROM golang:1.17.0-alpine as build
RUN apk --no-cache add curl bash git mercurial gcc g++ docker musl-dev
RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.35-r0/glibc-2.35-r0.apk
RUN apk --no-cache add glibc-2.35-r0.apk
WORKDIR /build
ENV GO111MODULE=on

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
