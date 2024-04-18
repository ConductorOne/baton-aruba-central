FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-aruba-central"]
COPY baton-aruba-central /