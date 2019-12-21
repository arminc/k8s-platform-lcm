FROM scratch

ARG VERSION

COPY dist/lcm-$VERSION-linux /lcm
COPY templates/* templates/*
COPY static/* static/*

ENTRYPOINT ["/lcm"]