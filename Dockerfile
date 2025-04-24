FROM golang:1.24 AS builder
WORKDIR /go/src/app
# Install upx
ARG upx_version=4.2.4
ARG TARGETARCH=${TARGETARCH:-amd64}

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends xz-utils && \
  curl -Ls https://github.com/upx/upx/releases/download/v${upx_version}/upx-${upx_version}-${TARGETARCH}_linux.tar.xz -o - | tar xvJf - -C /tmp && \
  cp /tmp/upx-${upx_version}-${TARGETARCH}_linux/upx /usr/local/bin/ && \
  chmod +x /usr/local/bin/upx && \
  apt-get remove -y xz-utils && \
  rm -rf /var/lib/apt/lists/*

COPY cmd cmd
COPY Makefile Makefile
COPY go.mod go.mod
RUN  make small-binary
RUN chmod +x bin/hsr

FROM debian:bookworm-slim AS runner
# Create appuser and group
RUN groupadd -r appgroup && useradd -r -g appgroup appuser
COPY --from=builder --chown=appuser:appgroup /go/src/app/bin/hsr /hsr
COPY  --chown=appuser:appgroup entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh
USER appuser
ENTRYPOINT ["/entrypoint.sh"]