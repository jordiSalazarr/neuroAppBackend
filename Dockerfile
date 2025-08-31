# syntax=docker/dockerfile:1

ARG OPENCV_TAG=4.10.0
ARG GO_VERSION=1.24.3    # subimos a la última 1.24.x

# -------- Build stage --------
FROM gocv/opencv:${OPENCV_TAG} AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates git pkg-config build-essential \
 && rm -rf /var/lib/apt/lists/*

# Instala Go
ARG GO_VERSION
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
 | tar -C /usr/local -xz

WORKDIR /app

# Evita problemas de flags heredados y desactiva SwissMap en build
ENV GOFLAGS=
ENV GOEXPERIMENT=noswissmap
# Alternativa si prefieres que Go resuelva toolchain exacto del go.mod:
# ENV GOTOOLCHAIN=auto

# Dependencias primero (mejor capa caché)
COPY go.mod go.sum ./
RUN go mod download

# Código
COPY . .

# CGO + OpenCV visibles
ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig

# Ajusta si tu main no está en el root (., ./cmd/api, etc.)
ARG MAIN_PKG=.
RUN go clean -cache -modcache
RUN go build -trimpath -ldflags="-s -w" -o /app/server "$MAIN_PKG"

# -------- Runtime stage --------
FROM gocv/opencv:${OPENCV_TAG} AS runtime

RUN useradd -u 10001 -m -s /usr/sbin/nologin appuser
WORKDIR /app

COPY --from=builder /app/server /app/server
# COPY internal/migrations ./migrations   # si aplican

ENV GIN_MODE=release
ENV PORT=8401
EXPOSE 8401

USER appuser
CMD ["/app/server"]
