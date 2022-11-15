# TK Set-Up
## Dependencies
- libarchive
### Arguments
#### Maintenance
TODO Runs TK in maintenance mode

CLI
    `-m`
#### Port
TK Server listens on the given port.

CLI
    `-p 4200`
#### Upload Directory
Directory to store uploaded files.

CLI
    `-u ""`
Config
    ```toml
    [server]
    upload = ""
    ```
#### Frontend Directory
Directory to serve to web traffic

CLI
    `-f ""`

Config
    ```toml
    [server]
    frontend = ""
    ```
#### Blob File Storage
Incompatible with [Blob Object Storage](#blob-object-storage).
If set, extracted files will be stored on the local filesystem at the given directory.

CLI
    `--fs ""`
Config
    ```toml
    [blob]
    directory = ""
    ```

#### Blob Object Storage
Incompatible with [Blob File Storage](#blob-file-storage).
If set, extracted files will be stored on the given object storage bucket.
Object storage connection info is found at /secrets/blob.json.

CLI
    `--bucket ""`
Config
    ```toml
    [blob]
    bucket = ""
    ```

#### Version
If set, program prints version information and exits.

CLI
    `--version`

#### Help
If set, program prints usage information and exits.

CLI
    `--help`

#### Threads
Number of goroutines archive processor runs.

CLI
    `--threads 1`
Config
    ```toml
    [server]
    threads = 1
    ```

#### Kafka Host
Kafka Hostname

CLI
    `--kafka ""`
Config
    ```toml
    [kafka]
    host = ""
    ```

#### Frontdoor Host
Frontdoor host to request source from.

CLI
    `--frontdoor ""`
Config
    ```toml
    [frontdoor]
    host = ""
    ```

#### Config
Path to config file.

CLI
    `--config`
#### AES key
/secrets/aes.key

### Database
/secrets/db.json
### File Storage
#### FS
#### Object Storage
/secrets/blob.json
### Kafka
### frontdoor

### Config
#### db.json
#### blob.json
#### aes.key
### config.toml
