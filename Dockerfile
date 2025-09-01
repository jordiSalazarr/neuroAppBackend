# ---- Stage 0: OpenCV 4.10 + contrib + aruco (precompilado) ----
FROM --platform=$BUILDPLATFORM gocv/opencv:4.10.0 AS opencv

# ---- Stage 1: Builder con Go 1.24.3 oficial ----
FROM --platform=$BUILDPLATFORM golang:1.24.3-bookworm AS builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "BUILDPLATFORM=${BUILDPLATFORM} TARGETPLATFORM=${TARGETPLATFORM}"

# Copiamos SOLO lo necesario de OpenCV (no pisamos /usr/local/go)
COPY --from=opencv /usr/local/include/opencv4 /usr/local/include/opencv4
COPY --from=opencv /usr/local/lib /usr/local/lib
# (si tu opencv4.pc estuviera en share, descomenta también)
# COPY --from=opencv /usr/local/share/pkgconfig /usr/local/share/pkgconfig

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH=/usr/local/lib
ENV GOEXPERIMENT=
ENV GOTOOLCHAIN=local

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential pkg-config ca-certificates git g++ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
# Descarga deps primero para caché estable
COPY go.mod go.sum ./
RUN go mod download

# IMPORTANTE: fija gocv en go.mod a una versión compatible con OpenCV 4.10
# En tu go.mod debe haber:   require gocv.io/x/gocv v0.41.0
# (Mejor fijarlo en el go.mod; evitamos 'go get' en Dockerfile para menos red.)
# Si no lo tienes aún, añade la línea en tu go.mod y vuelve a 'go mod download'

COPY . .

# Diagnósticos útiles: versión de Go y de OpenCV visible para pkg-config
RUN go version
RUN pkg-config --modversion opencv4
RUN pkg-config --cflags --libs opencv4

# Sanity check: compila un hello-world C++ enlazando a aruco
RUN printf '#include <opencv2/core.hpp>\n#include <opencv2/aruco.hpp>\nint main(){return 0;}\n' > /tmp/t.cc \
 && g++ /tmp/t.cc $(pkg-config --cflags --libs opencv4) -o /tmp/t && /tmp/t

# Compila tu binario con logs verbosos para depurar en Railway
RUN GOFLAGS="-x -v" go build -trimpath -ldflags="-s -w" -o /app/server .

# ---- Stage 2: Runtime (misma base con OpenCV para evitar misses) ----
FROM --platform=$TARGETPLATFORM gocv/opencv:4.10.0 AS runtime
ENV LD_LIBRARY_PATH=/usr/local/lib
WORKDIR /app
COPY --from=builder /app/server /app/server

EXPOSE 8080
ENV GIN_MODE=release
CMD ["/app/server"]
