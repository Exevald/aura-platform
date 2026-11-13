# Build locally first: mise run //modules/calendar:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/calendar /calendar
EXPOSE 8081 8082
ENTRYPOINT ["/calendar"]
CMD ["service"]
