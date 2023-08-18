# syntax=docker/dockerfile:1.2

FROM golang:1.20-alpine AS builder
RUN apk add --no-cache build-base git
WORKDIR /app
COPY . .
RUN \
   --mount=type=cache,target=/go/pkg \
   --mount=type=cache,target=/root/.cache/go-build \
   go build -o solcluster .

# Copy final binary into light stage.
FROM alpine:3

ARG GITHUB_SHA=local
ENV GITHUB_SHA=${GITHUB_SHA}

COPY --from=builder /app/solcluster /usr/local/bin/

ENV USER=solcluster
ENV UID=13852
ENV GID=13852
RUN addgroup -g "$GID" "$USER"
RUN adduser \
    --disabled-password \
    --gecos "solana" \
    --home "/opt/$USER" \
    --ingroup "$USER" \
    --no-create-home \
    --uid "$UID" \
    "$USER"
RUN chown solcluster /usr/local/bin/solcluster
RUN chmod u+x /usr/local/bin/solcluster

WORKDIR "/opt/$USER"
USER solcluster
ENTRYPOINT ["/usr/local/bin/solcluster"]
CMD ["sidecar"]

LABEL org.opencontainers.image.source="https://github.com/Blockdaemon/solana-cluster"
