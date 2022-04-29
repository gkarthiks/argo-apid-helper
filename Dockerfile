FROM docker.io/library/golang:1.18 AS builder
RUN mkdir /apid-helper
COPY . /apid-helper
WORKDIR /apid-helper
ENV GO111MODULE=on
RUN make apid-build-linux

FROM alpine:3.15.4
COPY --from=builder /apid-helper/dist/apid ./bin
ENTRYPOINT [ "/bin/apid" ]