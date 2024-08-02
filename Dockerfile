FROM golang:1.22.1-alpine3.19 AS build

RUN apk add --no-cache \
    build-base \
    imagemagick-dev \
    ;

WORKDIR /go/projects/resizer
COPY . .

# Stack size must be increased to avoid segmentation faults
# ImageMagick (some modules more than others) uses a lot of stack space
# and was not developped with Musl's tight stack limits in mind.
#
# https://wiki.musl-libc.org/functional-differences-from-glibc.html#Thread-stack-size
RUN CGO_ENABLED=1 CGOOS=linux CGO_CFLAGS_ALLOW='-Xpreprocessor' \
    go build \
    -tags=musl \
    -ldflags '-extldflags "-Wl,-z,stack-size=2097152"' \
    -o /go/bin/resizer

CMD "/go/bin/resizer"

FROM alpine:3.19 AS prod

RUN apk add --no-cache \
    imagemagick \
    imagemagick-jpeg \
    imagemagick-webp \
    ;

COPY --from=build /go/bin/resizer /usr/local/bin/resizer
CMD "resizer"
