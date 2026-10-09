# Exact versions only: every image is pinned by tag and digest.
# The builder's Go version is the one the tests run on, so the shipped binary is
# built by the same toolchain that proved it works; the widget's Node version is
# the one its tests run on, and moves with it.

# The widget comes first: the Go build embeds the one page it leaves.
FROM node:24.21.0-trixie-slim@sha256:8ec5d7557396cfe32d21c3f9c13072355ceab22b584578ca4bb28af31120cffe AS widget

WORKDIR /src/web

# The packages first, exactly as the lockfile pins them, so that this layer
# survives a change to the widget's sources. No package runs an install step of
# its own: the .npmrc beside the lockfile says so, and the install says it again
# where the step is read.
COPY web/package.json web/package-lock.json web/.npmrc ./
RUN npm ci --ignore-scripts --no-audit --no-fund

COPY web/ ./
# The design tokens live beside the page the build writes, in the package that
# embeds both, and the widget's styles import them from there.
COPY internal/widget/tokens.css /src/internal/widget/tokens.css
# The countries a parent can choose on the form are the ones the service takes,
# read from the one file both are built from.
COPY internal/domain/country/places.json /src/internal/domain/country/places.json
# The topics a card offers to keep the lessons to are the catalog's, in the
# groups of the site's page of topics, read from the files those are built from.
COPY content/catalogs/topics.json /src/content/catalogs/topics.json
COPY site/data.json /src/site/data.json

# The widget tells a host the build it came with.
ARG VERSION=dev
RUN VITE_VERSION="${VERSION}" npm run --silent build:widget

FROM golang:1.27.2-trixie@sha256:e58d6f83b3416618d8bcac2b3dde1b7f7e3c4a77d25e88637f8bbae81536c48d AS build

WORKDIR /src

# Dependencies first: they change far less often than the code, so this layer
# survives most rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=widget /src/internal/widget/widget.html internal/widget/widget.html

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

# The database of countries the sign-in looks an address up in: one month of
# DB-IP's IP to Country Lite, kept unchanged in an image of its own and pinned
# like any other. A new month is published and pinned with just
# countries-publish.
FROM ghcr.io/mathtrail/mathtrail-standalone/dbip-country-lite:2026-10@sha256:7b263f89896b0a802ac1f4e40280e092244b18e9590cf288e1933c99278dba6f AS countries

# The runtime holds the binary, a certificate bundle and the database of
# countries, and nothing else: no shell, no package manager, and a non-root
# user by default (uid 65532).
FROM gcr.io/distroless/static-debian13@sha256:58133991db06659feaabe0f4e97a35cebf15ef4ea08f8a4c6d2ee5f75e4aa6a0

COPY --from=build /server /server
COPY --from=countries /dbip-country-lite.mmdb /usr/share/mathtrail/dbip-country-lite.mmdb
ENV MATHTRAIL_COUNTRY_DB=/usr/share/mathtrail/dbip-country-lite.mmdb

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
