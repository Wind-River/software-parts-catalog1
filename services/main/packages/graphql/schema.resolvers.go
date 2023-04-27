package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/core/part"
	"wrs/tk/packages/editDistance"
	"wrs/tk/packages/generics"
	"wrs/tk/packages/graphql/generated"
	"wrs/tk/packages/graphql/model"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	errWrapper "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Sha256 is the resolver for the sha256 field.
func (r *archiveResolver) Sha256(ctx context.Context, obj *model.Archive) (string, error) {
	return hex.EncodeToString(obj.Sha256[:]), nil
}

// PartID is the resolver for the part_id field.
func (r *archiveResolver) PartID(ctx context.Context, obj *model.Archive) (*string, error) {
	if obj.PartID == nil {
		return nil, nil
	}

	ret := obj.PartID.String()
	return &ret, nil
}

// Part is the resolver for the part field.
func (r *archiveResolver) Part(ctx context.Context, obj *model.Archive) (*model.Part, error) {
	if obj.PartID == nil {
		return nil, nil
	}

	p, err := r.PartController.GetByID(*obj.PartID)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error getting part %s", obj.PartID.String())
	}

	ret := model.ToPart(p)
	return &ret, nil
}

// Md5 is the resolver for the md5 field.
func (r *archiveResolver) Md5(ctx context.Context, obj *model.Archive) (*string, error) {
	ret := hex.EncodeToString(obj.Md5[:])
	return &ret, nil
}

// Sha1 is the resolver for the sha1 field.
func (r *archiveResolver) Sha1(ctx context.Context, obj *model.Archive) (*string, error) {
	ret := hex.EncodeToString(obj.Sha1[:])
	return &ret, nil
}

// AddPartList is the resolver for the addPartList field.
func (r *mutationResolver) AddPartList(ctx context.Context, name string, parentID *int64) (*model.PartList, error) {
	if parentID != nil && *parentID != 0 {
		pl, err := r.PartListController.AddPartListWithParent(name, *parentID)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	if name != "" {
		pl, err := r.PartListController.AddPartList(name)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	return nil, errWrapper.New("Name was invalid")
}

// DeletePartList is the resolver for the deletePartList field.
func (r *mutationResolver) DeletePartList(ctx context.Context, id int64) (*model.PartList, error) {
	if id != 0 {
		pl, err := r.PartListController.DeletePartList(id)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	return nil, nil
}

// DeletePartFromList is the resolver for the deletePartFromList field.
func (r *mutationResolver) DeletePartFromList(ctx context.Context, listID int64, partID string) (*model.PartList, error) {
	partUUID, err := uuid.Parse(partID)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error parsing part_id \"%s\"", partUUID)
	}
	if listID != 0 {
		pl, err := r.PartListController.DeletePartFromList(listID, partUUID)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	return nil, nil
}

// UploadArchive is the resolver for the uploadArchive field.
func (r *mutationResolver) UploadArchive(ctx context.Context, file graphql.Upload, name *string) (*model.UploadedArchive, error) {
	ret := new(model.UploadedArchive)

	// Save upload to tmp and verify size
	tmpHandOff := false
	tmpFile, err := os.CreateTemp("", "graphql_upload_archive.*")
	if err != nil {
		log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error creating temp file")
		return ret, err
	}
	defer func(filePath string) {
		if !tmpHandOff {
			os.Remove(filePath)
		}
	}(tmpFile.Name())

	var fileName string
	if name != nil {
		fileName = *name
	} else {
		fileName = file.Filename
	}

	written, err := io.Copy(tmpFile, file.File)
	if err != nil {
		log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error writing temp file")
		return ret, err
	}

	if written != file.Size {
		log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Int64("written", written).Int64("size", file.Size).Msg("file size does not match actual")
		return ret, errWrapper.New("file size does not match actual")
	}

	// Process archive
	arch, err := archive.InitArchive(tmpFile.Name(), fileName)
	if err != nil {
		log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error initializing archive")
		return ret, err
	}

	// Check if see if archive already known
	if remoteArchive, err := r.ArchiveController.GetBySha256(arch.Sha256[:]); err == nil {
		// ret.Extracted = remoteArchive.ExtractStatus > 0
		remoteModel := model.ToArchive(remoteArchive)
		ret.Archive = &remoteModel
		return ret, nil
	}

	log.Debug().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").
		Interface("arch", arch).Msg("processing archive in background")
	tmpHandOff = true
	go func(arch *archive.Archive, archiveController *archive.ArchiveController) error {
		defer os.Remove(arch.StoragePath.String)

		if err := archiveController.Process(arch); err != nil {
			log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error processing archive")
			return err
		}

		return nil
	}(arch, r.ArchiveController)

	// Format response
	// ret.Extracted = arch.ExtractStatus > 0
	modelArchive := model.ToArchive(arch)
	ret.Archive = &modelArchive
	return ret, nil
}

// UpdateArchive is the resolver for the updateArchive field.
func (r *mutationResolver) UpdateArchive(ctx context.Context, sha256 string, license *string, licenseRationale *string, familyString *string) (*model.Archive, error) {
	// parse sha256
	rawSha256, err := hex.DecodeString(sha256)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error decoding sha256 \"%s\"", sha256)
	}

	// Sanity check input
	if (license == nil || *license == "") && (licenseRationale == nil || len(*licenseRationale) == 0) && (familyString == nil || *familyString == "") {
		return nil, errWrapper.New("no data was given to update")
	}

	archive, err := r.ArchiveController.GetBySha256(rawSha256)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error getting archive")
	}

	if archive.PartID == nil {
		return nil, errWrapper.New("no part to update")
	}

	part, err := r.PartController.GetByID(*archive.PartID)
	if err != nil {
		return nil, errWrapper.New("error getting part")
	}

	if err := r.PartController.UpdateTribalKnowledge(part.PartID, nil, nil, nil, nil, familyString, nil, license, licenseRationale, nil, nil); err != nil {
		return nil, errWrapper.Wrapf(err, "error updating file_collection")
	}

	archive, err = r.ArchiveController.GetBySha256(archive.Sha256[:])
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error getting updated archive")
	}

	ret := model.ToArchive(archive)

	return &ret, nil
}

// UpdatePartList is the resolver for the updatePartList field.
func (r *mutationResolver) UpdatePartList(ctx context.Context, id int64, name *string, parts []*string) (*model.PartList, error) {
	if id != 0 && parts != nil {
		partIDS := make([]*uuid.UUID, len(parts))
		for i, v := range parts {
			partID, err := uuid.Parse(*v)
			if err != nil {
				return nil, err
			}
			partIDS[i] = &partID
		}
		pl, err := r.PartListController.AddParts(id, partIDS)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	return nil, nil
}

// UpdatePart is the resolver for the updatePart field.
func (r *mutationResolver) UpdatePart(ctx context.Context, partInput *model.PartInput) (*model.Part, error) {
	partUUID, err := uuid.Parse(partInput.ID)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error parsing part_id \"%s\"", partInput.ID)
	}

	var rawVerificationCode []byte
	if partInput.FileVerificationCode != nil {
		rawVerificationCode, err = hex.DecodeString(*partInput.FileVerificationCode)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error decoding verification code")
		}
	}

	p, err := r.PartController.GetByID(part.ID(partUUID))
	if err != nil {
		return nil, errWrapper.New("error getting part")
	}

	var comprised *part.ID
	if partInput.Comprised != nil {
		comprisedUUID, err := uuid.Parse(*partInput.Comprised)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error parsing comprised \"%s\"", *partInput.Comprised)
		}
		comprisedID := part.ID(comprisedUUID)

		comprised = &comprisedID
	}

	var partType *string
	if partInput.Type != nil && *partInput.Type != "" {
		lTree, err := model.TypeToLTree(*partInput.Type)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "invalid part type")
		}
		partType = &lTree
	}

	if err := r.PartController.UpdateTribalKnowledge(p.PartID,
		partType, partInput.Name, partInput.Version, partInput.Label, partInput.FamilyName,
		rawVerificationCode, partInput.License, partInput.LicenseRationale, partInput.Description, comprised); err != nil {
		return nil, errWrapper.Wrapf(err, "error updating part")
	}

	p, err = r.PartController.GetByID(p.PartID)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error getting updated part")
	}

	ret := model.ToPart(p)

	return &ret, nil
}

// CreateAlias is the resolver for the createAlias field.
func (r *mutationResolver) CreateAlias(ctx context.Context, id string, alias string) (string, error) {
	partUUID, err := uuid.Parse(id)
	if err != nil {
		return id, errWrapper.Wrapf(err, "error parsing id")
	}

	partID, err := r.PartController.CreateAlias(part.ID(partUUID), alias)
	if err != nil {
		return partID.String(), err
	}

	return partID.String(), nil
}

// AttachDocument is the resolver for the attachDocument field.
func (r *mutationResolver) AttachDocument(ctx context.Context, id string, key string, title *string, document string) (bool, error) {
	partUUID, err := uuid.Parse(id)
	if err != nil {
		return false, errWrapper.Wrapf(err, "error parsing id")
	}

	if err := r.PartController.AttachDocument(part.ID(partUUID), key, title, json.RawMessage(document)); err != nil {
		return false, errWrapper.Wrapf(err, "error attaching document")
	}

	return true, nil
}

// PartHasPart is the resolver for the partHasPart field.
func (r *mutationResolver) PartHasPart(ctx context.Context, parent string, child string, path string) (bool, error) {
	parentUUID, err := uuid.Parse(parent)
	if err != nil {
		return false, errWrapper.Wrapf(err, "error parsing parent_id")
	}
	childUUID, err := uuid.Parse(child)
	if err != nil {
		return false, errWrapper.Wrapf(err, "error parsing child_id")
	}

	if err := r.PartController.AddPartToPart(part.ID(childUUID), part.ID(parentUUID), path); err != nil {
		return false, err
	}

	return true, nil
}

// PartHasFile is the resolver for the partHasFile field.
func (r *mutationResolver) PartHasFile(ctx context.Context, id string, fileSha256 string, path *string) (bool, error) {
	panic(fmt.Errorf("not implemented: PartHasFile - partHasFile"))
}

// CreatePart is the resolver for the createPart field.
func (r *mutationResolver) CreatePart(ctx context.Context, partInput model.NewPartInput) (*model.Part, error) {
	toNullString := func(s *string) sql.NullString {
		if s == nil {
			return sql.NullString{}
		}

		return sql.NullString{
			Valid:  true,
			String: *s,
		}
	}

	var comprised uuid.UUID
	if partInput.Comprised != nil && *partInput.Comprised != "" {
		parsed, err := uuid.Parse(*partInput.Comprised)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error parsing comprised")
		}

		comprised = parsed
	}

	var partType sql.NullString
	if partInput.Type != nil && *partInput.Type != "" {
		lTree, err := model.TypeToLTree(*partInput.Type)
		if err != nil {
			return nil, err
		}
		partType.Valid = true
		partType.String = lTree
	}

	p, err := r.PartController.CreatePart(part.Part{
		Type:             partType,
		Name:             toNullString(partInput.Name),
		Version:          toNullString(partInput.Version),
		Label:            toNullString(partInput.Label),
		FamilyName:       toNullString(partInput.FamilyName),
		License:          toNullString(partInput.License),
		LicenseRationale: toNullString(partInput.LicenseRationale),
		Description:      toNullString(partInput.Description),
		Comprised:        part.ID(comprised),
	})
	if err != nil {
		return nil, err
	}

	ret := model.ToPart(p)

	return &ret, nil
}

// ID is the resolver for the id field.
func (r *partResolver) ID(ctx context.Context, obj *model.Part) (string, error) {
	return obj.ID.String(), nil
}

// FileVerificationCode is the resolver for the file_verification_code field.
func (r *partResolver) FileVerificationCode(ctx context.Context, obj *model.Part) (*string, error) {
	if len(obj.FileVerificationCode) == 0 {
		return nil, nil
	}

	encoded := hex.EncodeToString(obj.FileVerificationCode)
	return &encoded, nil
}

// License is the resolver for the license field.
func (r *partResolver) License(ctx context.Context, obj *model.Part) (*string, error) {
	return obj.License, nil
}

// Comprised is the resolver for the comprised field.
func (r *partResolver) Comprised(ctx context.Context, obj *model.Part) (*string, error) {
	if obj.Comprised == nil {
		return nil, nil
	}

	comprisedID := obj.Comprised.String()
	return &comprisedID, nil
}

// Aliases is the resolver for the aliases field.
func (r *partResolver) Aliases(ctx context.Context, obj *model.Part) ([]string, error) {
	aliases, err := r.PartController.GetAliases(obj.ID)
	if err != nil {
		return nil, err
	}

	return aliases, nil
}

// Profiles is the resolver for the profiles field.
func (r *partResolver) Profiles(ctx context.Context, obj *model.Part) ([]*model.Profile, error) {
	profiles, err := r.PartController.GetProfiles(obj.ID)
	if err != nil {
		return nil, err
	}

	ret, err := generics.Map[part.Profile, *model.Profile](profiles, func(p part.Profile) (*model.Profile, error) {
		ret := new(model.Profile)
		ret.Key = p.Key

		documents, err := generics.Map[part.Document, *model.Document](p.Documents, func(d part.Document) (*model.Document, error) {
			ret := new(model.Document)
			ret.Title = d.Title
			ret.Document = string(d.Document)

			return ret, nil
		})
		if err != nil {
			return nil, err
		}

		ret.Documents = documents

		return ret, nil
	})
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error mapping internal profiles to model profiles")
	}

	return ret, nil
}

// SubParts is the resolver for the sub_parts field.
func (r *partResolver) SubParts(ctx context.Context, obj *model.Part) ([]*model.SubPart, error) {
	subParts, err := r.PartController.SubParts(obj.ID)
	if err != nil {
		return nil, err
	}

	ret := make([]*model.SubPart, 0)
	for _, subPart := range subParts {
		prt, err := r.PartController.GetByID(subPart.ID)
		if err != nil {
			return ret, err
		}

		modelPart := model.ToPart(prt)

		ret = append(ret, &model.SubPart{
			Path: subPart.Path,
			Part: &modelPart,
		})
	}

	return ret, nil
}

// Archive is the resolver for the archive field.
func (r *queryResolver) Archive(ctx context.Context, sha256 *string, name *string) (*model.Archive, error) {
	// Fetch by sha256 if given
	if sha256 != nil && *sha256 != "" {
		rawSha256, err := hex.DecodeString(*sha256)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error decoding sha256 string \"%s\"", *sha256)
		}
		archve, err := r.ArchiveController.GetBySha256(rawSha256)
		if err != nil {
			log.Error().Err(err).Str(zerolog.CallerFieldName, "queryResolver.Archive").
				Str("sha256", *sha256).
				Msg("error getting archive by sha256")
			return nil, err
		}

		ret := model.ToArchive(archve)
		return &ret, nil
	}

	// Fetch by name if given
	if name != nil && *name != "" {
		archve, err := r.ArchiveController.GetByName(*name)
		if err != nil {
			log.Error().Err(err).Str(zerolog.CallerFieldName, "queryResolver.Archive").
				Str("name", *name).
				Msg("error getting archive by name")
			return nil, err
		}

		ret := model.ToArchive(archve)
		return &ret, nil
	}

	return nil, nil // should this be an error, no arguments found?
}

// FindArchive is the resolver for the find_archive field.
func (r *queryResolver) FindArchive(ctx context.Context, query string, method *string, costs *model.SearchCosts) ([]*model.ArchiveDistance, error) {
	var methodValue archive.SearchMethod
	if method == nil || *method == "" {
		methodValue = archive.ParseMethod("levenshtein")
	} else {
		methodValue = archive.ParseMethod(*method)
	}

	if costs == nil {
		costs = &model.SearchCosts{
			Insert:      editDistance.OPERATION_INSERT_COST,
			Delete:      editDistance.OPERATION_DELETE_COST,
			Substitute:  editDistance.OPERATION_SUBSTITUTE_COST,
			MaxDistance: &editDistance.DISTANCE_MAX,
		}
	}
	if costs.MaxDistance == nil {
		costs.MaxDistance = &editDistance.DISTANCE_MAX_NIL
	}

	log.Debug().Str("query", query).Interface("methodValue", methodValue).Msg("SearchForArchiveAll")
	distances, err := r.ArchiveController.SearchForArchiveAll(query, methodValue, costs.Insert, costs.Delete, costs.Substitute, *costs.MaxDistance)
	if err != nil {
		return nil, err
	}
	log.Debug().Interface("distances", distances).Msg("Found Archive Distances")
	ret := make([]*model.ArchiveDistance, 0, len(distances))
	log.Debug().Str("query", query).Interface("methodValue", methodValue).Msg("SearchForArchiveAll")
	for _, v := range distances {
		if costs.MaxDistance != nil && *costs.MaxDistance != editDistance.DISTANCE_MAX_NIL && v.Distance > int64(*costs.MaxDistance) {
			// if max distance in use reject distances above it
			continue
		}
		internalArchive, err := r.ArchiveController.GetBySha256(v.Sha256[:])
		if err != nil {
			return nil, err
		}
		a := model.ToArchive(internalArchive)
		a.Name = v.MatchedName

		ret = append(ret, &model.ArchiveDistance{
			Distance: v.Distance,
			Archive:  &a,
		})
	}

	return ret, nil
}

// Part is the resolver for the part field.
func (r *queryResolver) Part(ctx context.Context, id *string, fileVerificationCode *string, sha256 *string, sha1 *string, name *string) (*model.Part, error) {
	if id != nil && *id != "" {
		partUUID, err := uuid.Parse(*id)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error parsing part_id \"%s\"", *id)
		}
		p, err := r.PartController.GetByID(part.ID(partUUID))
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error getting part by id: \"%s\"", partUUID.String())
		}

		ret := model.ToPart(p)
		return &ret, nil
	}
	if fileVerificationCode != nil && *fileVerificationCode != "" {
		rawFVC, err := hex.DecodeString(*fileVerificationCode)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error decoding file_verification_code: \"%s\"", *fileVerificationCode)
		}
		p, err := r.PartController.GetByVerificationCode(rawFVC)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error getting part by file_verification_code: \"%s\"", *fileVerificationCode)
		}

		ret := model.ToPart(p)
		return &ret, nil
	}
	if sha256 != nil && *sha256 != "" {
		rawSha256, err := hex.DecodeString(*sha256)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error decoding sha256: \"%s\"", *sha256)
		}
		a, err := r.ArchiveController.GetBySha256(rawSha256)
		if err != nil {
			return nil, err
		}
		if a.PartID == nil {
			return nil, nil
		}
		p, err := r.PartController.GetByID(*a.PartID)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error getting part by id: \"%s\"", a.PartID.String())
		}

		ret := model.ToPart(p)
		return &ret, nil
	}
	if sha1 != nil && *sha1 != "" {
		rawSha1, err := hex.DecodeString(*sha1)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error decoding sha1: \"%s\"", *sha1)
		}
		a, err := r.ArchiveController.GetBySha1(rawSha1)
		if err != nil {
			return nil, err
		}
		if a.PartID == nil {
			return nil, nil
		}
		p, err := r.PartController.GetByID(*a.PartID)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error getting part by id: \"%s\"", a.PartID.String())
		}

		ret := model.ToPart(p)
		return &ret, nil
	}
	if name != nil && *name != "" {
		a, err := r.ArchiveController.GetByName(*name)
		if err != nil {
			return nil, err
		}
		if a.PartID == nil {
			return nil, nil
		}
		p, err := r.PartController.GetByID(*a.PartID)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error getting part by id: \"%s\"", a.PartID.String())
		}

		ret := model.ToPart(p)
		return &ret, nil
	}

	return nil, nil // Should this be an error, no arguments found?
}

// Archives is the resolver for the archives field.
func (r *queryResolver) Archives(ctx context.Context, id *string, vcode *string) ([]*model.Archive, error) {
	var partID part.ID

	if id == nil || *id == "" {
		if vcode == nil || *vcode == "" {
			return nil, errWrapper.New("no identifiers provided")
		}

		rawVerificationCode, err := hex.DecodeString(*vcode)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error decoding verification_code")
		}

		fileCollection, err := r.PartController.GetByVerificationCode(rawVerificationCode)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error fetching file_collection")
		}

		partID = fileCollection.PartID
	} else {
		var err error
		partUUID, err := uuid.Parse(*id)
		if err != nil {
			return nil, errWrapper.Wrapf(err, "error parsing part_id: \"%s\"", *id)
		}

		partID = part.ID(partUUID)
	}

	archives, err := r.ArchiveController.GetByPart(partID)
	if err != nil {
		return nil, errWrapper.Wrap(err, "error fetching archives")
	}

	ret := make([]*model.Archive, len(archives))

	for i, v := range archives {
		m := model.ToArchive(&v)
		ret[i] = &m
	}

	return ret, nil
}

// Partlist is the resolver for the partlist field.
func (r *queryResolver) Partlist(ctx context.Context, id *int64, name *string) (*model.PartList, error) {
	if id != nil && *id != 0 {
		pl, err := r.PartListController.GetByID(*id)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	if name != nil && *name != "" {
		pl, err := r.PartListController.GetByName(*name)
		if err != nil {
			return nil, err
		}
		ret := model.ToPartList(pl)
		return &ret, nil
	}
	return nil, nil
}

// Partlist is the resolver for the partlist field.
func (r *queryResolver) PartlistParts(ctx context.Context, id int64) ([]*model.Part, error) {
	if id != 0 {
		parts, err := r.PartListController.GetParts(id)
		if err != nil {
			return nil, err
		}
		ret := make([]*model.Part, len(parts))
		for i, v := range parts {
			part := model.ToPart(v)
			ret[i] = &part
		}
		return ret, nil
	}
	return nil, nil
}

// Partlists is the resolver for the partlists field.
func (r *queryResolver) Partlists(ctx context.Context, parentID int64) ([]*model.PartList, error) {
	partlists, err := r.PartListController.GetByParentID(parentID)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error fetching partlists")
	}

	ret := make([]*model.PartList, len(partlists))

	for i, v := range partlists {
		m := model.ToPartList(&v)
		ret[i] = &m
	}
	return ret, nil
}

// FileCount is the resolver for the file_count field.
func (r *queryResolver) FileCount(ctx context.Context, id *string, vcode *string) (int64, error) {
	if (id == nil || *id == "") && (vcode == nil || *vcode == "") {
		return 0, errWrapper.New("no identifiers provided")
	}

	var partID part.ID

	if id != nil && *id != "" {
		partUUID, err := uuid.Parse(*id)
		if err != nil {
			return 0, errWrapper.Wrapf(err, "error parsing part_id \"%s\"", *id)
		}

		partID = part.ID(partUUID)
	} else {
		fileVerificationCode, err := hex.DecodeString(*vcode)
		if err != nil {
			return 0, errWrapper.Wrapf(err, "error decoding verification_code")
		}

		p, err := r.PartController.GetByVerificationCode(fileVerificationCode)
		if err != nil {
			return 0, errWrapper.Wrapf(err, "error getting part")
		}

		partID = p.PartID
	}

	count, err := r.PartController.CountFiles(partID)
	if err != nil {
		return count, errWrapper.Wrapf(err, "error counting files")
	}

	return count, nil
}

// Comprised is the resolver for the comprised field.
func (r *queryResolver) Comprised(ctx context.Context, id *string) ([]*model.Part, error) {
	if id == nil {
		return nil, nil
	}
	comprisedID, err := uuid.Parse(*id)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error parsing id")
	}

	parts, err := r.PartController.GetByComprised(part.ID(comprisedID))
	if err != nil {
		return nil, err
	}

	ret, err := generics.Map[part.Part, *model.Part](parts, func(p part.Part) (*model.Part, error) {
		ret := model.ToPart(&p)
		return &ret, nil
	})
	if err != nil {
		return ret, nil
	}

	return ret, nil
}

// Profile is the resolver for the profile field.
func (r *queryResolver) Profile(ctx context.Context, id *string, key *string) ([]*model.Document, error) {
	if id == nil || key == nil {
		return nil, nil
	}

	partId, err := uuid.Parse(*id)
	if err != nil {
		return nil, errWrapper.Wrapf(err, "error parsing id")
	}

	profile, err := r.PartController.GetProfile(part.ID(partId), *key)
	if err != nil {
		return nil, err
	}

	ret, err := generics.Map[part.Document, *model.Document](profile.Documents, func(d part.Document) (*model.Document, error) {
		ret := new(model.Document)
		ret.Title = d.Title
		ret.Document = string(d.Document)

		return ret, nil
	})
	if err != nil {
		return ret, err
	}

	return ret, nil
}

// Archive returns generated.ArchiveResolver implementation.
func (r *Resolver) Archive() generated.ArchiveResolver { return &archiveResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Part returns generated.PartResolver implementation.
func (r *Resolver) Part() generated.PartResolver { return &partResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type archiveResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type partResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
