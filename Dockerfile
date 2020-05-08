FROM registry.access.redhat.com/ubi8/go-toolset AS builder
COPY ./ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /opt/
COPY --from=builder /opt/app-root/src/app .
EXPOSE 8080
USER 1001
CMD ["/opt/app"] 
