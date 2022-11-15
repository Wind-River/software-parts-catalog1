FROM quay.io/minio/minio:RELEASE.2021-12-10T23-03-39Z

COPY object-storage-entrypoint.sh /usr/bin/
RUN chmod +x /usr/bin/object-storage-entrypoint.sh

ENTRYPOINT [ "/usr/bin/object-storage-entrypoint.sh" ]