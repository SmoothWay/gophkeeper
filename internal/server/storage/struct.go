package storage

type Item struct {
	UserID    int64
	Kind      string
	Key       string
	Data      []byte
	CreatedAt int64
}
