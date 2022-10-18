ARG build_image="golang:1.19-alpine3.16"
ARG final_image="alpine:3.16.2"
ARG os="linux"
ARG arch="amd64"
ARG version="development"

FROM --platform=$os/$arch $build_image as builder

ARG os
ARG version

RUN mkdir /build
WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GOOS=$os \
    go build -a -o archie -ldflags "-X archie/archie.Version=$version" .

FROM --platform=$os/$arch $final_image as final

ARG version

LABEL version="$version"

RUN apk add --update curl lsof tini su-exec
RUN rm /var/cache/apk/*

RUN adduser -D app -s /sbin/nologin
RUN mkdir /app

COPY --from=builder /build/archie /app/

RUN chown -R app:app /app/

WORKDIR /app

ENTRYPOINT ["/sbin/tini", "-g", "--"]

CMD ["su-exec", "app", "./archie"]
