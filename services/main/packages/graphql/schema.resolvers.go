package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"io"
	"os"
	"wrs/tk/packages/core/archive"
	"wrs/tk/packages/graphql/generated"
	"wrs/tk/packages/graphql/model"

	"github.com/99designs/gqlgen/graphql"
	exceptions "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FileCollection is the resolver for the file_collection field.
func (r *archiveResolver) FileCollection(ctx context.Context, obj *model.Archive) (*model.FileCollection, error) {
	if obj.FileCollectionID != nil && *obj.FileCollectionID > 0 {
		fc, err := r.FileCollectionController.GetByID(*obj.FileCollectionID)
		if err != nil {
			return nil, err
		}

		ret := model.ToFileCollection(fc)
		return &ret, nil
	}

	return nil, nil
}

// FlagExtract is the resolver for the flag_extract field.
func (r *fileCollectionResolver) FlagExtract(ctx context.Context, obj *model.FileCollection) (*bool, error) {
	return &obj.Extracted, nil
}

// FlagLicenseExtract is the resolver for the flag_license_extract field.
func (r *fileCollectionResolver) FlagLicenseExtract(ctx context.Context, obj *model.FileCollection) (*bool, error) {
	return &obj.LicenseExtracted, nil
}

// License is the resolver for the license field.
func (r *fileCollectionResolver) License(ctx context.Context, obj *model.FileCollection) (*model.License, error) {
	if obj.LicenseID != nil && *obj.LicenseID > 0 {
		l, err := r.LicenseController.GetByID(*obj.LicenseID)
		if err != nil {
			return nil, err
		}

		ret := model.ToLicense(l)
		return &ret, nil
	}

	return nil, nil
}

// VerificationCodeOne is the resolver for the verification_code_one field.
func (r *fileCollectionResolver) VerificationCodeOne(ctx context.Context, obj *model.FileCollection) (*string, error) {
	var ret string
	if obj.FVCOne != nil && len(obj.FVCOne) > 0 {
		ret = hex.EncodeToString(obj.FVCOne)
	}

	return &ret, nil
}

// VerificationCodeTwo is the resolver for the verification_code_two field.
func (r *fileCollectionResolver) VerificationCodeTwo(ctx context.Context, obj *model.FileCollection) (*string, error) {
	var ret string
	if obj.FVCTwo != nil && len(obj.FVCTwo) > 0 {
		ret = hex.EncodeToString(obj.FVCTwo)
	}

	return &ret, nil
}

// UploadArchive is the resolver for the uploadArchive field.
func (r *mutationResolver) UploadArchive(ctx context.Context, file graphql.Upload, name *string) (*model.UploadedArchive, error) {
	ret := new(model.UploadedArchive)

	// Save upload to tmp and verify size
	tmpFile, err := os.CreateTemp("", "graphql_upload_archive.*")
	if err != nil {
		log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error creating temp file")
		return ret, err
	}
	defer func(filePath string) {
		os.Remove(filePath)
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
		return ret, exceptions.New("file size does not match actual")
	}

	// Process archive
	arch, err := archive.InitArchive(tmpFile.Name(), fileName)
	if err != nil {
		log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error initializing archive")
		return ret, err
	}

	// Check if see if archive already known
	if remoteArchive, err := r.ArchiveController.GetBySha256(arch.Sha256.String); err == nil {
		ret.Extracted = remoteArchive.ExtractStatus > 0
		remoteModel := model.ToArchive(remoteArchive)
		ret.Archive = &remoteModel
		return ret, nil
	}

	log.Debug().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").
		Interface("arch", arch).Msg("processing archive in background")
	go func(arch *archive.Archive, archiveController *archive.ArchiveController) error {
		if err := archiveController.Process(arch); err != nil {
			log.Error().Str(zerolog.CallerFieldName, "mutationResolver.UploadArchive").Err(err).Msg("error processing archive")
			return err
		}

		return nil
	}(arch, r.ArchiveController)

	// Format response
	ret.Extracted = arch.ExtractStatus > 0
	modelArchive := model.ToArchive(arch)
	ret.Archive = &modelArchive
	return ret, nil
}

// UpdateArchive is the resolver for the updateArchive field.
func (r *mutationResolver) UpdateArchive(ctx context.Context, sha256 string, license *string, licenseRationale *string, familyString *string) (*model.Archive, error) {
	// Sanity check input
	if (license == nil || *license == "") && (licenseRationale == nil || *licenseRationale == "") && (familyString == nil || *familyString == "") {
		return nil, exceptions.New("no data was given to update")
	}

	archive, err := r.ArchiveController.GetBySha256(sha256)
	if err != nil {
		return nil, exceptions.Wrapf(err, "error getting archive")
	}

	if !archive.FileCollectionID.Valid {
		return nil, exceptions.New("no file_collection to update")
	}

	fileCollection, err := r.FileCollectionController.GetByID(archive.FileCollectionID.Int64)
	if err != nil {
		return nil, exceptions.New("error getting file_collection")
	}

	var licenseID int64
	if license != nil && *license != "" {
		licenseID, err = r.LicenseController.ParseLicenseExpression(*license)
		if err != nil {
			return nil, exceptions.Wrapf(err, "error parsing license expression")
		}
	}

	var rationale string
	if licenseRationale != nil {
		rationale = *licenseRationale
	}
	var path string
	if familyString != nil {
		path = *familyString
	}

	if err := r.FileCollectionController.UpdateTribalKnowledge(fileCollection.FileCollectionID, licenseID, rationale, path); err != nil {
		return nil, exceptions.Wrapf(err, "error updating file_collection")
	}

	archive, err = r.ArchiveController.GetByID(archive.ArchiveID)
	if err != nil {
		return nil, exceptions.Wrapf(err, "error getting updated archive")
	}

	ret := model.ToArchive(archive)

	return &ret, nil
}

// UpdateFileCollection is the resolver for the updateFileCollection field.
func (r *mutationResolver) UpdateFileCollection(ctx context.Context, verificationCode string, license *string, licenseRationale *string, familyString *string) (*model.FileCollection, error) {
	// Sanity check input
	if (license == nil || *license == "") && (licenseRationale == nil || *licenseRationale == "") && (familyString == nil || *familyString == "") {
		return nil, exceptions.New("no data was given to update")
	}

	rawVerificationCode, err := hex.DecodeString(verificationCode)
	if err != nil {
		return nil, exceptions.Wrapf(err, "error decoding verification code")
	}
	fileCollection, err := r.FileCollectionController.GetByVerificationCode(rawVerificationCode)
	if err != nil {
		return nil, exceptions.New("error getting file_collection")
	}

	var licenseID int64
	if license != nil && *license != "" {
		licenseID, err = r.LicenseController.ParseLicenseExpression(*license)
		if err != nil {
			return nil, exceptions.Wrapf(err, "error parsing license expression")
		}
	}

	var rationale string
	if licenseRationale != nil {
		rationale = *licenseRationale
	}
	var path string
	if familyString != nil {
		path = *familyString
	}

	if err := r.FileCollectionController.UpdateTribalKnowledge(fileCollection.FileCollectionID, licenseID, rationale, path); err != nil {
		return nil, exceptions.Wrapf(err, "error updating file_collection")
	}

	fileCollection, err = r.FileCollectionController.GetByID(fileCollection.FileCollectionID)
	if err != nil {
		return nil, exceptions.Wrapf(err, "error getting updated file_collection")
	}

	ret := model.ToFileCollection(fileCollection)

	return &ret, nil
}

// TestArchive is the resolver for the test_archive field.
func (r *queryResolver) TestArchive(ctx context.Context) (*model.Archive, error) {
	a, err := r.ArchiveController.GetByID(1)
	if err != nil {
		return nil, err
	}

	ret := model.ToArchive(a)
	return &ret, nil
}

// Archive is the resolver for the archive field.
func (r *queryResolver) Archive(ctx context.Context, id *int64, sha256 *string, sha1 *string, name *string) (*model.Archive, error) {
	if id != nil && *id != 0 {
		a, err := r.ArchiveController.GetByID(*id)
		if err != nil {
			return nil, err
		}

		ret := model.ToArchive(a)
		return &ret, nil
	}
	if sha256 != nil && *sha256 != "" {
		a, err := r.ArchiveController.GetBySha256(*sha256)
		if err != nil {
			return nil, err
		}

		ret := model.ToArchive(a)
		return &ret, nil
	}
	if sha1 != nil && *sha1 != "" {
		a, err := r.ArchiveController.GetBySha1(*sha1)
		if err != nil {
			return nil, err
		}

		ret := model.ToArchive(a)
		return &ret, nil
	}
	if name != nil && *name != "" {
		a, err := r.ArchiveController.GetByName(*name)
		if err != nil {
			return nil, err
		}

		ret := model.ToArchive(a)
		return &ret, nil
	}

	return nil, nil // Should this be an error, no arguments found?
}

// FindArchive is the resolver for the find_archive field.
func (r *queryResolver) FindArchive(ctx context.Context, query string, method *string) ([]*model.ArchiveDistance, error) {
	var methodValue archive.SearchMethod
	if method == nil || *method == "" {
		methodValue = archive.ParseMethod("levenshtein")
	} else {
		methodValue = archive.ParseMethod(*method)
	}

	log.Debug().Str("query", query).Interface("methodValue", methodValue).Msg("SearchForArchiveAll")
	distances, err := r.ArchiveController.SearchForArchiveAll(query, methodValue)
	if err != nil {
		return nil, err
	}
	ret := make([]*model.ArchiveDistance, len(distances))
	log.Debug().Str("query", query).Interface("methodValue", methodValue).Msg("SearchForArchiveAll")
	for i, v := range distances {
		internalArchive, err := r.ArchiveController.GetByID(v.ArchiveID)
		if err != nil {
			return nil, err
		}
		a := model.ToArchive(internalArchive)

		ret[i] = &model.ArchiveDistance{
			Distance: v.Distance,
			Archive:  &a,
		}
	}

	return ret, nil
}

// FileCollection is the resolver for the file_collection field.
func (r *queryResolver) FileCollection(ctx context.Context, id *int64, sha256 *string, sha1 *string, name *string) (*model.FileCollection, error) {
	if id != nil && *id != 0 {
		fc, err := r.FileCollectionController.GetByID(*id)
		if err != nil {
			return nil, err
		}

		ret := model.ToFileCollection(fc)
		return &ret, nil
	}
	if sha256 != nil && *sha256 != "" {
		a, err := r.ArchiveController.GetBySha256(*sha256)
		if err != nil {
			return nil, err
		}
		fc, err := r.FileCollectionController.GetByID(a.FileCollectionID.Int64)
		if err != nil {
			return nil, err
		}

		ret := model.ToFileCollection(fc)
		return &ret, nil
	}
	if sha1 != nil && *sha1 != "" {
		a, err := r.ArchiveController.GetBySha1(*sha1)
		if err != nil {
			return nil, err
		}
		fc, err := r.FileCollectionController.GetByID(a.FileCollectionID.Int64)
		if err != nil {
			return nil, err
		}

		ret := model.ToFileCollection(fc)
		return &ret, nil
	}
	if name != nil && *name != "" {
		a, err := r.ArchiveController.GetByName(*name)
		if err != nil {
			return nil, err
		}
		fc, err := r.FileCollectionController.GetByID(a.FileCollectionID.Int64)
		if err != nil {
			return nil, err
		}

		ret := model.ToFileCollection(fc)
		return &ret, nil
	}

	return nil, nil // Should this be an error, no arguments found?
}

// Archives is the resolver for the archives field.
func (r *queryResolver) Archives(ctx context.Context, id *int64, vcode *string) ([]*model.Archive, error) {
	if id == nil || *id == 0 {
		if vcode == nil || *vcode == "" {
			return nil, exceptions.New("no identifiers provided")
		}

		rawVerificationCode, err := hex.DecodeString(*vcode)
		if err != nil {
			return nil, exceptions.Wrapf(err, "error decoding verification_code")
		}

		fileCollection, err := r.FileCollectionController.GetByVerificationCode(rawVerificationCode)
		if err != nil {
			return nil, exceptions.Wrapf(err, "error fetching file_collection")
		}

		id = &fileCollection.FileCollectionID
	}

	archives, err := r.ArchiveController.GetByFileCollection(*id)
	if err != nil {
		return nil, exceptions.Wrap(err, "error fetching archives")
	}

	ret := make([]*model.Archive, len(archives))

	for i, v := range archives {
		m := model.ToArchive(&v)
		ret[i] = &m
	}

	return ret, nil
}

// FileCount is the resolver for the file_count field.
func (r *queryResolver) FileCount(ctx context.Context, id *int64, vcode *string) (int64, error) {
	if (id == nil || *id == 0) && (vcode == nil || *vcode == "") {
		return 0, exceptions.New("no identifiers provided")
	}

	var fileCollectionID int64

	if id != nil && *id != 0 {
		fileCollectionID = *id
	} else {
		fileVerificationCode, err := hex.DecodeString(*vcode)
		if err != nil {
			return 0, exceptions.Wrapf(err, "error decoding verification_code")
		}

		fileCollection, err := r.FileCollectionController.GetByVerificationCode(fileVerificationCode)
		if err != nil {
			return 0, exceptions.Wrapf(err, "error getting file_collection")
		}

		fileCollectionID = fileCollection.FileCollectionID
	}

	count, err := r.FileCollectionController.CountFiles(fileCollectionID)
	if err != nil {
		return count, exceptions.Wrapf(err, "error counting files")
	}

	return count, nil
}

// Archive returns generated.ArchiveResolver implementation.
func (r *Resolver) Archive() generated.ArchiveResolver { return &archiveResolver{r} }

// FileCollection returns generated.FileCollectionResolver implementation.
func (r *Resolver) FileCollection() generated.FileCollectionResolver {
	return &fileCollectionResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type archiveResolver struct{ *Resolver }
type fileCollectionResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
