package dq

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const (
	replicaNodes    = 3 // 默认副本节点数
	minWrittenNodes = 2 // 最少可写节点数
)

type (
	// 任务生产者
	Producer interface {
		At(body []byte, at time.Time) (string, error)           // 定时执行
		Delay(body []byte, delay time.Duration) (string, error) // 延迟执行
		Revoke(ids string) error                                // 撤回一批任务
		Close() error
	}

	// 生产者集群
	producerCluster struct {
		nodes []Producer
	}
)

func (pc producerCluster) At(body []byte, at time.Time) (string, error) {
	return pc.insert(func(node Producer) (string, error) {
		return node.At(pc.wrap(body, at), at)
	})
}

func (pc producerCluster) Delay(body []byte, delay time.Duration) (string, error) {
	panic("implement me")
}

func (pc producerCluster) Revoke(ids string) error {
	panic("implement me")
}

func (pc producerCluster) Close() error {
	panic("implement me")
}

func (pc *producerCluster) insert(fn func(node Producer) (string, error)) (string, error) {
	for _, node := range pc.getWriteNodes() {
		fmt.Println(node)
	}
	return "", nil
}

// wrap 将内容和执行时间包装为：UnixNano时间/内容
func (pc *producerCluster) wrap(body []byte, at time.Time) []byte {
	var b bytes.Buffer
	b.WriteString(strconv.FormatInt(at.UnixNano(), 10))
	b.WriteByte(timeSep)
	b.Write(body)
	return b.Bytes()
}

func (pc *producerCluster) getWriteNodes() []Producer {
	if len(pc.nodes) <= replicaNodes {
		return pc.nodes
	}

	// 实际节点数比预设的多，则按预设数量随机选择
	nodes := pc.cloneNodes()
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	return nodes[:replicaNodes]
}

func (pc *producerCluster) cloneNodes() []Producer {
	return append([]Producer(nil), pc.nodes...)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewProducer(beanstalks []Beanstalk) Producer {
	if len(beanstalks) < minWrittenNodes {
		log.Fatalf("节点数必须大于等于 %d", minWrittenNodes)
	}

	var nodes []Producer
	for _, node := range beanstalks {
		nodes = append(nodes, NewProducerNode(node.Endpoint, node.Tube))
	}
	return &producerCluster{nodes: nodes}
}
