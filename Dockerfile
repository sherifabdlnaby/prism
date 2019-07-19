ARG GOLANG="1.11.5"
FROM golang:${GOLANG} as builder

ARG PRISM_VERSION="dev"
ARG LIBVIPS_VERSION="8.7.4"
ARG GOLANG

# Installs libvips + required libraries
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  ca-certificates \
  automake build-essential curl \
  gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg62-turbo-dev libpng-dev \
  libwebp-dev libtiff5-dev libgif-dev libexif-dev libxml2-dev libpoppler-glib-dev \
  swig libmagickwand-dev libpango1.0-dev libmatio-dev libopenslide-dev libcfitsio-dev \
  libgsf-1-dev fftw3-dev liborc-0.4-dev librsvg2-dev && \
  cd /tmp && \
  curl -fsSLO https://github.com/libvips/libvips/releases/download/v${LIBVIPS_VERSION}/vips-${LIBVIPS_VERSION}.tar.gz && \
  tar zvxf vips-${LIBVIPS_VERSION}.tar.gz && \
  cd /tmp/vips-${LIBVIPS_VERSION} && \
	CFLAGS="-g -O3" CXXFLAGS="-D_GLIBCXX_USE_CXX11_ABI=0 -g -O3" \
    ./configure \
    --disable-debug \
    --disable-dependency-tracking \
    --disable-introspection \
    --disable-static \
    --enable-gtk-doc-html=no \
    --enable-gtk-doc=no \
    --enable-pyvips8=no && \
  make && \
  make install && \
  ldconfig

# Installing golangci-lint
WORKDIR /tmp
RUN curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOPATH}/bin" v1.16.0

# Enable GO Modules
ENV GO111MODULE=on

WORKDIR ${GOPATH}/src/github.com/sherifabdlnaby/prism/

COPY go.mod .
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Copy prism sources
COPY . .

WORKDIR ${GOPATH}/src/github.com/sherifabdlnaby/prism/cmd/

# Compile prism
RUN go build -a -o ${GOPATH}/bin/prism -ldflags="-s -w -h -X main.Version=${PRISM_VERSION}"

FROM debian:stretch-slim

ARG PRISM_VERSION

LABEL maintainer="sherifabdlnaby@gmail.com" \
      org.label-schema.description="Image Processing Engine" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="https://github.com/sherifabdlnaby/prism" \
      org.label-schema.vcs-url="https://github.com/sherifabdlnaby/prism" \
      org.label-schema.version="${PRISM_VERSION}"

COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# Install runtime dependencies
RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install --no-install-recommends -y \
  libglib2.0-0 libjpeg62-turbo libpng16-16 libopenexr22 \
  libwebp6 libwebpmux2 libtiff5 libgif7 libexif12 libxml2 libpoppler-glib8 \
  libmagickwand-6.q16-3 libpango1.0-0 libmatio4 libopenslide0 \
  libgsf-1-114 fftw3 liborc-0.4 librsvg2-2 libcfitsio5 && \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /go/bin/prism /usr/local/bin/prism

## Server port to listen
#ENV PORT 9000

WORKDIR /usr/local/bin/

# Run the entrypoint command by default when the container starts.
ENTRYPOINT ["/usr/local/bin/prism"]

## Expose the server TCP port
#EXPOSE ${PORT}
