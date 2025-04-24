FROM golang:1.24 AS builder
WORKDIR /go/src/app
# Install upx
RUN apt-get update && apt-get install -y upx-ucl && rm -rf /var/lib/apt/lists/*

COPY cmd cmd
COPY Makefile Makefile
COPY go.mod go.mod
RUN  make small-binary
RUN chmod +x bin/hsr

FROM debian:bookworm-slim AS runner
COPY --from=builder --chown=appuser:appgroup /go/src/app/bin/hsr /hsr
COPY  --chown=appuser:appgroup entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh
USER appuser
ENTRYPOINT ["/entrypoint.sh"]