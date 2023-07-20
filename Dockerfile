FROM docker.io/library/golang:1.19 AS builder
RUN mkdir /apid-helper
COPY . /apid-helper
WORKDIR /apid-helper
ENV GO111MODULE=on
RUN make apid-build

FROM alpine:3.18.2
COPY --from=builder /apid-helper/dist/apid ./bin
ENTRYPOINT [ "/bin/apid" ]