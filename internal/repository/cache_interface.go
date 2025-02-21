package repositories

type TaskCacheInterface interface {
	Set(key string, value any)
	Get(key string) (any, error)
	Delete(key string) error
}
