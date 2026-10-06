# Build locally first: mise run //modules/auditor:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/auditor /auditor
EXPOSE 8081 8082
ENTRYPOINT ["/auditor"]
CMD ["service"]
