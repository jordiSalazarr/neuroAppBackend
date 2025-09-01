# ---- Stage 0: OpenCV 4.10 + contrib + aruco (precompilado) ----
FROM --platform=$BUILDPLATFORM gocv/opencv:4.10.0 AS opencv

# ---- Stage 1: Builder con Go 1.24.3 oficial ----
FROM --platform=$BUILDPLATFORM golang:1.24.3-bookworm AS builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "BUILDPLATFORM=${BUILDPLATFORM} TARGETPLATFORM=${TARGETPLATFORM}"

# Copiamos SOLO headers y libs (no pisamos /usr/local/go)
COPY --from=opencv /usr/local/include/opencv4 /usr/local/include/opencv4
COPY --from=opencv /usr/local/lib /usr/local/lib
# Si tu opencv4.pc está en share, descomenta:
# COPY --from=opencv /usr/local/share/pkgconfig /usr/local/share/pkgconfig

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH=/usr/local/lib
ENV GOEXPERIMENT=
ENV GOTOOLCHAIN=local
# Importante en Railway: bajar concurrencia de build
ENV GOFLAGS=-p=1

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential pkg-config ca-certificates git g++ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
# deps primero para caché estable
COPY go.mod go.sum ./
RUN go mod download

# Asegúrate en tu go.mod:
#   require gocv.io/x/gocv v0.41.0
#   go 1.24
#   toolchain go1.24.3
# (mejor fijarlo en el go.mod y hacer go mod tidy antes de subir)

COPY . .

# Diagnóstico rápido
RUN go version
RUN pkg-config --modversion opencv4
RUN pkg-config --cflags --libs opencv4

# Sanity check: enlaza con aruco
RUN printf '#include <opencv2/core.hpp>\n#include <opencv2/aruco.hpp>\nint main(){return 0;}\n' > /tmp/t.cc \
 && g++ /tmp/t.cc $(pkg-config --cflags --libs opencv4) -o /tmp/t && /tmp/t

# Compila con logs verbosos (y baja concurrencia via GOFLAGS)
RUN GOFLAGS="-x -v -p=1" go build -trimpath -ldflags="-s -w" -o /app/server .

# ---- Stage 2: Runtime con mismas libs OpenCV ----
FROM --platform=$TARGETPLATFORM gocv/opencv:4.10.0 AS runtime
ENV LD_LIBRARY_PATH=/usr/local/lib

WORKDIR /app
COPY --from=builder /app/server /app/server

# Railway suele esperar que escuches en $PORT
ENV PORT=8401
EXPOSE 8401
ENV GIN_MODE=release
CMD ["/app/server"]
