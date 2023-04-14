package tree

type FileCollection struct {
	Name string
	LicenseData
	// Relationships
	Files []SubFile
	Nodes []SubNode

	FileVerificationCode []byte
	Duplicates           []Node // All archives should be inserted into the database, but the purpose of the trees is actually to turn them into parts, so a separate list of duplicates is required
}

func (fc FileCollection) GetName() string {
	return fc.Name
}

func (fc *FileCollection) SetName(s string) {
	fc.Name = s
}

func (fc FileCollection) GetFiles() []SubFile {
	return fc.Files
}

func (fc FileCollection) GetNodes() []SubNode {
	return fc.Nodes
}

func (fc FileCollection) GetFileVerificationCode() []byte {
	return fc.FileVerificationCode
}

func (fc *FileCollection) SetFileVerificationCode(code []byte) {
	fc.FileVerificationCode = code
}

func (fc *FileCollection) AddFile(path string, file *File) int {
	if fc.Files == nil {
		fc.Files = make([]SubFile, 0)
	}

	fc.Files = append(fc.Files, SubFile{
		Path: path,
		File: file,
	})

	return len(fc.Files)
}

func (fc *FileCollection) AddNode(path string, node Node) int {
	if fc.Nodes == nil {
		fc.Nodes = make([]SubNode, 0)
	}

	fc.Nodes = append(fc.Nodes, SubNode{
		Path: path,
		Node: node,
	})

	return len(fc.Nodes)
}

func (fc *FileCollection) Merge(node Node) error {
	if fc.Duplicates == nil {
		fc.Duplicates = make([]Node, 0)
	}

	fc.Files = node.GetFiles()
	fc.Nodes = node.GetNodes()
	fc.FileVerificationCode = node.GetFileVerificationCode()
	fc.Duplicates = append(fc.Duplicates, node)

	if fc.Name == "" {
		fc.Name = node.GetName()
	}

	return nil
}

func (fc FileCollection) GetDuplicates() []Node {
	return fc.Duplicates
}
