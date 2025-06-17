FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-sonatype-nexus"]
COPY baton-sonatype-nexus /