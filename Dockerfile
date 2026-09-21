# Exact versions only: both images are pinned by tag and digest.
# The builder's Go version is the one the tests run on, so the shipped binary is
# built by the same toolchain that proved it works.

FROM golang:1.27.1-trixie@sha256:433790e515d27dc6003e847e644cc0af956985cf315c1c58a3b73ee2dd305183 AS build

WORKDIR /src

# Dependencies first: they change far less often than the code, so this layer
# survives most rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# The build identity comes in as arguments rather than being read from git:
# the build context has no .git, and a version the image guesses is worse than
# one it is told (internal/version).
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# CGO off so the binary is static and can run on a distroless image with no
# libc; -trimpath so the paths of whoever built it are not in the binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/MathTrail/mathtrail-standalone/internal/version.Version=${VERSION} \
      -X github.com/MathTrail/mathtrail-standalone/internal/version.Commit=${COMMIT} \
      -X github.com/MathTrail/mathtrail-standalone/internal/version.Date=${DATE}" \
    -o /server ./cmd/server

# The runtime holds the binary, a certificate bundle and nothing else: no
# shell, no package manager, and a non-root user by default (uid 65532).
FROM gcr.io/distroless/static-debian13@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3

COPY --from=build /server /server

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
