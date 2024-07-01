package cache

type Cache interface {
	Close() error
	Del(pattern string) error
	RPush(key string, values ...interface{}) error
	LRange(key string) ([]string, error)
}
