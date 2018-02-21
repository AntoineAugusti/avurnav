package avurnav

import (
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
)

type Storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) Storage {
	return Storage{redis: client}
}

func (s *Storage) Set(a AVURNAV) error {
	pipe := s.redis.Pipeline()
	pipe.Set(s.key(a), a, 0).Err()
	pipe.SAdd(a.PreMarRegion, s.key(a))
	_, err := pipe.Exec()
	return err
}

func (s *Storage) Get(a AVURNAV) (AVURNAV, error) {
	err := s.redis.Get(s.key(a)).Scan(&a)
	return a, err
}

func (s *Storage) key(a AVURNAV) string {
	return fmt.Sprintf("%s:%s", a.PreMarRegion, strconv.Itoa(a.ID))
}
