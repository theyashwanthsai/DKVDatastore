package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

// state machine
// three operations to implement. Apply, Restore, Snapshot
// Apply - Keep all the nodes up-to-date
// Restore - Read all logs and apply them to fsm. Reset all local state
// Snapshot
type Fsm struct{
	db *sync.Map
}

func NewFsm(db *sync.Map) *Fsm {
    return &Fsm{
        db,
    }
}

type setPayload struct{
	Key string
	Value string
}

func (kf *Fsm) Apply(log *raft.Log)any{
	switch log.Type{
	case raft.LogCommand:
		var sp setPayload
		err := json.Unmarshal(log.Data, &sp)
		if err != nil{
			return fmt.Errorf("Could not parse: %s", err)
		}
		kf.db.Store(sp.Key, sp.Value)

	default:
		return fmt.Errorf("Unkown raft log type: %#v", log.Type)
	}
	return nil
}

func (kf *Fsm) Restore(rc io.ReadCloser) error{
	kf.db.Range(func(key any, _ any) bool{
		kf.db.Delete(key)
		return true
	})
	
	decoder := json.NewDecoder(rc)

	for decoder.More(){
		var sp setPayload
		err := decoder.Decode(&sp)
		if err != nil{
			return fmt.Errorf("Could not decode payload: %s", err)
		}
		kf.db.Store(sp.Key, sp.Value)
	}
	return rc.Close()
}

type snapshotNoop struct{

}

func (sn snapshotNoop) Persist(_ raft.SnapshotSink) error {
	return nil
}

func (sn snapshotNoop) Release(){
}

func (kf *Fsm) Snapshot() (raft.FSMSnapshot, error){
	return snapshotNoop{}, nil
}

func SetupRaft(dir, nodeId, raftAddress string, kf *Fsm) (*raft.Raft, error){
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil{
		return nil, fmt.Errorf("Could not create data directory: %s", err)
	}

	store, err := raftboltdb.NewBoltStore(path.Join(dir, "bolt"))
	if err != nil{
		return nil, fmt.Errorf("Could not create bolt store: %s", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(path.Join(dir, "snapshot"), 2, os.Stderr)
	if err != nil{
		return nil, fmt.Errorf("Could not create snapshot store: %s", err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not resolve address: %s", err)
    }

    transport, err := raft.NewTCPTransport(raftAddress, tcpAddr, 10, time.Second*10, os.Stderr)
    if err != nil {
        return nil, fmt.Errorf("Could not create tcp transport: %s", err)
    }

    raftCfg := raft.DefaultConfig()
    raftCfg.LocalID = raft.ServerID(nodeId)

    r, err := raft.NewRaft(raftCfg, kf, store, store, snapshots, transport)
    if err != nil {
        return nil, fmt.Errorf("Could not create raft instance: %s", err)
    }

    r.BootstrapCluster(raft.Configuration{
        Servers: []raft.Server{
            {
                ID:      raft.ServerID(nodeId),
                Address: transport.LocalAddr(),
            },
        },
    })

    return r, nil
}