FROM golang:1.20 as builder
WORKDIR /go/src/app
COPY . .
RUN make manager

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/src/app/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
