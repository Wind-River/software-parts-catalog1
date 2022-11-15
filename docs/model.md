# Data Model

## Leaf Part
*formerly known as file*

|Column|Notes|
|------|-----|
|Sha256|Primary Identifier|
|Sha1|Secondary Identifier|
|Size|Size of the part in bytes|
|Mime Type|Mime type stored to tell what kind of data it is since file name(s) if any are stored separately|

Additionally a list of file aliases (file names) should be stored

### TKDB Database
*A PostgreSQL database used by TKDB*

__file table__
|Column|Type|Nullable|Represents|
|------|----|--------|----------|
|id|BIGINT|NOT NULL|-|
|checksum_sha1|VARCHAR(40)|NOT NULL|Sha1|
|checksum_sha256|VARCHAR(64)|NOT NULL|Sha256|
|checksum_md5|VARCHAR(32)|NOT NULL|-|
|insert_date|TIMESTAMP|NOT NULL|-|
|flag_symlink|INTEGER|NOT NULL|-|
|flag_fifo|INTEGER|NOT NULL|-|
|size|BIGINT|NOT NULL|Size|

__file_alias table__

A table establishing a many-to-many relationship between the file table (representing a leaf part) and a list of aliases (file names)

|Column|Type|Nullable|Represents|
|------|----|--------|----------|
|id|BIGINT|NOT NULL|-|
|file_id|BIGINT|-|
|name|VARCHAR(400)|NOT NULL|part alias|

### Blob Storage Database
*A SQL Database associated with TKDB's blob storage*

*Currently is an sqlite3 file stored alongside the blob data*

__blob_metadata table__
|Column|Type|Nullable|Represents|
|------|----|--------|----------|
|sha256|BLOB|NOT NULL|Sha256|
|sha1|BLOB|NOT NULL|Sha1|
|size|BIGINT|NOT NULL|Size|
|mime|TEXT|-|Mime Type|