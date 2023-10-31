package zstore

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type (
	Store interface {
		Get(ctx context.Context, key string) (value []byte, err error)
		Set(ctx context.Context, key string, value []byte, exp time.Duration) (err error)
		Del(ctx context.Context, key string) (err error)
	}

	LoadFn = func() (v interface{}, exp time.Duration, err error)

	CacheLoader struct {
		m         sync.Map
		f         singleflight.Group
		Marshal   func(interface{}) ([]byte, error)
		Unmarshal func([]byte, interface{}) error
	}
)

var DefaultCacheLoader = CacheLoader{
	Marshal:   json.Marshal,
	Unmarshal: json.Unmarshal,
}

func WithCache(ctx context.Context, key string, fn LoadFn, cache Store, dst interface{}) (err error) {
	return DefaultCacheLoader.Load(ctx, key, fn, cache, dst)
}

func (l *CacheLoader) Load(ctx context.Context, key string, fn LoadFn, cache Store, dst interface{}) (err error) {
	if value, ok := l.m.Load(key); ok {
		return l.Unmarshal(value.([]byte), dst)
	}

	var caching chan error

	retChan := l.f.DoChan(key, func() (_ interface{}, err error) {
		v, _ := l.m.Load(key)
		data, ok := v.([]byte)
		if ok {
			return data, nil
		} else if data, err = cache.Get(ctx, key); err == nil && len(data) > 0 {
			return data, nil
		}
		v, exp, err := fn()
		if err != nil {
			return nil, err
		} else if data, err = l.Marshal(v); err != nil {
			return nil, err
		}
		l.m.Store(key, data)
		caching = make(chan error, 1)
		go func() { defer l.m.Delete(key); caching <- cache.Set(ctx, key, data, exp); close(caching) }()
		return data, nil
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ret := <-retChan:
		if ret.Err != nil {
			return ret.Err
		}
		if caching != nil {
			select {
			case <-ctx.Done():
			case <-caching:
			}
		}
		return l.Unmarshal(ret.Val.([]byte), dst)
	}
}
