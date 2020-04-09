package main

import (
	"time"
)

//specReplication checks if ratio of replication time to compile time
//is less than cond's value,if so it'll replicate the file,
//assuming that the ratio makes it convenient to replicate the file,
//i took this assumption from
//the example output in the pdf.
func canReplicate(f *file) bool {
	var ratio float64
	ratio = float64(f.ReplicationTime / f.CompileTime)

	if ratio > cond {
		return false
	}
	return true
}

//Compile simulates compilation
func Compile(f *file, serverID int) {
	if f.Replicated != nil { //possible reason for abnormal behaviour
		select {
		case f.pick <- struct{}{}:
			compileAndReplicate(f, serverID)
		case <-f.Replicated:
			return
		}
	}

	if !f.wasCompiledOn(serverID) {
		compileWithoutReplication(f, serverID)
	}
}

//compileAndReplicate compiles the files and then replicates it to all servers
func compileAndReplicate(f *file, serverID int) {
	time.Sleep((time.Duration((f.CompileTime + f.ReplicationTime) / 1000000000)) * time.Millisecond)
	close(f.Replicated) //signal that this file has been replicated
	f.CompiledOnServers = append(f.CompiledOnServers, serverID)
	addStep()
}

//compileWithoutReplication compiles the file, but doesn't replicate it
func compileWithoutReplication(f *file, serverID int) {
	time.Sleep((time.Duration(f.CompileTime / 1000000000)) * time.Millisecond)
	tok <- struct{}{}
	f.CompiledOnServers = append(f.CompiledOnServers, serverID)
	<-tok
	addStep()
}

//addStep allows for counting of compilation steps
func addStep() {
	token <- struct{}{}
	steps++
	<-token
}
