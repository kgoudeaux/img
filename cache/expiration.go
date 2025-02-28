package cache

// Expirations implements heap.Interface so that finding the entry with the oldest
// expiration is constant time
type Expirations []*Entry

func (e Expirations) Len() int           { return len(e) }
func (e Expirations) Less(i, j int) bool { return e[i].Expiration.Before(e[j].Expiration) }
func (e Expirations) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
	e[i].Index = i
	e[j].Index = j
}
func (e *Expirations) Push(x any) {
	n := len(*e)
	entry := x.(*Entry)
	entry.Index = n
	*e = append(*e, entry)
}
func (e *Expirations) Pop() any {
	old := *e
	n := len(old)
	entry := old[n-1]
	entry.Index = -1
	old[n-1] = nil

	*e = old[0 : n-1]
	return entry
}
