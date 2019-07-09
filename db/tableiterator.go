package db

/*
	tableIterator is a wrapper around the Iterator interface.
	It removes the
*/
type tableIterator struct {
	Iterator
	prefix int
}

func newTableIterator(it Iterator, table *table) Iterator {
	return &tableIterator{
		Iterator: it,
		prefix:   len(table.prefix),
	}
}

/*
	Key returns the value that is iterated over, but without the table's prefix.
*/
func (it *tableIterator) Key() string {
	return it.Iterator.Key()[it.prefix:]
}
