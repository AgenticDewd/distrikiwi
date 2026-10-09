package vector

import "errors"
var EmptyVectorError = errors.New("vector cannot be empty")
type VectorRecord struct {
	Key string
	Vector []float32
	Value [] byte
}

type SearchResult struct {
	Key string
	Value []byte
	Similarity float32
}

type VectorStore interface {
	Upsert(key string , vector []float32, value []byte) error

	Search(query []float32, limit int, threshold float32) ([]SearchResult, error)

	Close() error
}
