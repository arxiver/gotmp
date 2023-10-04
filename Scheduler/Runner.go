package Scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

func Init() {
	s := gocron.NewScheduler(time.UTC)
	s.StartAsync()
	fmt.Println("Initializing the scheduler ... [OK]")
}
