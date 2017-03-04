package maze

import deadlock "github.com/sasha-s/go-deadlock"

type safeMap2 struct {
	deadlock.RWMutex
	data map[*Cell]interface{}
}

type safeMapItem struct {
	Key   *Cell
	Value interface{}
}

func NewSafeMap2() *safeMap2 {
	return &safeMap2{
		data: make(map[*Cell]interface{}),
	}
}

func (sm *safeMap2) Keys() []*Cell {
	sm.RLock()
	defer sm.RUnlock()

	var keys []*Cell
	for k := range sm.data {
		keys = append(keys, k)
	}
	return keys
}

func (sm *safeMap2) Iter() <-chan safeMapItem {
	c := make(chan safeMapItem, 10)

	f := func() {
		sm.RLock()
		defer sm.RUnlock()
		for k, v := range sm.data {
			c <- safeMapItem{k, v}
		}
		close(c)
	}
	go f()

	return c
}

func (sm *safeMap2) Insert(key *Cell, value interface{}) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[key] = value
}

func (sm *safeMap2) Delete(key *Cell) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.data, key)
}

func (sm *safeMap2) Find(key *Cell) (interface{}, bool) {
	sm.RLock()
	defer sm.RUnlock()
	v, ok := sm.data[key]
	return v, ok
}

func (sm *safeMap2) Len() int {
	sm.RLock()
	defer sm.RUnlock()
	return len(sm.data)
}

func (sm *safeMap2) Update(key *Cell, value interface{}) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[key] = value
}
