# Data Model
## Core
Core structural tables, for recording the entities that will be described in the data tables.
### File
> #### TABLE: file
> Core file details.
> | Column | Type | Default | |DESCRIPTION|
> |--------|------|---------|-|-----------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |checksum_sha1|VARCHAR(40)|-|NOT NULL UNIQUE||
> |checksum_sha256|VARCHAR(64)|-|UNIQUE||
> |checksum_md5|VARCHAR(32)|-|UNIQUE||
> |size|BIGINT|-1|NOT NULL|File size in bytes|
> |insert_date|TIMESTAMP|NOW()|NOT NULL||
> |flag_symlink|INTEGER|0|NOT NULL|Boolean file is a symlink|
> |flag_fifo|INTEGER|0|NOT NULL|Boolean file is a named pipe|

> #### TABLE: file_alias
> Tie a file name to a file.
> 
> This is separate from the file table since the same file can be seen with many different file names.
> | Column | Type | Default | |Description|
> |--------|------|---------|-|-----------|
> |id|BIGSERIAL|AUTO|UNIQUE||
> |file_id|BIGINT|-|REFERENCES file(id)||
> |name|TEXT|-|NOT NULL|File name|
> * PRIMARY KEY (file_id, name)

> #### FUNCTION: insert_file(TEXT, BIGINT, VARCHAR(40), VARCHAR(32), INTEGER, INTEGER) -> BIGINT
> Convenience function to both insert into the file table(or update its sha1 and md5), then insert into the file_alias table.
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_name|TEXT|File Name|
> |_size|BIGINT|File Size (bytes)|
> |_sha1|VARCHAR(40)|Sha1 of the file|
> |_md5|VARCHAR(32)|Md5 of the file|
> |_symlink|INTEGER|Boolean is file a symlink|
> |_fifo|INTEGER|Boolean is file a named pipe|
> ##### OUTPUT 
> _fid BIGINT:
> id of the resulting file_alias
> ##### METHOD
> 1. Insert file, or update sha1 and md5
> 2. Insert file_alias
> 3. Return file_alias id
### Archive
The archive tables describe an archived package, and its contents during processing, before all of those are assigned to a file_collection entry.
The only table that should still have an entry after processing, should be the archive table.
> #### TABLE: archive
> A concrete file_collection. Different from a file_collection because the same file_collection can be compressed two different ways, and produce two distinct files.
> |Column|Type|Default| |Description|
> |--------|----|-----|-|-----------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |file_collection_id|BIGINT|-|REFERENCES file_collection(id)||
> |name|TEXT|-||File Name|
> |path|TEXT|-||Where the archive is stored|
> |size|INTEGER|-||File size in bytes|
> |checksum_sha1|VARCHAR(40)|-||Sha1 of the file|
> |checksum_sha256|VARCHAR(64)|-||Sha256 of the file|
> |checksum_md5|VARCHAR(32)|-||Md5 of the file|
> |insert_date|TIMESTAMP|NOW()|NOT NULL||
> |extract_status|INT|0|NOT NULL|0 unextracted; 1 extracted; -x error;|
> * UNIQUE(name, checksum_sha1)

> #### TABLE: archive_contains
> Table tracking archives contained within other archives.
> |Column|Type|Default| |Description|
> |------|----|-------|-|-----------|
> |parent_id|BIGINT|-|NOT NULL REFERENCES archive(id)|Parent|
> |child_id|BIGINT|-|NOT NULL REFERENCES file_collection(id)|Child|
> |path|TEXT|-|NOT NULL|Path where child archive can be found within parent archive|
> *  PRIMARY KEY(archive_id, file_id, path)

> #### TABLE: file_belongs_archive
> Files found within an archive. Other archives would instead be in archive_contains
> |Column|Type|Default| |Description|
> |------|----|-------|-|-----------|
> |archive_id|BIGINT|-|NOT NULL REFERENCES archive(id)|Parent archive|
> |file_id|BIGINT|-|NOT NULL REFERENCES file_alias(id)|File alias of the contained file|
> |path|TEXT|-|NOT NULL|Path where file can be found within parent archive|
> * PRIMARY KEY(archive_id, file_id, path)

> #### FUNCTION: calculate_archive_verification_code(BIGINT) -> VARCHAR(40)
> Calculates verification code of the archive. The archive needs to have all its files in file_belongs_archive table with sha1
> #### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_aid|BIGINT|Archive ID|
> #### OUTPUT
> VARCHAR(40)
> Verification Code (is a sha1)
> #### METHOD
> 1. Select sha1 of every file in archive contained by that archive (ignore symlinks and named pipes)
> 2. Sort list of sha1s
> 3. Return sha1 of that list

> #### FUNCTION: assign_archive_to_file_collection(BIGINT, BIGINT)
> Fill in details of a file_collection using an archive as a source
> #### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_aid|BIGINT|Archive ID|
> |_cid|BIGINT|File Collection ID|
> #### OUTPUT
> Files and subarchives reflected in file_collection entry
> #### METHOD
> 1. Iterate every file in file_belongs_archive and enter int file_belongs_collection
> 2. Delete file_belongns_archive entries
> 3. Iterate archive_contains entries and enter into file_collection_contains
> 4. Delete archive_contains entries
### File Collection
Describes a collection of files, and any sub-collections it may contain
> #### TABLE: file_collection
> Describes an abstract collection of files
> | Column | Type | Default| |Description|
> |--------|------|--------|-|-----------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |verification_code|VARCHAR(40)|-|UNIQUE|Make a list of every file's sha1, sort the list, then take the sha1 of that list|
> |insert_date|TIMESTAMP|NOW()|NOT NULL||
> |group_container_id|BIGINT|-|REFERENCES group_container(id)||
> |license_id|BIGINT|-|REFERENCES license_expression(id)||
> |license_rationale|TEXT|-||Rationale used to determine collections license|

> #### TABLE: file_collection_contains
> For tracking archives contained within other archives
> (There should not be a file entry for such an archive)
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |parent_id|BIGINT|-|NOT NULL REFERENCES file_collection(id)|Archive that contains another archive|
> |child_id|BIGINT|-|NOT NULL REFERENCES file_collection(id)|Archive contained by another archive|
> |path|TEXT|-|NOT NULL|Path where child archive can be found within parent archive|
> * Constraint: CHECK(parent_id<>child_id)
> * PRIMARY KEY(parent_id, child_id, path)

> #### TABLE: file_belongs_collection
> Associates a collection with its file
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |file_collection_id|BIGINT|-|NOT NULL REFERENCES file_collection(id)||
> |file_id|BIGINT|NOT NULL REFERENCES file(id)||
> |path|TEXT|NOT NULL|Path where file can be found within collection|
> * PRIMARY KEY(file_collection_id, file_id, path)

> #### FUNCTION: stub_file_collection(VARCHAR(40)) -> BIGINT
> Insert only verification_code into file_collection if it doesn't already exist
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |vcode|VARCHAR(40)|Verification Code|
> ##### OUTPUT 
> _ret BIGINT:
> id of the resulting, or already existing, file_collection
> ##### METHOD
> 1. Select id from file_Collection matching vcode
> 2. If exists, return
> 3. If not, insert and return

> #### FUNCTION: select_file_collections(BIGINT) -> SETOF BIGINT
> Recursively returns ids of file_collections that are contained within _cid
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_cid|BIGINT|ID of parent File Collection|
> #### OUTPUT
> {BIGINT, ...}
> id of file_collection contained within _cid
> #### METHOD
> 1. Select child_id from file_collection_contains
> 2. Return id, and recursively call and return ids from function

> #### FUNCTION: select_file_collection_files(BIGINT) -> SETOF BIGINT
> Returns file ids of files in collection
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_cid|BIGINT|ID of parent File Collection|
> #### OUTPUT
> {BIGINT, ...}
> id of file contained within _cid

> #### FUNCTION: calculate_collection_verification_code(BIGINT) -> VARCHAR(40)
> Calculates verification code of file collection. File Collection needs to have all its files in file_belongs_collection table with sha1
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_cid|BIGINT|ID of parent File Collection|
> #### OUTPUT
> VARCHAR(40)
> Verification Code (is a sha1)
> #### METHOD
> 1. Select sha1 of every file in file collection and collections contained by that file collection (ignore symlinks and named pipes)
> 2. Sort list of sha1s
> 3. Return sha1 of that list
### Component
Mechanism for relating file_collections, and other components, as a single entity
> #### TABLE: composite_component
> Base component table
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |name|TEXT|-||
> |description|TEXT|-||

> #### TABLE: component_contains_collection
> Many-to-Many relationship between composite_components and file_collections
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |component_id|BIGINT|-|NOT NULL REFERENCES composite_component(id)||
> |file_collection_id|BIGINT|-|NOT NULL REFERENCES file_collection(id)||
> * PRIMARY KEY (component_id, file_collection_id)

> #### TABLE: component_contains_component
> Many-to-Many relationship between composite_components and other composite_components
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |parent_component|BIGINT|NOT NULL REFERENCES composite_component(id)||
> |child_component|BIGINT|NOT NULL REFERENCES composite_component(id)||
> * PRIMARY KEY (parent_component, child_component)
> * CHECK (parent_component<>child_component) | A component cannot contain itself
## Data
### Groups
Groups are typically used to associate packages that should be largely similar. 
For example, several versions of the same library.
> #### TABLE: group_container
> Table of groups/families.
> Groups are typically the package name and major version
> For example, busybox-1.29.1.tar.bz2 would be in group 1.X under group busybox.
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |parent_id|BIGINT|-|REFERENCES group_container(id)|File Collection Name|
> |name|TEXT|-|NOT NULL||
> |type|VARCHAR(20)|-|||
> |associatedlicense|TEXT|-||License children packages should inherit if they don't have their own|
> |associatedratinoale|TEXT|-||Reasoning behind associatedlicense|
> |insert_date|TIMESTAMP|NOW()|NOT NULL||
> |description|TEXT|-|||
> |comments|TEXT|-|||
> * UNIQUE(name, parent_id)

> #### FUNCTION: parse_group_path(TEXT) -> BIGINT
> Takes a group path (i.e. /busybox/1.X), which represents a series of group parent/child relationships, and and returns the id of the last child
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |path|TEXT|Group path to parse|
> #### OUTPUT
> BIGINT
> id of the last group_container
> #### METHOD
> 1. Split path on '/'
> 2. Skip null(if path starts with '/')
> 3. Iterate through path parts, and find groups matching parent_id and name
> 4. Return last id

> #### FUNCTION: build_group_path(BIGINT) -> TEXT
> Takes a group id, and follows the chain of parents to return a group path
> ##### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_id|BIGINT|Group id|
> #### OUTPUT
> TEXT
> Group path of the given group(i.e. /busybox/1.X)
### License
> #### TABLE: license
> Individual License Information
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |name|TEXT|''||Not Used|
> |insert_date|TIMESTAMP|NOW()|NOT NULL||
> |type|VARCHAR(32)|'custom'|NOT NULL|Machine entered licenses should by typed as 'auto-fill'|
> |identifier|TEXT|-|NOT NULL UNIQUE|License identifier (i.e. GPL-2.0)|
> |text|TEXT|-||Whole license text|
> * UNIQUE(name, identifier)

> #### TRIGGER: license_cascade_update()
> Triggers update on license_expressions when underlying license entry changes
> ##### CONDITION
> AFTER UPDATE ON license
> ##### METHOD
> Set license_expressions's expression to null where license_id=id
> which triggers license_expression_update()

> #### TABLE: license_expression
> Table representing binary trees for license expressions.
> file_collections will reference license_expression s which will reference license s
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |type|TEXT|'LICENSE'|NOT NULL|'LICENSE' for license, 'BOP' for Binary Operator|
> |license_id|BIGINT|-|REFERENCES license(id)|NULL if not a 'LICENSE'|
> |expression|TEXT|-||Automatically updated license expression combining self and children|
> |operator|TEXT|-||NULL if not a 'BOP'|
> |left_id|BIGINT|-|REFERENCES License_expression(id)|NULL if no left child or if not a 'BOP'|
> |right_id|BIGINT|-|REFERENCESE license_expression(id)|NULL if no right child or if not 'BOP'|

> #### TRIGGER: license_expression_update()
> Updates expression when a child license_expression node changes
> ##### CONDITION
> AFTER INSERT OR UPDATE ON license_expression
> ##### METHOD
> 1. Set new expression from build_license_expression()
> 2. Set any parent license_expression's expression to null to trigger update

> #### FUNCTION: build_license_expression(BIGINT) -> TEXT
> Recursively traverse a license_expression tree and construct an expression
> #### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_id|BIGINT|License Expression ID|
> #### OUTPUT
> TEXT
> Full license expression

> #### FUNCTION: get_license(TEXT) -> BIGINT
> INSERT and/or SELECT license id for the given license identifier
> #### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_identifier|TEXT|License Identifier|
> #### OUTPUT
> BIGINT
> id of the related license entry

> #### FUNCTION: parse_license_expression(TEXT) -> BIGINT
> Parse a license expression into a license_expression tree
> #### ARGUMENTS
> |Argument|Type|Description|
> |--------|----|-----------|
> |_expression|TEXT|License Expression|
> #### OUTPUT
> BIGINT
> id of the root of the newly created license_expression tree

> #### FUNCTION: cleane_table_license_expression()
> Clean orphaned license expressions
> #### OUTPUT
> Rows that are not referenced by a file_collection or other license_expression are deleted from the license_expression table
### Crypto
This section is for the tables stroing machine collected evidence of cryptography, and recording human determinations on how accurate that stored data is.
> #### TABLE: file_have_crypto_record
> Stores a cryptography "hit" in a file. The hit is possibly evidence of cryptography in the associated file.
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |method|TEXT|-|NOT NULL|Method used to generate this hit|
> |file_id|BIGINT|-|NOT NULL REFERENCES file(id)||
> |match_type|TEXT|-|NOT NULL|cryptography type|
> |match_text|TEXT|-|NOT NULL|text that should be the evidence of cryptography|
> |match_line_number|INTEGER|-|NOT NULL|line number match_text was found|
> |match_file_index_begin|INTEGER|-|NOT NULL|Index within the file where match starts|
> |match_file_index_end|INTEGER|-|NOT NULL|Index within the file where match ends|
> |match_line_index_begin|INTEGER|NOT NULL|Index within the line where the match starts|
> |match_line_index_end|INTEGER|-|NOT NULL|Index within the line where the match ends|
> |line_text|TEXT|-|NOT NULL|The text of the whole line where the match was found|
> |line_text_before_1|TEXT|-||The surrounding lines of code before the matching line|
> |line_text_before_2|TEXT|-||The surrounding lines of code before the matching line|
> |line_text_before_3|TEXT|-||The surrounding lines of code before the matching line|
> |line_text_after_1|TEXT|-||The surrounding lines of code after the matching line|
> |line_text_after_2|TEXT|-||The surrounding lines of code after the matching line|
> |line_text_after_3|TEXT|-||The surrounding lines of code after the matching line|
> |human_reviewed|TEXT|-||Not Used|
> |comments|TEXT|-||
> |crypto_record_log|TEXT|-||
> |crypto_record_hash|VARCHAR(40)|UNIQUE|Sha1 for unique identification|
> * INDEX on file_id

> #### TRIGGER: compute_crypto_record_hash()
> Computes the crypto_record_hash of a crypto hit when the hit is inserted
> #### CODINITON
> BEFORE INSERT ON file_have_crypto_record
> #### METHOD
> 1. Create a dictonary with 'checksum_sha1', 'detection_method', 'evidence_type', 'file_index_begin', 'file_endex_end', 'line_index_begin', 'line_index_end', and 'line_number' from the record
> 2. Convert to json 
> 3. Take the sha1 of the json string

> #### TABLE: file_have_crypto_evidence
> Records cryptography determinations made on a file
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |file_id|BIGINT|-|NOT NULL REFERENCES file(id)||
> |type|TEXT|-|NOT NULL|Algorithm Type|
> |useraccount_id|BIGINT|0|NOT NULL|User|
> |review|REVIEW_TYPE|-|NOT NULL ENUM('false-positive', 'TBD', 'positive')|Lastest crypto review. Past reviews can be found in history.|
> |comment|TEXT|-||Latest comment. Past comments can be found in history.|
> |prediction|CRYPTO_PREDICTION|-|{prediction: BOOL, probability: DOUBLE PRECISION}|Machine Learning prediction and confidence|
> * PRIMARY KEY(file_id, type, useraccount_id)

> #### TRIGGER: file_crypto_evidence_modified()
> Records changes to file_have_crypto_evidence in history table
> ##### CONDITION
> AFTER INSERT OR UPDATE ON file_have_crypto_evidence
> ##### METHOD
> 1. Create a JSONB value
> 2. If comment is changed, store 'comment' in JSONB
> 3. If review is changed, store 'review' in JSONB
> 4. Insert into history setting value to JSONB value
## Misc
### History
Changes to tables can be tracked by creating a trigger on a table and logging to the history table.

> #### TYPE: op_type
> op_type enumerates the possible SQL operations
> | Operation Type |
> |-----------------|
> |INSERT|
> |UPDATE|
> |DELETE|

> #### TABLE: history
> Tracks changes in the database. Things logged include the affected table, the user who made the change, the SQL operation taken, relevant keys and value, and the time the change was made.
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |target|TEXT|-|NOT NULL|The table changed|
> |useraccount_id|BIGINT|-|NOT NULL|The user who made the change|
> |operation|OP_TYPE|-||The SQL operation taken|
> |unique_key|JSONB|-||Primary key columns mapped to values for the target table|
> |value|JSONB|-||Columns mapped to new values in the target table|
> |time|TIMESTAMP|NOW()|NOT NULL||
> * PRIMARY KEY(target, useraccount_id, time)
### Roles
> #### TABLE: analyst
> Identify the analysts to track changes they make
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |name|VARCHAR(100)|-|NOT NULL|File Collection Name|
> |insert_date|TIMESTAMP|NOW()|NOT NULL||
> |job_title|TEXT|-|||
> |note|TEXT|-|||
### Authentication
> #### TABLE api_key
> Stores api keys used by external applications
> | Column | Type | Default | | Description |
> |--------|------|---------|-|-------------|
> |id|BIGSERIAL|AUTO|PRIMARY KEY||
> |uuid|UUID|GEN_RANDOM_UUID()|NOT NULL|Requires pgcrypto extension|
> |key bytea|GEN_RANDOM_BYTES(64)|NOT NULL|Requires pgcrypto extension|
> |status|KEYS_STATUS|'inactive'|ENUM('active', 'inactive')||