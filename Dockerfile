FROM golang:1.12.6

RUN mkdir -p /build
WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

ADD main.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o promcon_registration_exporter

FROM alpine:3.9
RUN apk add --update ca-certificates
COPY --from=0 /build/promcon_registration_exporter /usr/bin
CMD ["/usr/bin/promcon_registration_exporter"]
