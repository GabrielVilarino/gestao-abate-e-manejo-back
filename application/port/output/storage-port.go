package output

type StoragePort interface {
	Download(fotoURL string) ([]byte, error)
	Upload(data []byte) (string, error)
}
