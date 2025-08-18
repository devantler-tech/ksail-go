# Use distroless static image for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy the binary (will be provided by GoReleaser)
COPY ksail /ksail

# Use nonroot user from distroless
USER nonroot:nonroot

# Set entrypoint
ENTRYPOINT ["/ksail"]