FROM golang:1.16

WORKDIR /app
ADD	. /app
RUN go mod download
RUN	go build -o tf-registry-generator -ldflags '-extldflags "-static"' .


FROM alpine:3
RUN apk add --no-cache gnupg gcompat
COPY --from=0 /app/tf-registry-generator /usr/local/bin/

ENTRYPOINT 	["/usr/local/bin/tf-registry-generator"]
