package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	chunkSize  = 5
	numWorkers = 5 // adjust based on your system's cores
)

type Task struct {
	lines []string
}

type Result struct {
	counts map[string]int
}

func worker(tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	fmt.Println("Worker started")
	defer wg.Done()
	for task := range tasks {
		localCount := make(map[string]int)
		for _, line := range task.lines {
			fmt.Println("Worker started")
			fmt.Println("Worker received task:", task.lines)
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 1 {
				continue
			}
			logLevel := strings.Trim(parts[0], "[]")
			localCount[logLevel]++
		}
		results <- Result{counts: localCount}
	}
}

func ReadFileWithWorkerPool(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("file open error:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	tasks := make(chan Task)
	results := make(chan Result)

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(tasks, results, &wg)
	}

	// Collector goroutine
	var collectorWg sync.WaitGroup
	container := make(map[string]int)
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for result := range results {
			for k, v := range result.counts {
				container[k] += v
			}
		}
	}()

	// Producer: scan lines and send tasks
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		fmt.Println("Sending task:", lines)
		if len(lines) == chunkSize {
			tasks <- Task{lines: lines}
			lines = nil
		}
	}
	if len(lines) > 0 {
		tasks <- Task{lines: lines}
	}
	close(tasks) // signal no more tasks

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Wait for collector to finish
	collectorWg.Wait()

	// Print results
	fmt.Println("Log Level Counts:")
	for k, v := range container {
		fmt.Printf("%s: %d\n", k, v)
	}
}

func main() {
	ReadFileWithWorkerPool("file.log")
}
