FROM golang:latest AS builder

WORKDIR /go/src/app

COPY . .

RUN CGO_ENABLED=0 go build -o webecho . && \
	chmod 755 ./webecho

###########################
FROM scratch

COPY --from=builder /go/src/app/webecho .

expose 8080
expose 8081

CMD ["/webecho"]