package maze

const (
	remove commandAction = iota
	end
	find
	insert
	length
	update
	keys // return a list of all keys
)

type commandAction int

type safeMap chan commandData

type SafeMap interface {
	Insert(*Cell, interface{})
	Delete(*Cell)
	Find(*Cell) (interface{}, bool)
	Len() int
	Update(*Cell, UpdateFunc)
	Close() map[*Cell]interface{}
	Keys() []*Cell
}

type UpdateFunc func(interface{}, bool) interface{}

type commandData struct {
	action commandAction
	key    *Cell
	value  interface{}
	result chan<- interface{}
	data   chan<- map[*Cell]interface{}
	update UpdateFunc
}

func (sm safeMap) Keys() []*Cell {
	reply := make(chan interface{})
	sm <- commandData{action: keys, result: reply}
	result := (<-reply).([]*Cell)
	return result
}

func (sm safeMap) Insert(key *Cell, value interface{}) {
	sm <- commandData{action: insert, key: key, value: value}
}

func (sm safeMap) Delete(key *Cell) {
	sm <- commandData{action: remove, key: key}
}

type findResult struct {
	value interface{}
	found bool
}

func (sm safeMap) Find(key *Cell) (interface{}, bool) {
	reply := make(chan interface{})
	sm <- commandData{action: find, key: key, result: reply}
	result := (<-reply).(findResult)
	return result.value, result.found
}

func (sm safeMap) Len() int {
	reply := make(chan interface{})
	sm <- commandData{action: length, result: reply}
	return (<-reply).(int)
}

func (sm safeMap) Update(key *Cell, updater UpdateFunc) {
	sm <- commandData{action: update, key: key, update: updater}
}

func (sm safeMap) Close() map[*Cell]interface{} {
	reply := make(chan map[*Cell]interface{})
	sm <- commandData{action: end, data: reply}
	return <-reply
}

func NewSafeMap() safeMap {
	g := make(safeMap)
	go g.run()
	return g
}

func (sm safeMap) run() {
	store := make(map[*Cell]interface{})

	for command := range sm {
		switch command.action {
		case insert:
			store[command.key] = command.value
		case remove:
			delete(store, command.key)
		case find:
			value, found := store[command.key]
			command.result <- findResult{value, found}
		case length:
			command.result <- len(store)
		case update:
			value, found := store[command.key]
			store[command.key] = command.update(value, found)
		case keys:
			var keys []*Cell
			for k := range store {
				keys = append(keys, k)
			}
			command.result <- keys
		case end:
			close(sm)
			command.data <- store
		}
	}
}
