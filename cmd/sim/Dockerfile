FROM golang:1.22.2-bookworm as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
ENV GOPROXY=https://proxy.golang.org
RUN make clean && make sim

FROM alpine:3.19.1
RUN apk add --update --no-cache ca-certificates
COPY --from=build /build_dir/cmd/sim/sim /usr/bin/sim
EXPOSE 80
ENTRYPOINT [ "/usr/bin/sim" ]
