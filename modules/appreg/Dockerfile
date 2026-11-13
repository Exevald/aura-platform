# Build locally first: mise run //modules/appreg:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/appreg /appreg
COPY config/apps.yaml /config/apps.yaml
EXPOSE 8081 8082
ENTRYPOINT ["/appreg"]
CMD ["service"]
