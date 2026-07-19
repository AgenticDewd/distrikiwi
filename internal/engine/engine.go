package engine
import "errors"
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyEmpty = errors.New("key is empty")
	ErrKeyAlreadyExists = errors.New("key already exists")
)
type Engine interface{
	Get(key string) ([]byte , error)
	Put(key string, value []byte) error
	Delete(key string) error
	Close() error
}