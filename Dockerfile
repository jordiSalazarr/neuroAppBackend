# ---- Stage 0: OpenCV 4.10 + contrib + aruco (precompilado) ----
FROM gocv/opencv:4.10.0 AS opencv

# ---- Stage 1: Builder con Go 1.24.3 oficial ----
FROM golang:1.24.3-bookworm AS builder

# Copiamos SOLO lo que necesitamos de OpenCV para compilar (no pisamos /usr/local/go)
COPY --from=opencv /usr/local/include/opencv4 /usr/local/include/opencv4
COPY --from=opencv /usr/local/lib /usr/local/lib
# (opcional) si tu opencv4.pc estuviera en /usr/local/share/pkgconfig:
# COPY --from=opencv /usr/local/share/pkgconfig /usr/local/share/pkgconfig

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH=/usr/local/lib
ENV GOEXPERIMENT=           
ENV GOTOOLCHAIN=auto        

# Toolchain extra para compilar CGO
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential pkg-config ca-certificates git && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Asegura versión de gocv compatible con OpenCV 4.10
RUN go get gocv.io/x/gocv@v0.41.0

COPY . .

# Diagnóstico: deben salir 1.24.3 y 4.10.x
RUN go version
RUN pkg-config --modversion opencv4

# Compila binario
RUN go build -trimpath -ldflags="-s -w" -o /app/server .

# ---- Stage 2: Runtime (mantenemos misma base para evitar incompatibilidades) ----
FROM gocv/opencv:4.10.0 AS runtime
ENV LD_LIBRARY_PATH=/usr/local/lib

WORKDIR /app
COPY --from=builder /app/server /app/server

EXPOSE 8080
ENV GIN_MODE=release
CMD ["/app/server"]
