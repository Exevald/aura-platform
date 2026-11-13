# Build locally first: mise run //modules/workspace:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/workspace /workspace
EXPOSE 8081 8082
ENTRYPOINT ["/workspace"]
CMD ["service"]
