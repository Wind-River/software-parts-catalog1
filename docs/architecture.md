# Architecture
TKDB consists of a database, and several different services accessing said database.
1. Vue frontend that communicates with [backend](#srcbackend) endpoints.
2. [REST endpoints](#srcroutes) for third-parties to ask for licensing information.
3. A [kafka](#srcdisclosureskafka) listener to pre-load archives. 
## Repo Map
### src
src contains the actual Go TKDB source code and module.
In VSCode this directory should be opened directly, or added as a workspace folder for the Go language server to work correctly.
### docker
docker/ contains the various Docker builds for the TK services.
#### docker/server
docker/server is the build for the public TK endpoints.
#### docker/db
docker/db is the build for the PostgreSQL database.
#### docker/bucket
docker/bucket is the build for the MinIO server for block storage.
### migrate
migrate contains the goose migration files used for updating a database's schema over time.
## Code Map
### main
The main package can be found directly under [src](#src).
The purpose of the main package is to read in configuration information (command-line arguments, config files, etc.) and set-up a server accordingly.
### src/server
The server package sets-up and serves all our endpoints.
This includes serving the frontend, backend queries for the frontend, and REST api calls.
### src/routes
The routes package defines the endpoints for third-parties to query TKDB about specific archives or file collections.
### src/vue
The vue package serves a vue frontend for IP analysts to upload archives, assign data to them, and look up previously entered data.
### src/backend
The backend package defines the endpoints required for the [vue](#srcvue) frontend to work.
This includes endpoints for uploading archives, and endpoints for queries that return list items.
### src/archive
The archive package defines an archive processor, which consists of 1 or more archive processor actors, and a public api that hides this underlying actor structure.
The archive processor actors read work messages from a channel and return an error, or nil, on a return channel that was included in the work message.
### src/disclosures/kafka
The kafka package listen to a kafka queue for disclosures.
The source code associated with the disclosure is then downloaded, and processed using the [archive processor](#srcarchive).