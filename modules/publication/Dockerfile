# Build locally first: mise run //modules/publication:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/publication /publication
EXPOSE 8082
ENTRYPOINT ["/publication"]
CMD ["service"]
