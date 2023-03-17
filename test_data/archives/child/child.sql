INSERT INTO file (sha256, file_size, md5, sha1) VALUES ('\xcfad3aee2dee60fed3b16bbcef9f290f298b7414026fa2d81fc7242674225e37', 32, '\xc4505bfb9b8c221cca707c92c4f33de5', '\xedfffc430babbda7c3b77a0ac903d3cb04c95d7e');
INSERT INTO file_alias (file_sha256, name) VALUES ('\xcfad3aee2dee60fed3b16bbcef9f290f298b7414026fa2d81fc7242674225e37', 'a.txt');
INSERT INTO file (sha256, file_size, md5, sha1) VALUES ('\x1e21f53f3afa9bd3c3c79dd6b969294806e28639348e52090c794ba03c978037', 32, '\x083045d8e291900840afb2fae6cdefac', '\x6746d52b21bc0c49c2c5905e14d1ec5778c46f9f');
INSERT INTO file_alias (file_sha256, name) VALUES ('\x1e21f53f3afa9bd3c3c79dd6b969294806e28639348e52090c794ba03c978037', 'b.txt');

INSERT INTO archive (sha256, archive_size, md5, sha1) VALUES ('\xc1f73f426bc941ac9501f18b28d0ab6cd2c7a66b0ca20db08aaf586167e38483', 151, '\xd6486fd8a451aa42a11f6a25d88392df', '\xc043c4c73a3519cfe4662421f07c59d717c04bc2');
INSERT INTO archive_alias (archive_sha256, name) VALUES ('\xc1f73f426bc941ac9501f18b28d0ab6cd2c7a66b0ca20db08aaf586167e38483', 'bar.tar.bz2');
INSERT INTO archive (sha256, archive_size, md5, sha1) VALUES ('\xc34164a0f4375b0be608a2ba04a59dec46a58a762a3a6cac1bd89d946d5dc1af', 422, '\xe734ba188c1c7a106b95641a8bec7ce1', '\xe53e584d02da55bcf0958392ad5f459b660243fe');
INSERT INTO archive_alias (archive_sha256, name) VALUES ('\xc34164a0f4375b0be608a2ba04a59dec46a58a762a3a6cac1bd89d946d5dc1af', 'child.tar.bz2');

INSERT INTO part (part_id, type, name) VALUES ('97379b1d-f2d5-41e1-9159-18df871ffb69', 'collection.test', 'bar');
INSERT INTO part (part_id, type, name, 
    version, family_name, size, license, license_rationale, license_notice, automation_license, automation_license_rationale) VALUES (
    '72c1bafc-82e1-4ea2-ab0a-0d51cdc2d8fe', 'collection.test.final', 'child',
    '0.0.1', 'testFamily', 15, 'test_license', '"this is a test rationale"', 'this is a test notice', 'scancode license would be here', '{"automated_rationale":"test"}');

INSERT INTO part_has_file (part_id, file_sha256, path) VALUES ('97379b1d-f2d5-41e1-9159-18df871ffb69', '\x1e21f53f3afa9bd3c3c79dd6b969294806e28639348e52090c794ba03c978037', 'b.txt');
INSERT INTO part_has_file (part_id, file_sha256, path) VALUES ('72c1bafc-82e1-4ea2-ab0a-0d51cdc2d8fe', '\xcfad3aee2dee60fed3b16bbcef9f290f298b7414026fa2d81fc7242674225e37', 'a.txt');

INSERT INTO part_has_part (parent_id, child_id, path) VALUES ('72c1bafc-82e1-4ea2-ab0a-0d51cdc2d8fe', '97379b1d-f2d5-41e1-9159-18df871ffb69', 'bar.tar.bz2');
UPDATE part SET file_verification_code='\x4656433200466df1af471c6f8230277b47bdfab3f3ca09cac8057d2a6a28b9f93a636dd769' WHERE part_id='72c1bafc-82e1-4ea2-ab0a-0d51cdc2d8fe';

UPDATE archive SET part_id='97379b1d-f2d5-41e1-9159-18df871ffb69' WHERE sha256='\xc1f73f426bc941ac9501f18b28d0ab6cd2c7a66b0ca20db08aaf586167e38483';
UPDATE archive SET part_id='72c1bafc-82e1-4ea2-ab0a-0d51cdc2d8fe' WHERE sha256='\xc34164a0f4375b0be608a2ba04a59dec46a58a762a3a6cac1bd89d946d5dc1af';