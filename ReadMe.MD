You see this:

```go
// Start workers
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go worker(tasks, results, &wg)
}
```

And then **after that**, you send the tasks in this loop:

```go
for scanner.Scan() {
    lines = append(lines, scanner.Text())
    if len(lines) == chunkSize {
        tasks <- Task{lines: lines}
        fmt.Println("Processing lines here:", lines)
        lines = nil
    }
}
```

And you wonder:
**How is the worker running *before* all lines are read?**

---

## 🔑 The Answer: **Goroutines run in parallel!**

In Go, when you write:

```go
go worker(...)
```

It **starts the `worker` function in the background**, immediately — while your main function **keeps going**.

So even though you're still reading lines in `scanner.Scan()`, the worker is **already waiting on the channel**:

```go
for task := range tasks {
    // This will block until it gets a task
}
```

So the moment you send this:

```go
tasks <- Task{lines: lines}
```

💥 The worker immediately **receives the task**, starts processing, and prints the line.

---

### 🧠 Think of It Like This:

1. You **start a waiter (worker)** who’s standing at the window, waiting for food.
2. You **prepare and send one dish (line)**.
3. The waiter **instantly grabs it** and starts serving.
4. Meanwhile, you're still in the kitchen preparing the next line.

They’re **both working at the same time**!

---

## ✅ Timeline Example with `chunkSize = 1`

| Time | Main Function                      | Worker Goroutine           |
| ---- | ---------------------------------- | -------------------------- |
| t0   | Start worker                       | Worker waits for task      |
| t1   | Read line 1                        |                            |
| t2   | Send line 1 → `tasks <- Task{...}` | Receives task, starts work |
| t3   | Read line 2                        | Processing line 1          |
| t4   | Send line 2                        | Receives line 2            |

They **interleave** naturally — that’s how concurrency works in Go.

---

## 👀 How to See This in Action

Try adding print statements:

```go
// In worker
fmt.Println("Worker started")
fmt.Println("Worker received task:", task.lines)

// In main (producer)
fmt.Println("Sending task:", lines)
```

You’ll see the **worker prints and producer prints may interleave**, like:

```
Sending task: [line 1]
Worker received task: [line 1]
Sending task: [line 2]
Worker received task: [line 2]
```

---

## ✅ Summary

* `go worker(...)` **starts workers immediately** in the background.
* `scanner.Scan()` runs **in parallel** with the workers.
* As soon as you send a task to the channel, a worker **receives it and processes it**.
* You don’t need to wait until all lines are read — that’s the power of **concurrency in Go**.

---

