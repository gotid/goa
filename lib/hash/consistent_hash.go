package hash

import (
	"fmt"
	"github.com/z-sdk/goa/lib/lang"
	"github.com/z-sdk/goa/lib/mapping"
	"sort"
	"strconv"
	"sync"
)

const (
	minReplicas = 100
	TopWeight   = 100
	prime       = 16777619
)

// Generate Hash 生成器
type HashFunc func(data []byte) uint64

// 一致性hash结构
//
// ConsistentHash
//
// 	-> ring: {hash1: [node1, node2]}
//	-> keys: [hash1, hash2]
//	-> nodes: {'': {}}
type ConsistentHash struct {
	hashFunc HashFunc
	replicas int
	keys     []uint64                 // 以hash值为key, Sorted hash list
	ring     map[uint64][]interface{} // 以hash值为键，node列表为值的map
	nodes    map[string]lang.PlaceholderType
	lock     sync.RWMutex
}

func NewConsistentHash() *ConsistentHash {
	return NewCustomConsistentHash(minReplicas, Hash)
}

func NewCustomConsistentHash(replicas int, hashFunc HashFunc) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}

	if hashFunc == nil {
		hashFunc = Hash
	}

	return &ConsistentHash{
		hashFunc: hashFunc,
		replicas: replicas,
		ring:     make(map[uint64][]interface{}),
		nodes:    make(map[string]lang.PlaceholderType),
	}
}

func (h *ConsistentHash) Add(node interface{}) {
	h.AddWithReplicas(node, h.replicas)
}

func (h *ConsistentHash) AddWithWeight(node interface{}, weight int) {
	replicas := h.replicas * weight / TopWeight
	h.AddWithReplicas(node, replicas)
}

// AddWithReplicas 向指定节点增加指定数量的副本。
// 如果副本数量比 h.replicas 多将被截断，后添加会覆盖之前添加的。
func (h *ConsistentHash) AddWithReplicas(node interface{}, replicas int) {
	h.Remove(node)

	if replicas > h.replicas {
		replicas = h.replicas
	}

	nodeRepr := repr(node)
	h.lock.Lock()
	defer h.lock.Unlock()

	h.addNode(nodeRepr)

	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(nodeRepr + strconv.Itoa(i)))
		h.keys = append(h.keys, hash)
		h.ring[hash] = append(h.ring[hash], node)
	}

	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

// Get 获取指定 v 对应的 node 节点
func (h *ConsistentHash) Get(v interface{}) (interface{}, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.ring) == 0 {
		return nil, false
	}

	hash := h.hashFunc([]byte(repr(v)))
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	}) % len(h.keys)

	nodes := h.ring[h.keys[index]]
	switch len(nodes) {
	case 0:
		return nil, false
	case 1:
		return nodes[0], true
	default:
		innerIndex := h.hashFunc([]byte(innerRepr(v)))
		pos := int(innerIndex % uint64(len(nodes)))
		return nodes[pos], true
	}
}

func (h *ConsistentHash) Remove(node interface{}) {
	nodeRepr := repr(node)

	h.lock.Lock()
	defer h.lock.Unlock()

	if !h.contains(nodeRepr) {
		return
	}

	for i := 0; i < h.replicas; i++ {
		hash := h.hashFunc([]byte(nodeRepr + strconv.Itoa(i)))
		index := sort.Search(len(h.keys), func(i int) bool {
			return h.keys[i] >= hash
		})
		if index < len(h.keys) {
			h.keys = append(h.keys[:index], h.keys[index+1:]...)
		}
		h.removeRingNode(hash, nodeRepr)
	}

	h.removeNode(nodeRepr)
}

// 删除 ring 中 该 hash 对应的 node
func (h *ConsistentHash) removeRingNode(hash uint64, nodeRepr string) {
	if nodes, ok := h.ring[hash]; ok {
		newNodes := nodes[:0]
		for _, x := range nodes {
			if repr(x) != nodeRepr {
				newNodes = append(newNodes, x)
			}
		}
		if len(newNodes) > 0 {
			h.ring[hash] = newNodes
		} else {
			// 若是节点列表中的最后一个节点，则删除该映射对
			delete(h.ring, hash)
		}
	}
}

func (h *ConsistentHash) removeNode(nodeRepr string) {
	delete(h.nodes, nodeRepr)
}

func (h *ConsistentHash) addNode(nodeRepr string) {
	h.nodes[nodeRepr] = lang.Placeholder
}

func (h *ConsistentHash) contains(nodeRepr string) bool {
	_, ok := h.nodes[nodeRepr]
	return ok
}

func repr(node interface{}) string {
	return mapping.Repr(node)
}

func innerRepr(node interface{}) string {
	return fmt.Sprintf("%d:%v", prime, node)
}
