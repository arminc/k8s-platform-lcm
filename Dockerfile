FROM alpine:3 as alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY k8s-platform-lcm /lcm
COPY templates /templates
COPY static /static
ENTRYPOINT ["/lcm"]
