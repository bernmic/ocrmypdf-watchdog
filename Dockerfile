FROM golang:alpine as builder
RUN apk update && apk add --no-cache git
COPY . $GOPATH/src/ocrmypdf-watchdog/
WORKDIR $GOPATH/src/ocrmypdf-watchdog/
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/main .
FROM jbarlow83/ocrmypdf:latest
COPY --from=builder /go/bin/main /app/
WORKDIR /app
VOLUME /in /out
ENTRYPOINT ["/app/main"]
# ENTRYPOINT ["sh", "-c", "ls -l /app"]
