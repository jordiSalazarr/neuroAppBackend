package domain

type FileSigner interface {
	SignFile(filePath string) (string, error) // Returns the signed URL or path
}
