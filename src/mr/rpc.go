package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Request Task
type RequestTaskArg struct{}

const (
	RequestMapTask    = 1
	RequestReduceTask = 2
	RequestWait       = 3
	RequestTaskNoMore = 4
	RequestTaskError  = 5
)

type RequestTaskReply struct {
	Ty              int
	MapTaskInput    MapTaskInput
	ReduceTaskInput ReduceTaskInput
}

type MapTaskInput struct {
	Id       int
	FileName string
	Content  string
}

type KeyValues struct {
	Key    string
	Values []string
}

type ReduceTaskInput struct {
	Id     int
	Inputs []KeyValues
}

// Finish Task
type FinishMapTaskArgs struct {
	Id        int
	KeyValues []KeyValue
}

type FinishReduceTaskArgs struct {
	Id         int
	OutContent string
}

type FinishTaskReply struct{}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
