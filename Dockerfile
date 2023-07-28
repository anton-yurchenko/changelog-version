FROM golang:1.20 as builder
WORKDIR /opt/src
COPY . .
RUN groupadd -g 1000 appuser &&\
    useradd -m -u 1000 -g appuser appuser
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /opt/app

FROM busybox:1.36
LABEL org.opencontainers.image.source="https://github.com/anton-yurchenko/changelog-version"
LABEL org.opencontainers.image.version="v1.0.0"
LABEL org.opencontainers.image.authors="Anton Yurchenko <anton.doar@gmail.com>"
LABEL org.opencontainers.image.licenses="MIT"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY LICENSE.md /LICENSE.md
COPY --from=builder --chown=1000:0 /opt/app /app
ENTRYPOINT [ "/app" ]
