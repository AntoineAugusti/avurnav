package avurnav

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

type Storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) Storage {
	return Storage{redis: client}
}

func (s *Storage) AVURNAVsForRegion(region string) AVURNAVs {
	avurnavs := make(AVURNAVs, 0)

	ids := s.redis.SMembers(region).Val()
	if len(ids) == 0 {
		return avurnavs
	}

	values := s.redis.MGet(ids...).Val()
	for _, value := range values {
		var item AVURNAV
		if str, ok := value.(string); ok {
			json.Unmarshal([]byte(str), &item)
		}
		avurnavs = append(avurnavs, item)
	}

	return avurnavs
}

func (s *Storage) RegisterAVURNAVs(avurnavs AVURNAVs) error {
	pipe := s.redis.Pipeline()

	pipe.Del(s.region(avurnavs[0]))

	for _, avurnav := range avurnavs {
		pipe.Set(s.key(avurnav), avurnav, 0).Err()
		pipe.SAdd(s.region(avurnav), s.key(avurnav))
	}

	_, err := pipe.Exec()
	return err
}

func (s *Storage) Get(a AVURNAV) (AVURNAV, error) {
	err := s.redis.Get(s.key(a)).Scan(&a)
	return a, err
}

func (s *Storage) region(a AVURNAV) string {
	return strings.ToLower(a.PreMarRegion)
}

func (s *Storage) key(a AVURNAV) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(a.PreMarRegion), strconv.Itoa(a.ID))
}
