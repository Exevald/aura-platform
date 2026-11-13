# Build locally first: mise run //modules/mailer:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/mailer /mailer
EXPOSE 8081 8082
ENTRYPOINT ["/mailer"]
CMD ["service"]
