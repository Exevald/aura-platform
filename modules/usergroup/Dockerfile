# Build locally first: mise run //modules/usergroup:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/usergroup /usergroup
EXPOSE 8081 8082
ENTRYPOINT ["/usergroup"]
CMD ["service"]
