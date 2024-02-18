package mr

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	for work(mapf, reducef) {
		time.Sleep(time.Millisecond * 100)
	}
}

func runMapTask(mapf func(string, string) []KeyValue, keyValues chan []KeyValue, reply *RequestTaskReply) {
	keyValues <- mapf(reply.MapTaskInput.FileName, reply.MapTaskInput.Content)
}

func runReduceTask(reducef func(string, []string) string, input []KeyValues, output chan string) {
	tmp := ""
	for _, kv := range input {
		ret := reducef(kv.Key, kv.Values)
		tmp += kv.Key + " " + ret + "\n"
	}
	output <- tmp
}

func checkTimeout(quit chan int) {
	time.Sleep(time.Second * 5)
	quit <- 0
}

func work(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) bool {
	args := RequestTaskArg{}
	reply := RequestTaskReply{}
	ok := call("Coordinator.RequestTask", &args, &reply)
	if ok {
		switch reply.Ty {
		case RequestMapTask:
			keyValuesChan := make(chan []KeyValue)
			quit := make(chan int)
			go checkTimeout(quit)
			go runMapTask(mapf, keyValuesChan, &reply)
			select {
			case keyValues := <-keyValuesChan:
				args := FinishMapTaskArgs{
					Id:        reply.MapTaskInput.Id,
					KeyValues: keyValues,
				}
				reply := FinishTaskReply{}
				call("Coordinator.FinishMapTask", &args, &reply)
			case <-quit:
				os.Exit(1)
			}

		case RequestReduceTask:
			outputChan := make(chan string)
			quit := make(chan int)
			go checkTimeout(quit)
			go runReduceTask(reducef, reply.ReduceTaskInput.Inputs, outputChan)
			select {
			case output := <-outputChan:
				args := FinishReduceTaskArgs{
					Id:         reply.ReduceTaskInput.Id,
					OutContent: output,
				}
				reply := FinishTaskReply{}
				call("Coordinator.FinishReduceTask", &args, &reply)
			case <-quit:
				os.Exit(1)
			}

		case RequestWait:
			return true
		case RequestTaskNoMore:
			return false
		}
	}
	return true
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
