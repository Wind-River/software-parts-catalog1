## Software Parts Data Model

Maintaining a catalog of software components (parts) is a requirement whether generating SBOMs for managing license compliance, security assurance, export control, or safety certification. We developed a highly searchable scalable software parts database (catalog) enabling companies to seamlessly manage 1000s, 10,000s if not 100,000s of software parts from which their products are comprised. The catalog stores specific intrinsct data for each software part. For example, name, version, content (binary or source code), size, the licensing, legal notices, cve information and so forth. 

A software part is any software component that represents a set of software files from which larger softwate solutions are comprised. Our definition supports a wide range of software types (parts). A part can be:
  - single source file (e.g., main.c)
  - single runtime binary file (e.g., app.exe)
  - a container image file 
  - a collection of parts (files)  - i.e., an archive of two  or more smaller software parts (e.g., busybox.1.32.1.tar.gz, file.rpm). It may include source, binary files, other archives and/or container images. 
  - a container's content - (e.g., a collection of arhives, source files and binaries)

The catalog stores specific intrinsct data for each software part. For example, the license

## Part Types
| Type              | Comprised Of* | Examples | Notes |
|-------------------| ------------ | -------- | ----- |
| file/src      | n/a | main.c, main.js      | Uploaded as an archive of 1 file |
| file/binary/app       | True | app.exe     | Uploaded as an archive of 1 file |
| file/collection/archive | link | busybox.1.31.2.tar.gz |  |
| file/binary/container   | link | |  |
| collection/contents     | logical | |  |
| file/binary/runtime     | link | |  | Uploaded as an archive of 1 file |

*The 'comprised of' column notes whether the type may have a link or logical structure. 

### List of Part Data fields
- UUID
- Name (e.g., busybox -> busybox-1.32.5.tar.gz)
- Version (optional) (e.g., 1.32.5)
- Family Name (e.g., /busybox/1.x)
- Content
  - List of archives
- FVC (if applicable)
- Size (kb)
- License
- License rational
- Automation License
- Legal Notices
- List of CVEs
- List of aliases*
- Cryptography algorithm*
- List of alternative external links/references*



## Part IDs

### Part Alises

### License Expressions





## Notes
  - single source file (main.c)
  - single runtime binary file (setup.exe)

