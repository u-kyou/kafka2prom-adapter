# Build 
FROM golang:1.21.3 as build_context

# About librdkafka
# Prebuilt librdkafka binaries are included with the Go client and librdkafka does not need to be installed separately on the build or target system. The following platforms are supported by the prebuilt librdkafka binaries:
# - Mac OSX x64 and arm64
# - glibc-based Linux x64 and arm64 (e.g., RedHat, Debian, CentOS, Ubuntu, etc) - without GSSAPI/Kerberos support
# - musl-based Linux amd64 and arm64 (Alpine) - without GSSAPI/Kerberos support
# - Windows amd64 - without GSSAPI/Kerberos support
# When building your application for Alpine Linux (musl libc) you must pass -tags musl to go get, go build, etc.
# 
# CGO_ENABLED must NOT be set to 0 since the Go client is based on the C library librdkafka.
# 
# If GSSAPI/Kerberos authentication support is required you will need to install librdkafka separately, see the Installing librdkafka chapter below, and then build your Go application with -tags dynamic.
# 
# Installing librdkafka: https://github.com/confluentinc/confluent-kafka-go?tab=readme-ov-file#installing-librdkafka
# * RUN apt-get update; apt-get install -y --no-install-recommends librdkafka-dev ; rm -rf /var/lib/apt/lists/*

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /workspace
ADD . /workspace/

RUN sh deploy/build.sh cmd/main.go
RUN mv /workspace/bin/kafka2prom-adapter ${OUTDIR}/usr/local/bin/

#############
# Final image
#############

FROM ubuntu:22.04

COPY --from=build_context /out/ /

WORKDIR /

EXPOSE 8081
CMD ["/usr/local/bin/kafka2prom-adapter"]
