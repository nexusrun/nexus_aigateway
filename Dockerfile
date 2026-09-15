FROM golang:1.27.1-alpine3.24 AS builder
RUN apk add --no-cache git ca-certificates nodejs npm
ARG REPO_URL=https://github.com/nexusrun/nexus_aigateway.git
ARG REPO_REF=main
ARG VERSION=nexus
WORKDIR /src
RUN git clone "${REPO_URL}" . && git checkout "${REPO_REF}"
RUN cd web/dashboard && npm ci --no-audit --no-fund --ignore-scripts && npm run build
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/nexusrun/nexus_aigateway/internal/version.Version=${VERSION} -X github.com/nexusrun/nexus_aigateway/internal/version.Commit=$(git rev-parse --short HEAD)" -o /usr/local/bin/gomodel ./cmd/gomodel

RUN mkdir -p /out/config /out/.cache /out/data \
	&& cp /src/config/*.yaml /out/config/ \
	&& cp /usr/local/bin/gomodel /out/gomodel

FROM gcr.io/distroless/static-debian12:nonroot


COPY --from=builder /out/gomodel /gomodel
COPY --from=builder /out/config/*.yaml /app/config/
COPY --from=builder --chown=65532:65532 /out/.cache /app/.cache
COPY --from=builder --chown=65532:65532 /out/data /app/data

WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/gomodel"]
