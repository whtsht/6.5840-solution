package kvsrv

type HasId interface {
	GetId() int64
	SetId(int64)
}

// Put or Append
type PutAppendArgs struct {
	Key   string
	Value string
	// You'll have to add definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Id int64
}

func (p *PutAppendArgs) GetId() int64 {
	return p.Id
}

func (p *PutAppendArgs) SetId(v int64) {
	p.Id = v
}

type PutAppendReply struct {
	Value string
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	Id int64
}

func (g *GetArgs) GetId() int64 {
	return g.Id
}

func (g *GetArgs) SetId(v int64) {
	g.Id = v
}

type GetReply struct {
	Value string
}
