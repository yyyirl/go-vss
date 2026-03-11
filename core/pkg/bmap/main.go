package bmap

import (
	"bytes"
	"encoding/base64"
	"sync"
	"time"
)

type (
	Item struct {
		Data      *bytes.Buffer
		CreatedAt int64
		Bytes,
		// 转换后的数据
		G711AEncodeBytes []byte
	}

	BufferManager struct {
		items map[string]*Item
		mu    sync.RWMutex
	}
)

func NewBufferManager() *BufferManager {
	return &BufferManager{
		items: make(map[string]*Item),
	}
}

func (bm *BufferManager) Add(key, base64Data string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	item, exists := bm.items[key]
	if !exists {
		item = &Item{
			Data:      &bytes.Buffer{},
			Bytes:     nil,
			CreatedAt: time.Now().UnixMilli(),
		}
		bm.items[key] = item
	}

	// 清空缓存字节，下次获取时重新生成
	item.Bytes = nil
	_, err = item.Data.Write(data)
	return err
}

func (bm *BufferManager) Set(key string, data *Item) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.items[key] = data
}

func (bm *BufferManager) Get(key string) *Item {
	bm.mu.RLock()
	item, exists := bm.items[key]
	bm.mu.RUnlock()

	if !exists || item.Data.Len() == 0 {
		return nil
	}

	// 懒加载 只在第一次请求时生成Bytes副本
	if item.Bytes == nil {
		data := item.Data.Bytes()
		item.Bytes = make([]byte, len(data))
		copy(item.Bytes, data)
	}
	return item
}

func (bm *BufferManager) GetBufferSize(key string) int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	item, exists := bm.items[key]
	if !exists {
		return 0
	}

	return item.Data.Len()
}

func (bm *BufferManager) Remove(key string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	delete(bm.items, key)
}

func (bm *BufferManager) Reset(key string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	item, exists := bm.items[key]
	if exists {
		item.Data.Reset()
		item.Bytes = nil
	}
}

func (bm *BufferManager) All() []string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var keys = make([]string, 0, len(bm.items))
	for key := range bm.items {
		keys = append(keys, key)
	}
	return keys
}

func (bm *BufferManager) Len() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return len(bm.items)
}

func (bm *BufferManager) Size() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var totalBytes int
	for _, item := range bm.items {
		totalBytes += item.Data.Len()
	}

	return totalBytes
}

func (bm *BufferManager) Range(callback func(key string, item *Item)) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	for key, item := range bm.items {
		callback(key, item)
	}
}

// 清空所有过期的缓冲区（超过指定毫秒数）
func (bm *BufferManager) Cleanup(maxAgeMillis int64) int {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	var (
		now     = time.Now().UnixMilli()
		removed = 0
	)
	for key, item := range bm.items {
		if now-item.CreatedAt > maxAgeMillis {
			delete(bm.items, key)
			removed++
		}
	}
	return removed
}

func (bm *BufferManager) Exists(key string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	_, exists := bm.items[key]
	return exists
}
