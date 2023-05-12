
# Software Parts Catalog
## Overview
Maintaining a catalog of software components (parts) is a requirement whether generating SBOMs for managing license compliance, security assurance, export control, or safety certification. We developed a highly searchable scalable software parts database (catalog) enabling companies to seamlessly manage 1000s, 10,000s if not 100,000s of software parts from which their products are comprised.

## Software Parts 
There are different ways to define what a software part is. In the most basic sense the different parts represent the building blocks from which software solutions are comprised. Software parts stored in the catalog can consist of any of the following levels of granulatity:  
- a single file (source or binary)
- a collection for files (e.g., archive, package)
- application/program/library binary
- a container - which itself could be comprised of files and packages of files. 
- a IoT/device runtime (e.g., Linux Runtime)
- some composition of two or more of the above (e.g., containers, files and packages) - This later condition enables one to represent a complex system such as a SaaS solution or a system comprised of mutliple containers. 

## Project Directory Structure
### Services
  - object-storage - The minio bucket for storage.
  - db - the postgresql databse container and goose db migration support
  - main - the main container that runs the server holding all API endpoints.
  - nginx - The reverse proxy and file host for the frontend.


## Project License

The source code for this project is provided under the Apache 2.0 license license. Because this project is built on other projects different files may be under different licenses. Each source file should include a license notice that designates the licensing terms for the respective file.

## Docker Compose
The repo provides a compose.yml for a default set-up, as simple as installing the docker compose plugin and running `docker compose up`.

If you want to configure for your local environment without changing the default, the .gitignore will ignore .env and compose.yaml files. The compose plugin will automatically prefer compose.yaml over compose.yml.

## Install & Run
The only prerequisites for first run is [docker](https://docs.docker.com/get-docker/) and the [docker compose plugin](https://docs.docker.com/compose/install/).

The Software Parts Catalog can then be started in the background with `docker compose up -d`.
Use `docker compose down` to stop it, or `docker compose down -v` to also stop the virtual volumes containing the database and uploads directories.

## Querying Data
We use Graphql to provide access to our data.
More details can be found in the [data access document](/docs/data-access.md), or at out [Graphql schema](/services/main/packages/graphql/schema.graphqls).

## Persistent Data
To persist data, you can mount volumes onto your local system.
To do so edit the volumes section at the bottom of the compose.yml.

For example the following would mount the Postgres database volume at /database.
```yml
  database:
    driver_opts:
      type: none
      o: bind
      device: /database
```
Change the value of device to wherever you want to store the data.

## Configuring For Production
### Environment Variables
Many of the values in compose.yml can be set via environment variable, you can set environment variables by creating a .env file where each line is `VARIABLE_NAME=VARIABLE_VALUE`.

|Environment Variable|Default| |
|--------------------|-------|-|
|DB_HOST|db|Host and Port define the database that main will check is available before starting|
|DB_PORT|5432||
|S3_ENDPOINT|object-storage:9000|Explicitly set to empty if just using default AWS S3 endpoint|
|S3_REGION|docker|
|S3_BUCKET|storage|
|S3_USER|testuser|
|S3_SECRET|testsecret||
|MINIO_ROOT_USER|testuser||
|MINIO_ROOT_PASSWORD|testsecret|
|MINIO_DEFAULT_BUCKETS|storage|
### MINIO
The provided Minio instance is for testing purposes only.
For production purposes it should be replaced, either with your own Minio instance, AWS S3, or any other S3-compatible storage.

This S3 storage can be configured by setting the appropriate S3_* environment variables above.
S3_ENDPOINT can be set to "" if using AWS's normal S3 endpoint.
### PostgreSQL
The Catalog uses a Main database for storing software parts and data about those parts, and a Blob database to store metadata (size, checksums, etc) about files and archives stored in S3.

These can be hosted as on the same Postgres instance, but they should be two separate databases, otherwise there will be conflicts with the migration table.

To configure an external database, you should create your own secrets directory with the database connection info for main, db.json, and for the blob storage database, blob.json.
### Configuring HTTPS
First obtain an ssl certificate from [Let's Encrypt](https://letsencrypt.org/getting-started/) or some other Certificate Authority

Once that is done, update your compose file to point ssl_certificate and ssl_certificate_key's file attribute to point to your ssl certificate and your private key.

If your certificate authority also gave you chained certificates, append those to the end of your ssl certificate file.
You can then enable ssl by setting the environment variable USE_SSL to true, yes, or on.

If you already have the catalog running, bring it down and back up again, and the Nginx config should how have ssl enabled.
#### Configuring Secrets
There is an example in default_secrets.
You can create your own secrets, as "secrets", which is ingnored by the gitignore.

The db.json and blob.json are in the form
```json
{
    "dbname": "database name",
    "user": "postgres user",
    "password": "da39a3ee5e6b4b0d3255bfef95601890afd80709",
    "host": "postgres host",
    "port": 5432
}
```

The aes key and password encryption can be done user services/main/packages/cryptography.

The key can be generated with `go run services/main/packages/cryptography/cli.go -k secrets/aes.key -g`.

Passwords can be encrypted with `go run services/main/packages/cryptography/cli.go -k /tmp/aes.key -e ${password}`.

When you encrypt a password it will be returned to you in the form `${password} -> tag:payload + nonce`.

The format this should be in the json files is `${nonce}${tag}${payload}`.

So encrypting "test" would give me `test -> eba07b32:34f1a29e68d14151a5e4ade31225e783 + 46ba3238a56f99be711f2b3e` which would be written as `46ba3238a56f99be711f2b3eeba07b3234f1a29e68d14151a5e4ade31225e783`.

## Legal Notices
All product names, logos, and brands are property of their respective owners. All company,
product and service names used in this software are for identification purposes only.
Wind River is a registered trademarks of Wind River Systems, Inc. 

Disclaimer of Warranty / No Support: Wind River does not provide support
and maintenance services for this software, under Wind River’s standard
Software Support and Maintenance Agreement or otherwise. Unless required
by applicable law, Wind River provides the software (and each contributor
provides its contribution) on an “AS IS” BASIS, WITHOUT WARRANTIES OF ANY
KIND, either express or implied, including, without limitation, any warranties
of TITLE, NONINFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A PARTICULAR
PURPOSE. You are solely responsible for determining the appropriateness of
using or redistributing the software and assume any risks associated with
your exercise of permissions under the license.
