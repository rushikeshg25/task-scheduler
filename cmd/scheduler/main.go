package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rushikeshg25/task-scheduler/internal/scheduler"
)

type logWriter struct {
	p *tea.Program
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf("[%s] %s", ts, string(p))
	w.p.Send(msg)
	return len(p), nil
}

func main() {
	log.SetFlags(0)
	s := scheduler.NewScheduler(4)
	s.Start()

	p := tea.NewProgram(initialModel(s), tea.WithAltScreen())

	log.SetOutput(&logWriter{p: p})

	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("boot-task-%d", i)
		s.Schedule(scheduler.NewBaseTask(id, 10, func(ctx context.Context) error {
			time.Sleep(time.Duration(i*200) * time.Millisecond)
			return nil
		}))
	}

	s.Schedule(scheduler.NewBaseTask("cleanup", 5, func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		return nil
	}).WithDelay(5 * time.Second))

	s.Schedule(scheduler.NewBaseTask("heartbeat", 1, func(ctx context.Context) error {
		return nil
	}).WithPeriodic(3 * time.Second))

	s.Schedule(scheduler.NewBaseTask("unstable-service", 1, func(ctx context.Context) error {
		time.Sleep(500 * time.Millisecond)
		return fmt.Errorf("connection reset")
	}).WithRetry(5))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running UI: %v", err)
		os.Exit(1)
	}
	s.Stop()
}
