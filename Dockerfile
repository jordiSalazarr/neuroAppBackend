# syntax=docker/dockerfile:1

# ==== Versions ====
ARG OPENCV_TAG=4.10.0      # versión del runtime con OpenCV
ARG GO_VERSION=1.22.5

# ==== Build stage: OpenCV + Go toolchain + CGO ====
FROM gocv/opencv:${OPENCV_TAG} AS builder

# Herramientas de compilación
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates git pkg-config build-essential \
 && rm -rf /var/lib/apt/lists/*

# Instala Go oficial
ARG GO_VERSION
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
 | tar -C /usr/local -xz

WORKDIR /app

# Copia módulos primero para aprovechar caché
COPY go.mod go.sum ./
RUN go mod download

# Copia el resto del código
COPY . .

# CGO habilitado y pkg-config apuntando a OpenCV del contenedor
ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig

# Verificación opcional (útil para debug de compilación)
# RUN pkg-config --modversion opencv4 && go env CGO_ENABLED

# Compila tu binario (ajusta el path del main)
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ==== Runtime: misma base con OpenCV ====
FROM gocv/opencv:${OPENCV_TAG} AS runtime

# Usuario no root
RUN useradd -u 10001 -m -s /usr/sbin/nologin appuser

WORKDIR /app
COPY --from=builder /app/server /app/server
# Si tienes migraciones, descomenta:
# COPY internal/migrations ./migrations

ENV GIN_MODE=release
ENV PORT=8401
EXPOSE 8401

USER appuser
CMD ["/app/server"]
