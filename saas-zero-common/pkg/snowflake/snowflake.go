package snowflake

import (
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	epoch        = 1700000000000
	workerBits   = 10
	sequenceBits = 12
	workerMax    = -1 ^ (-1 << workerBits)
	sequenceMask = -1 ^ (-1 << sequenceBits)
	workerShift  = sequenceBits
	timeShift    = workerBits + sequenceBits
)

var defaultNode *Node

func init() {
	workerID := int64(0)
	if v := os.Getenv("SNOWFLAKE_WORKER_ID"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			workerID = id
		}
	}
	if workerID < 0 || workerID > workerMax {
		workerID = 0
	}
	defaultNode = NewNode(workerID)
}

type Node struct {
	mu       sync.Mutex
	lastMS   int64
	workerID int64
	sequence int64
}

func NewNode(workerID int64) *Node {
	return &Node{workerID: workerID}
}

func NextID() int64 {
	return defaultNode.NextID()
}

func (n *Node) NextID() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()
	if now == n.lastMS {
		n.sequence = (n.sequence + 1) & sequenceMask
		if n.sequence == 0 {
			for now <= n.lastMS {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.sequence = 0
	}
	n.lastMS = now
	return (now-epoch)<<timeShift | n.workerID<<workerShift | n.sequence
}
