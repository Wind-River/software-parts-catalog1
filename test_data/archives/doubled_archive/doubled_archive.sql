INSERT INTO file (sha256, file_size, md5, sha1) VALUES ('\xadd3ef473d4391ac56cce6e4103b0b5233243b44e705366d363d470466229279', 32, '\xf795f3c8ef3d55b5196a3409d8da967c', '\xb05c030bbbc7d54d1cfcdd70eeab1680e1ca70e5');
INSERT INTO file_alias (file_sha256, name) VALUES ('\xadd3ef473d4391ac56cce6e4103b0b5233243b44e705366d363d470466229279', 'a.txt');
INSERT INTO file (sha256, file_size, md5, sha1) VALUES ('\x9ba6117a5e1abc17c8ddfc0c4eebeb137afb875dfae9e20a9ef5c65061d1adcf', 32, '\xec6a1e97f80e2e69941276ca2e9e910c', '\x21ef5d7b841d5ea7168abf46706c85cbfb4674df');
INSERT INTO file_alias (file_sha256, name) VALUES ('\x9ba6117a5e1abc17c8ddfc0c4eebeb137afb875dfae9e20a9ef5c65061d1adcf', 'b.txt');

INSERT INTO archive (sha256, archive_size, md5, sha1) VALUES ('\xb23f12f78a8c6d1b2ff25ef07e2cdd31b43b8a74b4da270c826428aced759ff4', 149, '\x951a9601839b2015b29b21b697107b99', '\x9ea12e78d980d3af2cd7aa0be0210eb352f5f869');
INSERT INTO archive_alias (archive_sha256, name) VALUES ('\xb23f12f78a8c6d1b2ff25ef07e2cdd31b43b8a74b4da270c826428aced759ff4', 'bar.tar.bz2');
INSERT INTO archive_alias (archive_sha256, name) VALUES ('\xb23f12f78a8c6d1b2ff25ef07e2cdd31b43b8a74b4da270c826428aced759ff4', 'foo.tar.bz2');
INSERT INTO archive (sha256, archive_size, md5, sha1) VALUES ('\x25abf2e10fe6d86daabc50b56d57eb2209de6df6a32821e0dc8c6e95faa67f3a', 482, '\x4a34fe2f151ec0f0940d43ade312e2f1', '\xef29b71062f9160ffa0d50003990ebf2f35b25ab');
INSERT INTO archive_alias (archive_sha256, name) VALUES ('\x25abf2e10fe6d86daabc50b56d57eb2209de6df6a32821e0dc8c6e95faa67f3a', 'doubled_archive.tar.bz2');

INSERT INTO part (part_id, type, name) VALUES ('ad1b4b1a-dbaf-431f-b0ea-7aa6d2a9b40e', 'collection.test', 'bar');
INSERT INTO part (part_id, type, name) VALUES ('29873f87-567c-42f1-81a1-3193db22e023', 'collection.test.final', 'doubled_archive');

INSERT INTO part_has_file (part_id, file_sha256, path) VALUES ('ad1b4b1a-dbaf-431f-b0ea-7aa6d2a9b40e', '\x9ba6117a5e1abc17c8ddfc0c4eebeb137afb875dfae9e20a9ef5c65061d1adcf', 'b.txt');
INSERT INTO part_has_file (part_id, file_sha256, path) VALUES ('29873f87-567c-42f1-81a1-3193db22e023', '\xadd3ef473d4391ac56cce6e4103b0b5233243b44e705366d363d470466229279', 'a.txt');

INSERT INTO part_has_part (parent_id, child_id, path) VALUES ('29873f87-567c-42f1-81a1-3193db22e023', 'ad1b4b1a-dbaf-431f-b0ea-7aa6d2a9b40e', 'bar.tar.bz2');
INSERT INTO part_has_part (parent_id, child_id, path) VALUES ('29873f87-567c-42f1-81a1-3193db22e023', 'ad1b4b1a-dbaf-431f-b0ea-7aa6d2a9b40e', 'foo.tar.bz2');
UPDATE part SET file_verification_code='\x4656433200e837a294e694486cc8bb56706fbd068b7a9aec4539eae340554d70424d6b03de' WHERE part_id='29873f87-567c-42f1-81a1-3193db22e023';

UPDATE archive SET part_id='29873f87-567c-42f1-81a1-3193db22e023' WHERE sha256='\x25abf2e10fe6d86daabc50b56d57eb2209de6df6a32821e0dc8c6e95faa67f3a';
UPDATE archive SET part_id='ad1b4b1a-dbaf-431f-b0ea-7aa6d2a9b40e' WHERE sha256='\xb23f12f78a8c6d1b2ff25ef07e2cdd31b43b8a74b4da270c826428aced759ff4';