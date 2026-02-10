package cluster

import (
	"io"
	"log"

	"github.com/hashicorp/memberlist"
)

type Cluster struct {
	List *memberlist.Memberlist
}

// Delegate sederhana untuk menangani event (optional untuk level lanjut)
type eventDelegate struct {
	onJoin  func(string)
	onLeave func(string)
}

func (ed *eventDelegate) NotifyJoin(node *memberlist.Node)   { ed.onJoin(node.Name) }
func (ed *eventDelegate) NotifyLeave(node *memberlist.Node)  { ed.onLeave(node.Name) }
func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {}

func NewCluster(nodeName string, port int, joinAddr string, onJoin, onLeave func(string)) (*Cluster, error) {
	config := memberlist.DefaultLocalConfig()
	config.Name = nodeName
	config.BindPort = port
	config.AdvertisePort = port

	// Agar log tidak terlalu berisik di terminal
	config.LogOutput = io.Discard

	// Setup delegate untuk memberitahu Hash Ring saat ada perubahan
	delegate := &eventDelegate{onJoin: onJoin, onLeave: onLeave}
	config.Events = delegate

	list, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}

	if joinAddr != "" {
		_, err := list.Join([]string{joinAddr})
		if err != nil {
			log.Printf("Gagal bergabung ke cluster: %v", err)
		}
	}

	return &Cluster{List: list}, nil
}
