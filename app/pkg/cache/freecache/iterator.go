package freecache

import (
	"github.com/charopevez/eob-wayfinder-worker/pkg/cache"
	"github.com/coocood/freecache"
)

type iterator struct {
	iter *freecache.Iterator
}

func (i *iterator) Next() *cache.Entry {
	entry := i.iter.Next()
	if entry == nil {
		return nil
	}

	return &cache.Entry{
		Key:   entry.Key,
		Value: entry.Value,
	}
}
