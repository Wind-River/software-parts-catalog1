# Graphql
The goal of the Software Parts Catalog is to store all the data we can on a Part.
To then deliver this data to the user we have defined a Graphql API that allows the flexibility to requset just the data they need, without us having to guess every endpoint that might need to be created to satisfy such requests.
If you are used to reading Graphql schemas, you can find our schema at services/main/packages/graphql/schema.graphqls.

The data we return is defined as Graphql Types, which can be queried or modified with Queries and Mutations, the former for read-only data access, and the later for modifying data.
## Types
### Archive
Archive represents an archive that was uploaded to represent a software part.
It is extracted and cataloged such that sub-archives are sub-parts and files are attached to the associated parts.

The data that is contained by an Archive is the archive's identifying information, and the part that was created, or was already in the database and had the same file verification code.

|Field|Type|
|-----|----|
|sha256|hex-encoded string|
|Size|integer count of bytes|
|part_id|UUID referencing a [Part](#part)|
|part|[Part](#part)|
|md5|hex-encoded md5|
|sha1|hex-encoded sha1|
|name|archive filename|
|insert_date|timestamp of archive creation|
### Part
Part represents a software part. This is the core piece that will be associated with any data profiles we have on a given software part.
Every part is identified by a UUID that is generated on creation, or a file verification code that can be calculated from the files it contains.
Additionally other [aliases can be created](#createalias) to also identify a part.
|Field|Type|
|-----|----|
|id|UUID|
|type|string in hierarchichal path form|
|name|string|
|version|string|
|label|string|
|family_name|string|
|file_verification_code|hex-encoded string|
|size|integer count of bytes|
|license|string|
|license_rationale|string|
|description|string|
|comprised|UUID referencing another Part|
|aliases|list of strings|
|profiles|list of [Profiles](#profile) associated with the part|
|sub_parts|list of Parts and their path within this part|
### PartList
|Field|Type|
|-----|----|
|id|integer|
|name|string|
|parent_id|integer referencing a parent PartList|
### Document
Documents are arbitrary data that you can store about a Part.
If your document has an obvious title that may be queried on, you can define a title, which will give the document its own row in the database.
For a single large document however, you can leave title null.
|Field|Type|
|-----|----|
|title|string|
|document|ajson|
### Profile
Profile is a category of data we store about a Part.
The `key` is like the profile name that tells us what we can expect from the associated documents.
For example, in a `security` profile, we'd expect to find CVE documents.
|Field|Type|
|-----|----|
|key|string|
|documents|list of [Documents](#document)|
## Queries
### archive
> archive(sha256: hex-encoded String, name: String): [Archive](#archive)
The archive query looks up and returns an Archive by exact sha256 or name matches.
### find_archive
> find_archive(query: String!, method: String, costs: [SearchCosts](#searchcosts)): [[ArchiveDistance](#archivedistance)!]!

find_archive takes a search term, and optionally a method and search costs to fine tune the search, and returns a list of ArchiveDistance tuples, which is an integer distance from the search term, and the associated archive.

By default the search method is `levenshtein`, but can also be `levenshtein_less_equal` for a more efficient search that cuts off results that are too different, or `fast` that first applies a substring match before narrowing results further with levensthein distances.

Insert, delete, and substitute are operations taken by the search to edit the archive name to match the query.
The defaults below define a low delete cost to tune for the case where we don't expect the user to type the whole name, so many deletions are required to match the archive name to the search query.

Max distance only applies to levenshtein_less_equal, which will abort processing early, and not return that result, if a distance exceeds it.
#### SearchCosts
|Field|Default|Description|
|-----|-------|-----------|
|insert|20|cost if characters need to be added to match the search term|
|delete|2|cost if characters need to be deleted to match the search term|
|substitute|30|cost if characters need to be substituted to match the search term|
|max_distance|75|for levensthein_less_equal costs above max_distance are not returned|
#### ArchiveDistance
|Field|Type|
|-----|----|
|distance|integer|
|archive|[Archive](#archive)|
### part
part tries every non-null argument given to it, and returns the first Part match it finds.
In order:
    1. id: UUID
    2. file_verification_code: Hex-encoded file verification code
    3. sha256: Hex-encoded sha256 of an archive with a non-null part
    4. sha1: Hex-encoded sha1 of an archive with a non-null part
    5. name: file name of an archive with a non-null part
### archives
archives lists [archives](#archive) that match the given part, by part id or file verification code.
### partlist
partlist returns a [PartList](#partlist) matched by id or name.
### partlist_parts
partlist_parts lists the [Parts](#part) contained by a partlist
### partlists
partlists returns the list of other [partlists](#partlist) the given partlist contains.
If parent_id is 0, it returns every root partlist.
### file_count
file_count returs of number of files owned by the given part and its sub-parts.
### comprised
comprised returns the list of parts that are comprised by the given part.
See [Part.comprised](#part) if you are looking for what comprised a given part.
### profile
profile returns a list of [documents](#document) attached to a part.

## Mutations
### addPartList
addPartList creates a new part list with the given parent, or a root part if no parent given
### deletePartList
deletPartList deletes the given empty part list
### deletePartFromList
deletePartFromList removes the given part from the given list
### uploadArchive
Upload an archive to be processed into a part
### updateArchive
Updates the part associated with the given archive
An error will be returned if the associated part hasn't been created yet
### updatePart
updatePartLists adds a list of parts to the given part
### createAlias
Create a part alias
### attachDocument
Attach a document to a part.
To attach a single large document as a profile, do not provide a title.
To attach a smaller, more queryable document, provide a title. (e.g. a CVE id)
### partHasPart
Adds a sub-part to a part at a path
### partHasFile
Adds a file to a part, potentially at a path
### createPart
Create a new part with the given input
