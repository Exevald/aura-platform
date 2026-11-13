# Build locally first: mise run //modules/realtime:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/realtime /realtime
EXPOSE 8081 8082
ENTRYPOINT ["/realtime"]
CMD ["service"]
