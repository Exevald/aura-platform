# Build locally first: mise run //modules/gateway:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/gateway /gateway
EXPOSE 8081 8082
ENTRYPOINT ["/gateway"]
CMD ["service"]
