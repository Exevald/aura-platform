# Build locally first: mise run //modules/registerer:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/registerer /registerer
EXPOSE 8081 8082
ENTRYPOINT ["/registerer"]
CMD ["service"]
