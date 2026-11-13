# Build locally first: mise run //modules/filestorage:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/filestorage /filestorage
EXPOSE 8081 8082
ENTRYPOINT ["/filestorage"]
CMD ["service"]
