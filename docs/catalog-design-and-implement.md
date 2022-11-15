## Overview
TK – Tribal Knowledge DB is a component repository analogous to IMDB (for movies) but for of all third-party components (parts) used by an organization (e.g., Wind River, Honda, Aptiv, …). We refer to software components stored in TK as software parts. A part can be a single file, a collection of files (and directories), a container or an entire shipping product (VxWorks7, Simics or even WR Studio (SaaS)). We define a software part in a latter section. The common services provide by the TK microservice include:
  * Store and retrieve software parts
  * Query for specific intrinsic data about a given part:
    * Size, content, unique identifier (e.g., SHA256, verification code)
    * License
    * security vulnerabilities (future)
    * quality assurance rating and known bugs (future)
    * export compliance classification for different jurisdictions (future)
    * certification data (future)
  * Obtain the content (e.g., source code or file collection) for a part. The content could be just the collection of files or the original archive upload to the system or a link to the source at another location. We might have multiple archive instances for a given part. 
  * Obtain a URL to a single file contained within the part
  * Info accessible via GraphGL API (future)

### High Level Requirements:
  * For each software part
    * UUID - specific to TK
    * File Verification Code (FVC) – may not be available if contain is not available 
    * Name (e.g., busybox)
    * Version (e.g., 1.32.2)
    * Maintain a name-version id (busybox-1.32.2)
    * Family name (e.g., busybox-1.X)
    * Size in bytes
    * Description
    * type (archive, container, composite, ...)
    * Content - Store all file content when available – part records may be created prior to obtaining and storing conent
    * Maintain archive sha256 id (may have multiple values if stored using different archive methods
    * Concluded Top license, determined By who
    * Automation Top License, by what method
    * List of Licenses
    * license notices (text field)
    * Contains OpenSource T/F
    * Contains Proprietary T/F
    * Notes log
    * Date last reviewed
  * TK interface: (tbd)
  * Access/permissions/roles and responsibilities control

#### Content
Content is one or more files and/or collection of files and other parts (items with tkid). Content 

## Software Parts

A part can be: 
  1. a single file 
  1. a collection of files (and directories)
  1. a container,
  1. an entire shipping product (VxWorks7, Simics) or even a more complex systm such as a software as a service solution (e.g., WR Studio). 
  
  Currently TK supports levels 1 and 2. 

### 


## Design

### Data Model
#### Stored procedures

## Implementation

### Coding Standards
We follow the goland standards describe here:
  * https://github.com/golang/go/wiki/CodeReviewComments 
  * https://go.dev/doc/effective_go

Use the following license notice in every file:

    // Copyright (c) 2020 Wind River Systems, Inc.
    //
    // Licensed under the Apache License, Version 2.0 (the "License");
    // you may not use this file except in compliance with the License.
    // You may obtain a copy of the License at:
    //       http://www.apache.org/licenses/LICENSE-2.0
    // Unless required by applicable law or agreed to in writing, software  distributed
    // under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
    // OR CONDITIONS OF ANY KIND, either express or implied.

### Data Model

### Code Organization

### UI
### REST API





