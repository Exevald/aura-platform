# Build locally first: mise run //modules/user:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/user /user
EXPOSE 8081 8082
ENTRYPOINT ["/user"]
CMD ["service"]
