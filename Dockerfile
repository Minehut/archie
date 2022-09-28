# builder image
FROM --platform=linux/amd64 golang:1.18-alpine3.16 as builder

RUN mkdir /build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o archie .

# final image for end users
FROM --platform=linux/amd64 alpine:3.16.2

RUN apk add --update \
                curl \
                lsof \
                tini \
                su-exec

RUN rm /var/cache/apk/*

RUN adduser -D app -s /sbin/nologin

RUN mkdir /app/

COPY --from=builder /build/archie /app/

RUN chown -R app:app /app/

WORKDIR /app/

ENTRYPOINT ["/sbin/tini", "-g", "--"]

CMD ["su-exec", "app", "./archie"]
