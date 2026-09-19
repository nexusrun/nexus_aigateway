FROM golang:1.27.1-alpine3.24 AS builder
RUN apk add --no-cache git ca-certificates nodejs npm
ARG VERSION=nexus
ARG COMMIT=none
WORKDIR /src
COPY . .
RUN cd web/dashboard && npm install --include=dev --package-lock=false --no-audit --no-fund --ignore-scripts && npm run build
RUN CGO_ENABLED=0 go build -tags=swagger -ldflags="-s -w -X github.com/nexusrun/nexus_aigateway/internal/version.Version=${VERSION} -X github.com/nexusrun/nexus_aigateway/internal/version.Commit=${COMMIT}" -o /usr/local/bin/aigateway ./cmd/aigateway

RUN mkdir -p /out/config /out/.cache /out/data \
	&& cp /src/config/*.yaml /out/config/ \
	&& cp /usr/local/bin/aigateway /out/aigateway

FROM gcr.io/distroless/static-debian12:nonroot


COPY --from=builder /out/aigateway /aigateway
COPY --from=builder /out/config/*.yaml /app/config/
COPY --from=builder --chown=65532:65532 /out/.cache /app/.cache
COPY --from=builder --chown=65532:65532 /out/data /app/data

WORKDIR /app
ENV SWAGGER_ENABLED=true
EXPOSE 8080
ENTRYPOINT ["/aigateway"]
