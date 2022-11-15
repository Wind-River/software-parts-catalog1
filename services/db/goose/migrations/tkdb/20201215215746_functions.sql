-- +goose Up
--
--

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION assign_archive_to_file_collection(_aid BIGINT, _cid BIGINT) RETURNS VOID LANGUAGE plpgsql AS $$
    DECLARE
        _row RECORD;
    BEGIN
        FOR _row IN SELECT f.id, fba.path FROM file f
            INNER JOIN file_alias fa ON fa.file_id=f.id
            INNER JOIN file_belongs_archive fba ON fba.file_id=fa.id
            WHERE fba.archive_id=_aid
        LOOP
            INSERT INTO file_belongs_collection (file_collection_id, file_id, path)
            VALUES (_cid, _row.id, _row.path) ON CONFLICT (file_collection_id, file_id, path) DO NOTHING;
        END LOOP;

        DELETE FROM file_belongs_archive WHERE archive_id=_aid;

        FOR _row IN SELECT child_id, path FROM archive_contains WHERE parent_id=_aid
        LOOP
            IF _cid<>_row.child_id THEN
                INSERT INTO file_collection_contains(parent_id, child_id, path)
                VALUES(_cid, _row.child_id, _row.path) ON CONFLICT (parent_id, child_id, path) DO NOTHING;
            END IF;
        END LOOP;

        DELETE FROM archive_contains WHERE parent_id=_aid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION build_group_path(_id BIGINT) RETURNS TEXT LANGUAGE plpgsql AS $$
    DECLARE
        _row RECORD;

        _path TEXT;
        _parent BIGINT;
    BEGIN
        SELECT name, parent_id FROM group_container WHERE id=_id INTO _row;
        _path := '/' || _row.name;
        _parent := _row.parent_id;

        WHILE _parent IS NOT NULL
        LOOP
            SELECT name, parent_id FROM group_container WHERE id=_parent INTO _row;
            _path := '/' || _row.name || _path;
            _parent := _row.parent_id;
        END LOOP;

        RETURN _path;
    END;
$$;
-- +goose StatementEnd

CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_archive_verification_code(_aid BIGINT) RETURNS VARCHAR(40) LANGUAGE plpgsql AS $$
    DECLARE
        _row RECORD;
        _shas VARCHAR(40)[] := ARRAY[]::VARCHAR(40)[];
        _data TEXT := '';
        _vcode TEXT;
    BEGIN
        FOR _row IN SELECT f.checksum_sha1 FROM file f
            INNER JOIN file_alias fa ON f.id=fa.file_id
            INNER JOIN file_belongs_archive fba ON fba.file_id=fa.id
            INNER JOIN archive a ON a.id=fba.archive_id
            WHERE a.id=_aid AND f.flag_symlink=0 AND f.flag_fifo=0
        LOOP
            RAISE NOTICE 'sha: %', _row.checksum_sha1;
            _shas := _shas || _row.checksum_sha1;
        END LOOP;

        FOR _row IN WITH RECURSIVE collections AS (
                SELECT ac.child_id as file_collection_id FROM archive_contains ac WHERE ac.parent_id=_aid
                UNION
                SELECT child_id as file_collection_id FROM file_collection_contains
                INNER JOIN collections ON collections.file_collection_id=file_collection_contains.parent_id
            )
            SELECT f.checksum_sha1 FROM file f
            INNER JOIN file_belongs_collection fc ON fc.file_id=f.id
            INNER JOIN file_collection c ON c.id=fc.file_collection_id
            INNER JOIN collections ON collections.file_collection_id=c.id
            WHERE f.flag_symlink=0 AND f.flag_fifo=0
        LOOP
            _shas := _shas || _row.checksum_sha1;
        END LOOP;

        FOR _row IN SELECT UNNEST(_shas) AS s ORDER BY s
        LOOP
            _data := _data || _row.s;
        END LOOP;

        SELECT encode(digest(_data, 'sha1'), 'hex') INTO _vcode;

        RETURN _vcode;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION calculate_collection_verification_code(_cid BIGINT) RETURNS VARCHAR(40) LANGUAGE plpgsql AS $$
    DECLARE
        _row RECORD;
        _shas VARCHAR(40)[] := ARRAY[]::VARCHAR(40)[];
        _data TEXT := '';
        _vcode TEXT;
    BEGIN
        FOR _row IN SELECT f.checksum_sha1 FROM file f
            INNER JOIN file_belongs_collection fbc ON fbc.file_id=f.id
            INNER JOIN file_collection c ON c.id=fbc.file_collection_id
            WHERE f.flag_symlink=0 AND f.flag_fifo=0 AND c.id=_cid
        LOOP
            RAISE NOTICE 'sha: %', _row.checksum_sha1;
            _shas := _shas || _row.checksum_sha1;
        END LOOP;

        FOR _row IN WITH RECURSIVE collections AS (
                SELECT cc.child_id as file_collection_id FROM file_collection_contains cc WHERE cc.parent_id=_cid
                UNION
                SELECT child_id as file_collection_id FROM file_collection_contains
                INNER JOIN collections ON collections.file_collection_id=file_collection_contains.parent_id
            )
            SELECT f.checksum_sha1 FROM file f
            INNER JOIN file_belongs_collection fc ON fc.file_id=f.id
            INNER JOIN file_collection c ON c.id=fc.file_collection_id
            INNER JOIN collections ON collections.file_collection_id=c.id
            WHERE f.flag_symlink=0 AND f.flag_fifo=0
        LOOP
            _shas := _shas || _row.checksum_sha1;
        END LOOP;

        FOR _row IN SELECT UNNEST(_shas) AS s ORDER BY s
        LOOP
            _data := _data || _row.s;
        END LOOP;

        SELECT encode(digest(_data, 'sha1'), 'hex') INTO _vcode;

        RETURN _vcode;
    END;
$$;
-- +goose StatementEnd

CREATE EXTENSION IF NOT EXISTS plperl;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION create_file(name TEXT, checksum_sha1 VARCHAR(40), flag_symlink INTEGER) RETURNS BIGINT LANGUAGE plperl AS $$
    use strict;
    my ($name, $sha1, $isSymlink) = @_;

    my $filePlan = spi_prepare('INSERT INTO file (checksum_sha1, flag_symlink) VALUES ($1, $2) ON CONFLICT DO UPDATE SET NEW.flag_symlink=$2 RETURNING id', ['varchar(40)', 'integer']);
    my $aliasPlan = spi_prepare('INSERT INTO file_alias (file_id, name) VALUES ($1, $2)', ['bigint', 'text']);

    my $row = spi_exec_prepared($filePlan, {limit=>1}, $sha1, $isSymlink)->{rows}[0];
    my $fid = $row->{id};

    spi_exec_prepared($aliasPlan, $fid, $name);

    return $fid;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION delete_archive(_id BIGINT) RETURNS VOID LANGUAGE plpgsql AS $$
    BEGIN
        DELETE FROM archive_contains WHERE parent_id=_id OR child_id=_id;
        DELETE FROM file_belongs_archive WHERE archive_id=_id;
        DELETE FROM archive WHERE id=_id;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION delete_collection(_id BIGINT) RETURNS VOID LANGUAGE plpgsql AS $$
    BEGIN
        --SELECT delete_archive(id) FROM archive WHERE file_collection_id=_id;
        UPDATE archive SET file_collection_id=NULL WHERE file_collection_id=_id;
        DELETE FROM archive_contains WHERE child_id=_id;
        DELETE FROM component_contains_collection WHERE file_collection_id=_id;
        DELETE FROM file_belongs_collection WHERE file_collection_id=_id;
        DELETE FROM file_collection_contains WHERE parent_id=_id OR child_id=_id;
        DELETE FROM file_collection WHERE id=_id;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION get_license(_identifier TEXT) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        ret BIGINT;
    BEGIN
        IF _identifier IS NULL THEN
            _identifier := '';
        END IF;
        SELECT id FROM license WHERE identifier=_identifier INTO ret;
        IF NOT FOUND THEN
            INSERT INTO license (type, identifier) VALUES ('auto-fill', _identifier) RETURNING id INTO ret;
        END IF;

        RETURN ret;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_archive(_name TEXT, _sha1 VARCHAR(40)) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _cid BIGINT;
    BEGIN
        RAISE NOTICE 'insert_archive: %', _name || ' '|| _sha1;
        EXECUTE 'SELECT c.id FROM archive a INNER JOIN file_collection c ON a.file_collection_id=c.id WHERE a.checksum_sha1=$1'
            INTO _cid
            USING _sha1;
        IF _cid IS NOT NULL THEN
            RETURN _cid;
        ELSE
            INSERT INTO file_collection (insert_date) VALUES(NOW()) RETURNING id INTO _cid;
            INSERT INTO archive (name, checksum_sha1, file_collection_id) VALUES (_name, _sha1, _cid);
            RETURN _cid;
        END IF;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_file(_name TEXT, _sha256 VARCHAR(64), _sha1 VARCHAR(40), _md5 VARCHAR(32), _symlink INTEGER) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _fid BIGINT;
    BEGIN
        INSERT INTO file (checksum_sha256, checksum_sha1, checksum_md5, flag_symlink)
        VALUES (_sha256, _sha1, _md5, _symlink)
        ON CONFLICT (checksum_sha1) DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, checksum_sha256=EXCLUDED.checksum_sha256
        RETURNING id
        INTO _fid;

        IF _fid IS NOT DISTINCT FROM NULL THEN
            RAISE EXCEPTION 'Insert into raw file table did not return an ID';
            RETURN NULL;
        END IF;

        INSERT INTO file_alias (file_id, name) VALUES (_fid, _name)
        ON CONFLICT (file_id, name) DO UPDATE SET name=EXCLUDED.name
        RETURNING id
        INTO _fid;

        RETURN _fid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_file(_name TEXT, _sha256 VARCHAR(64), _sha1 VARCHAR(40), _md5 VARCHAR(32), _symlink INTEGER, _fifo INTEGER) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _fid BIGINT;
    BEGIN
        INSERT INTO file (checksum_sha256, checksum_sha1, checksum_md5, flag_symlink, flag_fifo)
        VALUES (_sha256, _sha1, _md5, _symlink, _fifo)
        ON CONFLICT (checksum_sha1) DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, checksum_sha256=EXCLUDED.checksum_sha256
        RETURNING id
        INTO _fid;

        IF _fid IS NOT DISTINCT FROM NULL THEN
            RAISE EXCEPTION 'Insert into raw file table did not return an ID';
            RETURN NULL;
        END IF;

        INSERT INTO file_alias (file_id, name) VALUES (_fid, _name)
        ON CONFLICT (file_id, name) DO UPDATE SET name=EXCLUDED.name
        RETURNING id
        INTO _fid;

        RETURN _fid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_file(_name TEXT, _size BIGINT, _sha256 VARCHAR(64), _sha1 VARCHAR(40), _md5 VARCHAR(32), _symlink INTEGER, _fifo INTEGER) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _fid BIGINT;
    BEGIN
        INSERT INTO file (size, checksum_sha256, checksum_sha1, checksum_md5, flag_symlink, flag_fifo)
        VALUES (_size, _sha256, _sha1, _md5, _symlink, _fifo)
        ON CONFLICT (checksum_sha1) DO UPDATE SET checksum_md5=EXCLUDED.checksum_md5, size=EXCLUDED.size, checksum_sha256=EXCLUDED.checksum_sha256
        RETURNING id
        INTO _fid;

        IF _fid IS NOT DISTINCT FROM NULL THEN
            RAISE EXCEPTION 'Insert into raw file table did not return an ID';
            RETURN NULL;
        END IF;

        INSERT INTO file_alias (file_id, name) VALUES (_fid, _name)
        ON CONFLICT (file_id, name) DO UPDATE SET name=EXCLUDED.name
        RETURNING id
        INTO _fid;

        RETURN _fid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION list_tar_name(_s TEXT) RETURNS SETOF TEXT LANGUAGE plpgsql AS $$
    DECLARE
        _parts TEXT[];
        _part TEXT;
        _RET text := '';
        l INTEGER;
        i INTEGER;
    BEGIN
        _parts := regexp_split_to_array(_s, '\.');
        l := array_length(_parts, 1);

        IF l >= 3 AND _parts[l-1] = 'tar' THEN
            RETURN NEXT _s;

            _ret := _parts[1];
            FOR i IN 2..l-1 LOOP
                _ret := _ret || '.' || _parts[i];
            END LOOP;

            RETURN NEXT _ret;
        ELSIF l >= 2 AND _parts[l] = 'tar' THEN
            RETURN NEXT _s;
            RETURN NEXT _s || '.' || 'gz';
            RETURN NEXT _s || '.' || 'bz2';
            RETURN NEXT _s || '.' || 'xz';
        ELSE
            RETURN NEXT _s;
        END IF;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION merge_file_collections() RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        ret BIGINT;
        rec record;
    BEGIN
        FOR rec IN SELECT n.id as nid, a.id as aid, n.name as nname, a.name as aname, n.checksum_sha1 as nsha1, a.checksum_sha1 as asha1, n.file_collection_id as nfid, a.file_collection_id as afid, n.path as npath, a.path as apath FROM archive n INNER JOIN archive a ON n.name=a.name AND n.checksum_sha1 IS NOT NULL AND a.checksum_sha1 IS NULL LOOP
            RAISE NOTICE 'rec: %', rec;
            IF rec.aid IS NOT NULL THEN

                EXECUTE 'SELECT move_file_belongs($1,$2)'
                    USING rec.nfid, rec.afid;
                UPDATE file_collection_license SET file_collection_id=rec.afid WHERE file_collection_id=rec.nfid;

                UPDATE archive SET path=COALESCE(rec.npath, rec.apath), checksum_sha1=COALESCE(rec.nsha1, rec.asha1) WHERE id=rec.aid;

                DELETE FROM archive WHERE id=rec.nid;

                ret := ret + 1;
            END IF;
        END LOOP;

        RETURN ret;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION move_file_belongs(_source BIGINT, _target BIGINT) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        rec record;
        ret BIGINT;
        check BOOLEAN;
    BEGIN
        FOR rec IN SELECT * FROM file_belongs_collection WHERE file_collection_id=_source LOOP
            check := false;
            EXECUTE 'SELECT update_file_belongs($1,$2,$3,$4)'
                INTO check
                USING _source, rec.file_id, rec.path, _target;
            ret := ret + 1;
        END LOOP;

        RETURN ret;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION parse_group_path(path TEXT) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        parts TEXT[];
        start INTEGER := 1;
        part TEXT;

        current BIGINT;
        parent BIGINT;
    BEGIN
        SELECT regexp_split_to_array(path, '/') INTO parts;

        IF parts[0] = '' THEN
            start := 2;
        END if;

        FOR ind IN start..ARRAY_LENGTH(parts, 1)
        LOOP
            parent := current;
            current := NULl;
            part = parts[ind];

            INSERT INTO group_container (name, parent_id) VALUES (part, parent)
            ON CONFLICT (name, parent_id) DO UPDATE SET id=EXCLUDED.id
            RETURNING id INTO current;
        END LOOP;

        RETURN current;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION parse_license_expression(expression TEXT) RETURNS BIGINT LANGUAGE plperl AS $$
    use strict;

    my $eText = $_[0];

    if (not defined $eText)
    {
        return undef;
    }
    $eText =~ s/;/ ; /g;

    # Parse Tree
    my @parts = split(/\s+/, $eText);
    my @eParts;
    foreach my $p (@parts[0..$#parts]) {
        if($p eq 'AND' || $p eq 'OR' || $p eq ';') {
            push @eParts, $p;
        } else {
            my $tmp = pop @eParts;
            my $value;
            if (not defined $tmp) {
                $value = $p;
            } elsif ($tmp eq 'AND' || $tmp eq 'OR' || $tmp eq ';') {
                push @eParts, $tmp;
                $value = $p;
            } else {
                $value = $tmp.' '.$p;
            }
            push @eParts, $value;
        }
    }
    my $root = {};
    $root->{VALUE} = $eParts[0];
    my $current = \$root;
    foreach my $p (@eParts[1..$#eParts]) {
        if ($p eq 'AND' || $p eq 'OR' || $p eq ';') {
            my $newNode = {};
            $newNode->{VALUE} = ${$current}->{VALUE};
            $newNode->{LEFT} = ${$current}->{LEFT};
            $newNode->{RIGHT} = ${$current}->{RIGHT};
            ${$current}->{VALUE} = $p;
            ${$current}->{LEFT} = $newNode;
        } else {
            my $newNode = {};
            $newNode->{VALUE} = $p;
            ${$current}->{RIGHT} = $newNode;
        }
    }

    # Recursively build license_expression 
    sub get_License {
        my $node = $_[0];
        my $value = $node->{VALUE};

        if($value eq 'AND' || $value eq 'OR' || $value eq ';') {
            my $left;
            my $right;
            my $insertOperatorPlan = spi_prepare('INSERT INTO license_expression (type, operator, left_id, right_id) VALUES ($1, $2, $3, $4) RETURNING id', 'TEXT', 'TEXT', 'BIGINT', 'BIGINT');

            if(defined $node->{LEFT}) {
                $left = get_License($node->{LEFT});
            }
            if(defined $node->{RIGHT}) {
                $right = get_License($node->{RIGHT});
            }

            my $id = spi_exec_prepared($insertOperatorPlan, {limit=>1}, 'BOP', $value, $left, $right)->{rows}[0]->{id};
            return $id;
        } else {
            my $insertLicensePlan = spi_prepare('INSERT INTO license_expression (type, license_id) VALUES ($1, (SELECT get_license($2))) RETURNING id', 'TEXT', 'TEXT');
            my $id = spi_exec_prepared($insertLicensePlan, {limit=>1}, 'LICENSE', $value)->{rows}[0]->{id};
            return $id;
        }
    }

    # Get license_expression_id
    return get_License($root);
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION repair_rationales() RETURNS INTEGER LANGUAGE plpgsql AS $$
    DECLARE
        count INTEGER := 0;
        rec RECORD;
    BEGIN
        FOR rec IN SELECT id, associate_list, associate_license_rationale FROM file_collection WHERE associate_list LIKE '||%' ORDER BY id LOOP
            count = count+1;
            UPDATE file_collection SET associate_list='rationale', associate_license_rationale=rec.associate_license_rationale WHERE id=rec.id;
        END LOOP;
        RETURN count;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION select_file_collection_files(_cid BIGINT) RETURNS SETOF BIGINT LANGUAGE plpgsql AS $$
    BEGIN
        RETURN QUERY EXECUTE('SELECT file_id FROM file_belongs_collection WHERE file_collection_id IN (SELECT select_file_collections($1))') USING _cid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION select_file_collections(_cid BIGINT) RETURNS SETOF BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _row RECORD;
        _rec RECORD;
    BEGIN
        RETURN NEXT _cid;

        FOR _row IN SELECT child_id FROM file_collection_contains WHERE parent_id=_cid
        LOOP
            RETURN NEXT _row.child_id;
            FOR _rec IN EXECUTE('SELECT select_file_collections($1) as rid') USING _row.child_id
            LOOP
                RETURN NEXT _rec.rid;
            END LOOP;
        END LOOP;

        RETURN;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_file_collection_license(_cid BIGINT, _lid BIGINT, _rationale TEXT) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        _clid BIGINT;
        _expression TEXT;
    BEGIN
        INSERT INTO file_collection_license (file_collection_id, license_id, rationale) VALUES (_cid, _lid, _rationale) ON CONFLICT (file_collection_id, license_id, analyst_id) DO UPDATE SET rationale=EXCLUDED.rationale RETURNING id INTO _clid;
        IF _clid IS NOT NULL THEN
            EXECUTE 'SELECT build_license_expression($1)'
                INTO _expression
                USING _lid;
            IF _expression IS NOT NULL THEN
                UPDATE license_expression SET expression=_expression WHERE id=_lid;
            END IF;
        END IF;

        RETURN _clid;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION stub_archives() RETURNS INTEGER LANGUAGE plpgsql AS $$
    DECLARE
        count INTEGER := 0;
        rec RECORD;
    BEGIN
        FOR rec IN SELECT c.id, c.full_name, c.checksum_sha1, c.archive_size FROM file_collection c INNER JOIN archive a ON a.file_collection_id=c.id AND a.name=c.full_name ORDER BY c.id LOOP
            count = count+1;
            EXECUTE 'UPDATE archive SET checksum_sha1=$1, size=$2 WHERE file_collection_id=$3 AND name=$4'
                USING rec.checksum_sha1, rec.archive_size, rec.id, rec.full_name;
        END LOOP;

        FOR rec IN SELECT c.id, c.full_name, c.checksum_sha1, c.archive_size FROM file_collection c LEFT JOIN archive a ON a.file_collection_id=c.id AND a.name=c.full_name WHERE a.id IS NULL ORDER BY c.id LOOP
            count = count+1;
            EXECUTE 'INSERT INTO archive (file_collection_id, name, checksum_sha1, size) VALUES ($1, $2, $3, $4) ON CONFLICT(checksum_sha1) DO UPDATE SET checksum_sha1=$3'
                USING rec.id, rec.full_name, rec.checksum_sha1, rec.archive_size;
        END LOOP;

        RETURN count;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION stub_file_collection(vcode VARCHAR(40)) RETURNS BIGINT LANGUAGE plpgsql AS $$
    DECLARE
        ret BIGINT;
    BEGIN
        SELECT id FROM file_collection WHERE verification_code=vcode INTO ret;
        IF ret IS NULL THEN
            INSERT INTO file_collection (verification_code) VALUES (vcode) RETURNING id INTO ret;
        END IF;

        RETURN ret;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION transfer_licenses() RETURNS INTEGER LANGUAGE plpgsql AS $$
    DECLARE
        count INTEGER := 0;
        rec RECORD;
        lid BIGINT;
        exp TEXT;
        lic_info_id BIGINT;
    BEGIN
        FOR rec IN SELECT c.id, c.associate_list, c.associate_license_rationale FROM file_collection c LOOP
            count = count+1;
            EXECUTE 'SELECT parse_license_expression($1)'
                INTO lid
                USING rec.associate_list;
            IF lid IS NOT NULL THEN
                SELECT build_license_expression(lid) INTO exp;
                UPDATE license_expression SET expression=exp WHERE id=lid;
                INSERT INTO file_collection_license (file_collection_id, license_id, rationale, insert_date) VALUES (rec.id, lid, rec.associate_license_rationale, NOW()) RETURNING id INTO lic_info_id;
                UPDATE file_collection SET license_info_id=lic_info_id WHERE id=rec.id;
            END IF;
        END LOOP;
        RETURN count;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION transfer_files() RETURNS INTEGER LANGUAGE plpgsql AS $$
   DECLARE
        count INTEGER;
        rec RECORD;
    BEGIN
        FOR rec IN SELECT id, full_name FROM file LOOP
            count = count+1;
            INSERT INTO file_alias (file_id, name) VALUES (rec.id, rec.full_name);
        END LOOP;
        RETURN count;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION transfer_sub_collections() RETURNS INTEGER LANGUAGE plpgsql AS $$
    DECLARE
        count INTEGER;
        rec RECORD;
    BEGIN
        FOR rec IN SELECT container_id, instance_id FROM container_contains LOOP
            count = count+1;
            INSERT INTO file_collection_contains (parent_id, child_id) VALUES (container_id, instance_id);
        END LOOP;
        RETURN count;
    END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_file_belongs(_s_cid BIGINT, _s_fid BIGINT, _s_path TEXT, _target BIGINT) RETURNS BOOLEAN LANGUAGE plpgsql AS $$
    BEGIN
        UPDATE file_belongs_collection SET file_collection_id=_target WHERE file_collection_id=_s_cid AND file_id=_s_fid AND path=_s_path;
        RETURN true;
    EXCEPTION WHEN OTHERS THEN
        RETURN false;
    END;
$$;
-- +goose StatementEnd

CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;

--
--

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd