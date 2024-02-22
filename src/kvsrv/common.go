package kvsrv

type HasId interface {
	SetId(int64)
	SetCount(int64)
}

// Put or Append
type PutAppendArgs struct {
	Key   string
	Value string
	// You'll have to add definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Id    int64
	Count int64
}

func (p *PutAppendArgs) SetId(v int64) {
	p.Id = v
}

func (p *PutAppendArgs) SetCount(v int64) {
	p.Count = v
}

type PutAppendReply struct {
	Value string
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	Id    int64
	Count int64
}

func (g *GetArgs) SetId(v int64) {
	g.Id = v
}

func (g *GetArgs) SetCount(v int64) {
	g.Count = v
}

type GetReply struct {
	Value string
}
