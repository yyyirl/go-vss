/**
 * @Author:         yi
 * @Description:    queue
 * @Version:        1.0.0
 * @Date:           2024/12/25 10:19
 */
package redis

import "errors"

const (
	// 流保活
	QueueStreamPubListen = "stream-keepalive"
)

type Queue struct {
	client *Client
}

func NewQueue(client *Client) *Queue {
	return &Queue{client}
}

func (q *Queue) cacheKey(table string) string {
	return cachePrefix + "queue:" + table
}

func (q *Queue) Set(filed string, data interface{}) error {
	_, err := q.client.RPush(q.cacheKey(filed), data)
	return err
}

func (q *Queue) Clear(filed string) error {
	_, err := q.client.Del(q.cacheKey(filed))
	return err
}

func (q *Queue) Get(filed string, limit int, callback func(data [][]byte)) error {
	if limit <= 0 {
		return errors.New("limit 不能小于0")
	}

	// key
	cache, err := q.client.LRange(q.cacheKey(filed), 0, limit-1)
	if err != nil {
		if err == RedisNil {
			return nil
		}

		return err
	}

	if len(cache) <= 0 {
		return nil
	}

	if len(cache) > 0 {
		var records [][]byte
		for _, item := range cache {
			records = append(records, []byte(item))
		}
		go callback(records)

		// 保留limit到最后的元素
		_, err = q.client.LTrim(q.cacheKey(filed), limit, -1)
	}

	return err
}
