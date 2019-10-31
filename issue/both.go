package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"

	// "github.com/RichardKnop/machinery/v1/log"
	"github.com/RichardKnop/machinery/v1/tasks"
)

func main() {

	cfg := &config.Config{
		Broker:          "amqp://guest:guest@localhost:5672/",
		ResultBackend:   "eager",
		ResultsExpireIn: 3600,
		DefaultQueue:    "machinery_tasks",
		AMQP: &config.AMQPConfig{
			Exchange:      "machinery_exchange",
			ExchangeType:  "direct",
			BindingKey:    "machinery_task",
			PrefetchCount: 3,
		},
	}
	taskServer, err := machinery.NewServer(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to get Machinery server: %v", err))
	}

	taskServer.RegisterTask("testTask", func(arg string) error {
		log.Printf("The value is: %s", arg)
		return nil
	})

	go func() {
		worker := taskServer.NewWorker("test", 1)

		worker.SetErrorHandler(func(err error) {
			log.Println("%v", err)
		})

		if err := worker.Launch(); err != nil {
			log.Fatalf("unrecoverable error from task worker: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)

	for index := 0; index < 10; index++ {
		sig, err := tasks.NewSignature("testTask", []tasks.Arg{
			{Type: "string", Value: fmt.Sprintf("task %d", index)},
		})

		if err != nil {
			log.Fatal(err)
		}
		_, err = taskServer.SendTaskWithContext(context.Background(), sig)
		if err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(30 * time.Second)
}
