# main
Main is the backend server for the Software Parts Catalog.
It communicates with your chosen database and cloud object storage, to handle any requests on the endpointsit serves.

Program execution starts at main.go, which handles any input, command-line arguments, environment variables, and/or configuration files, and passes on the results of that to the server package.

## vendor
The vendor directory is created by `go mod vendor` and saves required dependencies locally.
This pins dependencies at an exact version, and currently allows our private libraries to be accessible to the docker build.