package main

import (
	"runtime"
	"sync"
)

var (
	alpha     []int            //first line of input file
	noServers int              //number of given servers
	steps     int              //compilation steps
	targets   []string         //targets stores the names of target files
	files     map[string]*file //files maps file name to the file's pointer
	fchan     chan *file       //fchan sends file pointer to server routine
	token     chan struct{}    //token prevents concurrent access to steps and files
	wg        sync.WaitGroup   //wg allows waiting for the server routines
	send      sync.WaitGroup   //send allows waiting for sender routine
)

//The constant defining special Replication cases
//see specReplication in compile.go
const cond = float64(1.8)

//clear resets all variables
func clear() {
	alpha = make([]int, 0)
	noServers = 0
	steps = 0
	targets = make([]string, 0)
	files = make(map[string]*file)
	fchan = make(chan *file)
	token = make(chan struct{}, 1)
	go runtime.GC() //save time
}

//The file struct fully represents a file and all its provided properties
type file struct {
	sync.Mutex
	Name              string   //The Name of this Fiile
	CompileTime       int      //Time to compile file
	ReplicationTime   int      //Time to replicate file
	Deps              []string //File dependencies
	CompiledOnServers []int    //Servers this file has been compiled on
	IsCompiled        bool
	Replicated        chan struct{} //Channel to signal file replication
	pick              chan struct{}
}

//wasCompiledOn checks if f was compiled on the server whose id is id
func (f *file) wasCompiledOn(id int) bool {
	token <- struct{}{}
	defer func() { <-token }()
	for _, val := range f.CompiledOnServers {
		if val == id {
			return true
		}
	}
	return false
}
