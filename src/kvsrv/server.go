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

	tmpData map[int64]struct {
		int64
		string
	}
}

func (kv *KVServer) getTmpData(id int64, count int64) (string, bool) {
	if tmp, ok := kv.tmpData[id]; ok && tmp.int64 == count {
		return tmp.string, true
	} else {
		return "", false
	}
}

func (kv *KVServer) setTmpData(id int64, count int64, value string) {
	if tmp, ok := kv.tmpData[id]; !ok || tmp.int64 < count {
		kv.tmpData[id] = struct {
			int64
			string
		}{count, value}
	}
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock()
	if tmp, ok := kv.getTmpData(args.Id, args.Count); !ok {
		reply.Value = kv.kv[args.Key]
		kv.setTmpData(args.Id, args.Count, reply.Value)
	} else {
		reply.Value = tmp
	}
	kv.mu.Unlock()
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	if _, ok := kv.getTmpData(args.Id, args.Count); !ok {
		kv.kv[args.Key] = args.Value
		kv.setTmpData(args.Id, args.Count, "")
	}
	kv.mu.Unlock()
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	if tmp, ok := kv.getTmpData(args.Id, args.Count); !ok {
		oldValue := kv.kv[args.Key]
		kv.kv[args.Key] = oldValue + args.Value
		reply.Value = oldValue
		kv.setTmpData(args.Id, args.Count, oldValue)
	} else {
		reply.Value = tmp
	}
	kv.mu.Unlock()
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	kv.kv = map[string]string{}
	kv.tmpData = map[int64]struct {
		int64
		string
	}{}

	return kv
}
