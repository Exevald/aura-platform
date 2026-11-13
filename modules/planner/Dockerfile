# Build locally first: mise run //modules/planner:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/planner /planner
EXPOSE 8081 8082
ENTRYPOINT ["/planner"]
CMD ["service"]
