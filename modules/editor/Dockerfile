# Build locally first: mise run //modules/editor:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/editor /editor
EXPOSE 8081 8082
ENTRYPOINT ["/editor"]
CMD ["service"]
