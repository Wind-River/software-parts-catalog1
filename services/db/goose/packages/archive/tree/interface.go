package tree

import "database/sql"

type Node interface {
	GetName() string
	GetFiles() []SubFile
	GetNodes() []SubNode
	GetFileVerificationCode() []byte

	SetName(string)
	AddFile(path string, file *File) int
	AddNode(path string, node Node) int
	SetFileVerificationCode([]byte)

	Merge(Node) error
	GetDuplicates() []Node

	GetLicense() sql.NullString
	GetLicenseRationale() sql.NullString
	GetLicenseNotice() sql.NullString
	GetAutomationLicense() sql.NullString
	GetAutomationLicenseRationale() sql.NullString
	SetLicense(string)
	SetLicenseRationale(string)
	SetLicenseNotice(string)
	SetAutomationLicense(string)
	SetAutomationLicenseRationale(string)
}
