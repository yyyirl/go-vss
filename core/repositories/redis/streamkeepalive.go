// @Title        streamkeepalive
// @Description  流保活状态
// @Create       yiyiyi 2025/7/14 10:32

package redis

import (
	"errors"

	r "github.com/go-redis/redis"
)

// 缓存键
const streamKeepAliveRunningStateKey string = cachePrefix + "keepalive:running-state"

type StreamKeepaliveRunningState struct {
	client *Client
}

func NewStreamKeepaliveRunningState(client *Client) *StreamKeepaliveRunningState {
	return &StreamKeepaliveRunningState{client}
}

func (s *StreamKeepaliveRunningState) Get(uniqueId string) (bool, error) {
	v, err := s.client.HGetInt64(streamKeepAliveRunningStateKey, uniqueId)
	if err != nil {
		if r.Nil == err {
			return false, nil
		}

		return false, err
	}

	return v == 1, nil
}

func (s *StreamKeepaliveRunningState) Clear() error {
	_, err := s.client.Del(streamKeepAliveRunningStateKey)
	return err
}

func (s *StreamKeepaliveRunningState) Del(uniqueIds []string) error {
	_, err := s.client.HDel(streamKeepAliveRunningStateKey, uniqueIds...)
	return err
}

func (s *StreamKeepaliveRunningState) Set(uniqueId string, v bool) (bool, error) {
	var val = 0
	if v {
		val = 1
	}

	return s.client.HSet(false, streamKeepAliveRunningStateKey, uniqueId, val)
}

func (s *StreamKeepaliveRunningState) Keys() ([]string, error) {
	return s.client.HKeys(streamKeepAliveRunningStateKey)
}

func (s *StreamKeepaliveRunningState) MGet(uniqueIds []string) (map[string]bool, error) {
	// 获取缓存
	res, err := s.client.HMGet(streamKeepAliveRunningStateKey, uniqueIds...)
	if err != nil {
		if r.Nil == err {
			return nil, nil
		}
		return nil, err
	}

	var maps = make(map[string]bool)
	for i, item := range res {
		if item == nil {
			continue
		}

		v, ok := item.(string)
		if ok {
			maps[uniqueIds[i]] = v == "1"
		} else {
			return nil, errors.New("invalid type")
		}

	}

	return maps, nil
}
