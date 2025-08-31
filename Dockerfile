# syntax=docker/dockerfile:1

ARG OPENCV_TAG=4.10.0
ARG GO_VERSION=1.22.5

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

# Dependencias primero (cacheables por capas)
COPY go.mod go.sum ./
RUN go mod download

# Código
COPY . .

# CGO + OpenCV
ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig

# Compila tu binario (ajusta el path si tu main está en otro sitio)
RUN go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# -------- Runtime stage --------
FROM gocv/opencv:${OPENCV_TAG} AS runtime

RUN useradd -u 10001 -m -s /usr/sbin/nologin appuser
WORKDIR /app

COPY --from=builder /app/server /app/server
# COPY internal/migrations ./migrations  # si las usas

ENV GIN_MODE=release
ENV PORT=8401
EXPOSE 8401

USER appuser
CMD ["/app/server"]
