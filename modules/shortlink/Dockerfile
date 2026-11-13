# Build locally first: mise run //modules/shortlink:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/shortlink /shortlink
EXPOSE 8081 8082
ENTRYPOINT ["/shortlink"]
CMD ["service"]
