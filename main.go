package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	clear() //use clear to initalize variables
}

func main() {
	stt := time.Now()
	// create output directory
	err := os.Mkdir("outputs", os.ModePerm)
	if err != nil {
		fmt.Printf("Could not create outputs directory: %v\n", err)
		os.Exit(3)
	}
	// Read the `input` directory so that we don't have to
	// modify the code whenever we want to test other inputs
	filepath.Walk("inputs", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				// we dont want to stop the whole app if just one file does not open
				return err
			}
			defer file.Close()
			lines := extract(file)

			initVars(lines)
			go start()
			servers()
			printToFile(path)
			clear()
		}
		return nil
	})
	fmt.Println("Time", time.Since(stt))
}

//extract the file lines into a slice
func extract(f *os.File) *[]string {
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return &lines
}

//Take each element in the lines slice and extract into allFiles
func initVars(data *[]string) {
	lines := *data
	for _, val := range strings.Split(lines[0], " ") {
		tmp, _ := strconv.Atoi(val)
		alpha = append(alpha, tmp)
	}

	lastLine := len(lines) - 1
	for line := lastLine; line > lastLine-alpha[1]; line-- {
		tmp := strings.Split(lines[line], " ")[0]
		targets = append(targets, tmp)
	}

	noServers = alpha[2]
	for i := 1; i < len(lines)-alpha[1]; i = i + 2 {
		wg.Add(1)
		go func(i int) {
			tmp := strings.Split(lines[i], " ")
			struc := &file{}
			struc.Name = tmp[0]
			struc.CompileTime, _ = strconv.Atoi(tmp[1])
			struc.ReplicationTime, _ = strconv.Atoi(tmp[2])
			struc.Deps = strings.Split(lines[i+1], " ")[1:]
			if canReplicate(struc) {
				struc.Replicated = make(chan struct{}, 1)
				struc.pick = make(chan struct{}, 1)
			}
			token <- struct{}{} //acquire token
			files[struc.Name] = struc
			<-token //release token
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func start() {
	for _, target := range targets {
		traverse(files[target])
	}
	send.Wait()
	close(fchan)
}

func traverse(file *file) {
	for _, dep := range file.Deps {
		traverse(files[dep])
	}
	send.Add(1)
	//sender routine
	go func() {
		defer send.Done()
		fchan <- file
	}()
}

func servers() {
	for id := 0; id < noServers; id++ {
		wg.Add(1)
		//server routine
		go func(id int) {
			defer wg.Done()
			for file := range fchan {
				Compile(file, id)
			}
		}(id)
	}
	wg.Wait()
}

//Print needed output to file
func printToFile(path string) {
	outFile := path[7:8]
	outFile = "outputs/" + outFile + ".out"

	output := ""
	output = strconv.Itoa(steps) + "\n"
	for _, file := range files {
		for _, server := range file.CompiledOnServers {
			output += file.Name + " " + strconv.Itoa(server) + "\n"
		}
	}

	f, err := os.Create(outFile)
	defer f.Close()
	if err != nil {
		fmt.Println("Cannot create output file: ", err)
		return // return since there is no file to write to
	}

	_, err = f.WriteString(output)
	if err != nil {
		fmt.Println("Cannot write output to file: ", err)
	}
	f.Sync()
}
