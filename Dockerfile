FROM scratch
COPY k8s-platform-lcm /lcm
COPY templates /templates
COPY static /static
ENTRYPOINT ["/lcm"]