FROM golang:1.19-alpine3.17 as builder

WORKDIR /streams-egress-proxy-agent

RUN apk update && apk add --no-cache git

COPY go.mod go.sum /streams-egress-proxy-agent/
RUN go mod download
RUN go mod verify

ADD . .

RUN CG0_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /egress-proxy-agent ./cmd/proxy-agent/main.go

FROM alpine:3.17 as runtime

LABEL maintainer = "Alexandria OpenSource <oss@thealexandriaproject.com>"
LABEL org.name="The Alexandria Project"
LABEL org.image.title="Egress Proxy Agent for Alexandria Streams"
LABEL org.image.description="An traffic egress proxy agent for Alexandria Streams toolkit, forwards buffered streams traffic from a database to actual streaming infrastructure"

# Required by health checks
RUN apk --no-cache add curl

RUN adduser \
        -g "log daemon user" \
        --disabled-password \
        "egress-proxy-agent"

COPY --from=builder /egress-proxy-agent /usr/local/bin/egress-proxy-agent

RUN chown -R "egress-proxy-agent":"egress-proxy-agent" /var
RUN chmod 766 /var
USER "egress-proxy-agent":"egress-proxy-agent"

#EXPOSE 52450

ENTRYPOINT ["/usr/local/bin/egress-proxy-agent"]
