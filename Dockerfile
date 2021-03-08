FROM golang:1.15

WORKDIR /app
ADD	. /app
RUN go mod download
RUN	go build -o tf-registry-generator -ldflags '-extldflags "-static"' .


FROM alpine:3
RUN apk add --no-cache gnupg
COPY --from=0 /app/tf-registry-generator /usr/local/bin/

ENTRYPOINT 	["/usr/local/bin/tf-registry-generator"]
