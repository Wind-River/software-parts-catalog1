package tree

import (
	"database/sql"
	"path/filepath"

	"golang.org/x/text/runes"
)

type Sha256 [32]byte

type File struct {
	Sha256 Sha256
	Size   int64
	Md5    [16]byte
	Sha1   [20]byte
}

type SubFile struct {
	*File
	Path string
}

func (sf SubFile) GetPath() string {
	return runes.ReplaceIllFormed().String(sf.Path)
}

func (sf SubFile) GetName() string {
	return runes.ReplaceIllFormed().String(filepath.Base(sf.Path))
}

type ArchiveIdentifiers struct {
	// Identifying information
	Sha256 Sha256
	Size   int64
	Md5    [16]byte
	Sha1   [20]byte
	// Misc
	Name string
}

func (i *ArchiveIdentifiers) SetName(s string) {
	i.Name = s
}

func (i ArchiveIdentifiers) GetName() string {
	return runes.ReplaceIllFormed().String(i.Name)
}

type LicenseData struct {
	License                    sql.NullString
	LicenseRationale           sql.NullString
	LicenseNotice              sql.NullString
	AutomationLicense          sql.NullString
	AutomationLicenseRationale sql.NullString
}

func (ld LicenseData) GetLicense() sql.NullString {
	return ld.License
}
func (ld LicenseData) GetLicenseRationale() sql.NullString {
	return ld.LicenseRationale
}
func (ld LicenseData) GetLicenseNotice() sql.NullString {
	return ld.LicenseNotice
}
func (ld LicenseData) GetAutomationLicense() sql.NullString {
	return ld.AutomationLicense
}
func (ld LicenseData) GetAutomationLicenseRationale() sql.NullString {
	return ld.AutomationLicenseRationale
}

func (ld *LicenseData) SetLicense(s string) {
	if s == "" {
		return
	}
	ld.License.String = s
	ld.License.Valid = true
}
func (ld *LicenseData) SetLicenseRationale(s string) {
	if s == "" {
		return
	}

	ld.LicenseRationale.String = s
	ld.LicenseRationale.Valid = true
}
func (ld *LicenseData) SetLicenseNotice(s string) {
	if s == "" {
		return
	}

	ld.LicenseNotice.String = s
	ld.LicenseNotice.Valid = true
}
func (ld *LicenseData) SetAutomationLicense(s string) {
	if s == "" {
		return
	}

	ld.AutomationLicense.String = s
	ld.AutomationLicense.Valid = true
}
func (ld *LicenseData) SetAutomationLicenseRationale(s string) {
	if s == "" {
		return
	}

	ld.AutomationLicenseRationale.String = s
	ld.AutomationLicenseRationale.Valid = true
}

type Archive struct {
	ArchiveIdentifiers
	LicenseData
	// Relationships
	Files []SubFile
	Nodes []SubNode

	FileVerificationCode []byte
	Duplicates           []Node // All archives should be inserted into the database, but the purpose of the trees is actually to turn them into parts, so a separate list of duplicates is required
}

func (a Archive) GetFiles() []SubFile {
	return a.Files
}

func (a Archive) GetNodes() []SubNode {
	return a.Nodes
}

func (a Archive) GetFileVerificationCode() []byte {
	return a.FileVerificationCode
}

func (a *Archive) SetFileVerificationCode(code []byte) {
	a.FileVerificationCode = code
}

func (a *Archive) AddFile(path string, file *File) int {
	if a.Files == nil {
		a.Files = make([]SubFile, 0)
	}

	a.Files = append(a.Files, SubFile{
		Path: path,
		File: file,
	})

	return len(a.Files)
}

func (a *Archive) AddNode(path string, node Node) int {
	if a.Nodes == nil {
		a.Nodes = make([]SubNode, 0)
	}

	a.Nodes = append(a.Nodes, SubNode{
		Path: path,
		Node: node,
	})

	return len(a.Nodes)
}

func (a *Archive) Merge(node Node) error {
	if a.Duplicates == nil {
		a.Duplicates = make([]Node, 0)
	}

	a.Files = node.GetFiles()
	a.Nodes = node.GetNodes()
	a.FileVerificationCode = node.GetFileVerificationCode()
	a.Duplicates = append(a.Duplicates, node)

	if a.Name == "" {
		a.Name = node.GetName()
	}

	return nil
}

func (a *Archive) GetDuplicates() []Node {
	return a.Duplicates
}

type SubNode struct {
	Path string
	Node Node
}
