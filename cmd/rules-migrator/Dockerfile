FROM golang:1.22.2-bookworm as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
ENV GOPROXY=https://proxy.golang.org
RUN make clean && make rules-migrator

FROM alpine:3.19.1
RUN apk add --update --no-cache ca-certificates
COPY --from=build /build_dir/cmd/rules-migrator/rules-migrator /usr/bin/rules-migrator
EXPOSE 80
ENTRYPOINT [ "/usr/bin/rules-migrator" ]
