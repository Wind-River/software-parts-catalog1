package archive_web

import (
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"wrs/tk/packages/core/archive"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HandleArchiveDownload expects an archive sha256 and possibly an archive name, and then serves that archive.
// If no name is given, it picks the first one known in the database and redirects to that, because we want the end of the path to be the file name, so even a direct browser download will use the correct file name.
// The function depends on an archive controller from the request context to get the archive itself
func HandleArchiveDownload(w http.ResponseWriter, r *http.Request) {
	aSha256String := chi.URLParam(r, "archiveSha256")
	aSha256, err := hex.DecodeString(aSha256String)
	if err != nil {
		http.Error(w, "error decoding sha256", 400)
		log.Error().Err(err).Str("sha256_string", aSha256String).Msg("error decoding sha256")
		return
	}

	archiveName := chi.URLParam(r, "archiveName")

	archiveController, err := archive.GetArchiveController(r.Context())
	if err != nil {
		http.Error(w, "error getting archive controller", 500)
		log.Error().Err(err).Msg("error getting archive controller")
		return
	}

	arch, err := archiveController.GetBySha256(aSha256)
	if err == archive.ErrNotFound {
		log.Debug().Str(zerolog.CallerFieldName, "HandleContainerDownload").Hex("sha256", aSha256).Msg("Returning 404 on missing archive")
		http.Error(w, "archive not found", 404)
		return
	} else if err != nil {
		http.Error(w, "error selecting archive by Sha256", 500)
		log.Error().Err(err).Msg("error selecting archive to serve")
		return
	}

	if archiveName == "" { // No name given
		if len(arch.Aliases) > 0 {
			archiveName = arch.Aliases[0]
			http.Redirect(w, r, aSha256String+"/"+archiveName, http.StatusTemporaryRedirect)
			return
		} else { // No name exists for the archive
			archiveName = aSha256String
		}
	}

	// Download archive to a temporary directory that is deleted afterwards
	tmpDir, err := os.MkdirTemp("", "tk-serve")
	if err != nil {
		http.Error(w, "error serving file", 500)
		log.Error().Err(err).Msg("error serving file")
		return
	}
	defer os.RemoveAll(tmpDir)

	f, err := os.Create(filepath.Join(tmpDir, archiveName))
	if err != nil {
		http.Error(w, "error serving file", 500)
		log.Error().Err(err).Msg("error creating file in temp directory")
		return
	}
	defer f.Close()

	if err := archiveController.DownloadTo(arch, f); err != nil {
		http.Error(w, "error downloading archive", 500)
		log.Error().Err(err).Msg("error downloading archive")
		return
	}
	if _, err := f.Seek(0, 0); err != nil {
		http.Error(w, "error serving archive", 500)
		log.Error().Err(err).Msg("error seeking archive")
		return
	}

	w.Header().Set("File-Name", archiveName) // not necessary if all clients just get the file-name from the path

	http.ServeFile(w, r, filepath.Join(tmpDir, archiveName))
}
