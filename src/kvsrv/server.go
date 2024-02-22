package kvsrv

import (
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
	kv map[string]string

	tmpData map[int64]string
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock()
	if tmp, ok := kv.tmpData[args.Id]; !ok {
		reply.Value = kv.kv[args.Key]
		kv.tmpData[args.Id] = reply.Value
	} else {
		reply.Value = tmp
	}
	kv.mu.Unlock()
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	if _, ok := kv.tmpData[args.Id]; !ok {
		kv.kv[args.Key] = args.Value
		kv.tmpData[args.Id] = args.Value
	}
	kv.mu.Unlock()
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	if tmp, ok := kv.tmpData[args.Id]; !ok {
		oldValue := kv.kv[args.Key]
		kv.kv[args.Key] = oldValue + args.Value
		reply.Value = oldValue
		kv.tmpData[args.Id] = oldValue
	} else {
		reply.Value = tmp
	}
	kv.mu.Unlock()
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	kv.kv = map[string]string{}
	kv.tmpData = map[int64]string{}

	return kv
}
