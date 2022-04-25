FROM golang:1.17.7-alpine3.14 as gobuild
RUN apk add -U --no-cache build-base ca-certificates
COPY . /src
RUN cd /src && go build -ldflags="-s -w" -v -o ./k8s-platform-lcm cmd/lcm/main.go

FROM scratch
COPY --from=gobuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuild /src/k8s-platform-lcm /lcm
COPY templates /templates
COPY static /static
ENTRYPOINT ["/lcm"]
