FROM alpine:3
RUN apk --no-cache add ca-certificates && update-ca-certificates
COPY k8s-platform-lcm /lcm
COPY templates /templates
COPY static /static
ENTRYPOINT ["/lcm"]
