package main

import (
	"fmt"
	"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

type Task struct {
	id         int
	createdAt  string // время создания
	finishedAt string // время выполнения
	result     string
	isDone     bool
}

const (
	defaultTasksCount = 10
	interval          = -20 * time.Second
)

func main() {
	numTasks := defaultTasksCount
	superChan := make(chan Task, numTasks)
	doneTasks := make(chan Task)
	undoneTasks := make(chan Task)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go taskCreator(superChan, &wg, numTasks)

	go func() {
		defer close(doneTasks)
		defer close(undoneTasks)

		for t := range superChan {
			wg.Add(1)
			go func(task Task) {
				defer wg.Done()
				t = taskWorker(t)

				if t.isDone {
					doneTasks <- task
				} else {
					undoneTasks <- task
				}
			}(t)
		}

		wg.Wait()
	}()

	result := map[int]Task{}
	errors := map[int]Task{}
	mu := sync.Mutex{}

	go func() {
		for r := range doneTasks {
			mu.Lock()
			result[r.id] = r
			mu.Unlock()
		}
	}()

	go func() {
		for r := range undoneTasks {
			mu.Lock()
			errors[r.id] = r
			mu.Unlock()
		}
	}()

	time.Sleep(3 * time.Second)

	fmt.Println("Errors:")

	for _, err := range errors {
		fmt.Printf("id: %d, result: %s\n", err.id, err.result)
	}

	fmt.Println("Done tasks:")
	for _, res := range result {
		fmt.Printf("id: %d, result: %s\n", res.id, res.result)
	}
}

func taskCreator(tasks chan<- Task, wg *sync.WaitGroup, numTasks int) {
	defer wg.Done()
	for i := 0; i < numTasks; i++ {
		crunchTime := time.Now().Format(time.RFC3339)
		if time.Now().Nanosecond()%2 > 0 {
			crunchTime = "Some error occurred" // Я бы не стал так делать. Поле Time отвечает за время, а не за какое-то сообщение
		}
		tasks <- Task{createdAt: crunchTime, id: i + 1}
	}
}

func taskWorker(task Task) Task {
	createTime, _ := time.Parse(time.RFC3339, task.createdAt)

	task.isDone = createTime.After(time.Now().Add(interval))

	if task.isDone {
		task.result = "task has been success"
	} else {
		task.result = "something went wrong"
	}

	task.finishedAt = time.Now().Format(time.RFC3339Nano)
	time.Sleep(time.Millisecond * 150)

	return task
}
