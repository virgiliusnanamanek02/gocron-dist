package hash

import (
	"hash/crc32"
	"sort"
	"sync"
)

type Consistent struct {
	nodes   []uint32          // Menyimpan hash dari setiap node yang terurut
	nodeMap map[uint32]string // Map dari hash ke nama asli node (misal: "node-1")
	mu      sync.RWMutex
}

func NewConsistent() *Consistent {
	return &Consistent{
		nodeMap: make(map[uint32]string),
	}
}

// AddNode memasukkan node baru ke dalam ring
func (c *Consistent) AddNode(nodeName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := crc32.ChecksumIEEE([]byte(nodeName))
	c.nodes = append(c.nodes, hash)
	c.nodeMap[hash] = nodeName
	sort.Slice(c.nodes, func(i, j int) bool { return c.nodes[i] < c.nodes[j] })
}

// GetNode mencari node mana yang bertanggung jawab atas sebuah ID (Job ID)
func (c *Consistent) GetNode(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.nodes) == 0 {
		return ""
	}

	hash := crc32.ChecksumIEEE([]byte(key))

	// Mencari node pertama yang hash-nya >= hash key
	idx := sort.Search(len(c.nodes), func(i int) bool {
		return c.nodes[i] >= hash
	})

	// Jika tidak ketemu yang lebih besar, putar balik ke node pertama (circular)
	if idx == len(c.nodes) {
		idx = 0
	}

	return c.nodeMap[c.nodes[idx]]
}
