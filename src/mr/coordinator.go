package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	TaskStateNoAssigned = 0
	TaskStateAssigned   = 1
	TaskStateFinished   = 2
)

type Coordinator struct {
	// Your definitions here.
	mut        sync.Mutex
	inputFiles []string
	nMap       int
	nReduce    int
	// TODO
	// This should be saved to a file
	mapResults []KeyValue
	// TODO
	// This too!
	keyValues   [][]KeyValues
	mTaskStates []int
	rTaskStates []int
}

func newCoordinator(files []string, nReduce int) Coordinator {
	nMap := len(files)

	mTaskStates := make([]int, nMap)
	for i := 0; i < nMap; i++ {
		mTaskStates[i] = TaskStateNoAssigned
	}

	rTaskStates := make([]int, nReduce)
	for i := 0; i < nReduce; i++ {
		rTaskStates[i] = TaskStateNoAssigned
	}

	return Coordinator{
		nMap:        nMap,
		nReduce:     nReduce,
		inputFiles:  files,
		mTaskStates: mTaskStates,
		rTaskStates: rTaskStates,
	}
}

func (c *Coordinator) assignMapTask(iMap int) RequestTaskReply {
	reply := RequestTaskReply{}
	filename := c.inputFiles[iMap]
	content, err := os.ReadFile(filename)
	if err != nil {
		reply.Ty = RequestTaskError
	} else {
		c.mTaskStates[iMap] = TaskStateAssigned
		reply.Ty = RequestMapTask
		reply.MapTaskInput = MapTaskInput{
			Id:       iMap,
			FileName: filename,
			Content:  string(content),
		}
	}
	return reply
}

type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func (c *Coordinator) mergeValues() {
	output := make([][]KeyValues, c.nReduce)
	sort.Sort(ByKey(c.mapResults))

	count := 0
	for i := 0; i < len(c.mapResults); {
		j := i + 1
		for j < len(c.mapResults) && c.mapResults[j].Key == c.mapResults[i].Key {
			j++
		}

		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, c.mapResults[k].Value)
		}
		output[count%c.nReduce] =
			append(output[count%c.nReduce],
				KeyValues{Key: c.mapResults[i].Key, Values: values})
		i = j
		count += 1
	}

	c.keyValues = output
}

func (c *Coordinator) assignReduceTask(iReduce int) RequestTaskReply {
	keyValues := c.keyValues[iReduce]
	c.rTaskStates[iReduce] = TaskStateAssigned
	return RequestTaskReply{
		Ty: RequestReduceTask,
		ReduceTaskInput: ReduceTaskInput{
			Id:     iReduce,
			Inputs: keyValues,
		},
	}
}

func (c *Coordinator) allMapTasksFinished() bool {
	for _, mt := range c.mTaskStates {
		if mt != TaskStateFinished {
			return false
		}
	}
	return true
}

func (c *Coordinator) allReduceTasksFinished() bool {
	for _, rt := range c.rTaskStates {
		if rt != TaskStateFinished {
			return false
		}
	}
	return true
}

func (c *Coordinator) checkCrashMapTask(iMap int) {
	time.Sleep(time.Second * 5)
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.mTaskStates[iMap] == TaskStateAssigned {
		c.mTaskStates[iMap] = TaskStateNoAssigned
	}
}

func (c *Coordinator) checkCrashReduceTask(iReduce int) {
	time.Sleep(time.Second * 5)
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.rTaskStates[iReduce] == TaskStateAssigned {
		c.rTaskStates[iReduce] = TaskStateNoAssigned
	}
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) RequestTask(args *RequestTaskArg, reply *RequestTaskReply) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	// assign any unassigned MapTasks
	for i, mt := range c.mTaskStates {
		if mt == TaskStateNoAssigned {
			*reply = c.assignMapTask(i)
			go c.checkCrashMapTask(i)
			return nil
		}
	}

	// wait until all MapTasks are completed
	if !c.allMapTasksFinished() {
		reply.Ty = RequestWait
		return nil
	}

	// assign any unassigned ReduceTasks
	for i, rt := range c.rTaskStates {
		if rt == TaskStateNoAssigned {
			*reply = c.assignReduceTask(i)
			go c.checkCrashReduceTask(i)
			return nil
		}
	}

	// wait until all ReduceTasks are completed
	if !c.allReduceTasksFinished() {
		reply.Ty = RequestWait
		return nil
	}

	reply.Ty = RequestTaskNoMore
	return nil
}

func (c *Coordinator) FinishMapTask(args *FinishMapTaskArgs, reply *FinishTaskReply) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.mapResults = append(c.mapResults, args.KeyValues...)
	c.mTaskStates[args.Id] = TaskStateFinished
	if c.allMapTasksFinished() {
		c.mergeValues()
	}
	return nil
}

func (c *Coordinator) FinishReduceTask(args *FinishReduceTaskArgs, reply *FinishTaskReply) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	os.WriteFile(fmt.Sprintf("mr-out-%d", args.Id), []byte(args.OutContent), 0664)
	c.rTaskStates[args.Id] = TaskStateFinished
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mut.Lock()
	defer c.mut.Unlock()
	for _, m := range c.mTaskStates {
		if m != TaskStateFinished {
			return false
		}
	}

	for _, m := range c.rTaskStates {
		if m != TaskStateFinished {
			return false
		}
	}

	return true
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := newCoordinator(files, nReduce)

	// Your code here.
	c.server()
	return &c
}
