# Docker Services
The compose.yml services uses the following directories as build directories or simply configuration files for mounting.

## main
The main server for the Software Parts Catalog.

## db
The PostgreSQL instance used by main, for the main data storage, and potentially blob metadata.
Can be replaced by any PostgreSQL instance.
The data and blob databeses are configured separately, but can point to the same instance.

## nginx
Nginx is the entrypoint to the system.
Nginx serves the Vue frontend, and reverse proxies the main server.

## object-storage
A minio object storage configuration for storing archives and files.
Can be replaced with any S3-compatible service.