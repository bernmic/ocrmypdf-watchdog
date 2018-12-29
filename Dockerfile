FROM golang:alpine as builder
RUN apk update && apk add --no-cache git
COPY . $GOPATH/src/go2music/
WORKDIR $GOPATH/src/go2music/
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/main .
RUN cp -r $GOPATH/src/go2music/assets /go/bin
RUN cp -r $GOPATH/src/go2music/static /go/bin
FROM jbarlow83/ocrmypdf:latest
COPY --from=builder /go/bin/main /app/
WORKDIR /app
VOLUME /in /out
CMD ["./main"]
