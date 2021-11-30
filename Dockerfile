FROM golang:1.16 as builder
WORKDIR /go/src/app
COPY . .
RUN make manager

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
COPY --from=builder /go/src/app/manager .
USER 1000
ENTRYPOINT ["/manager"]
