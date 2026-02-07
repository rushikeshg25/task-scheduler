package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rushikeshg25/task-scheduler/internal/scheduler"
)

var (
	nord3  = lipgloss.Color("#4C566A")
	nord4  = lipgloss.Color("#D8DEE9")
	nord6  = lipgloss.Color("#ECEFF4")
	nord7  = lipgloss.Color("#8FBCBB")
	nord8  = lipgloss.Color("#88C0D0")
	nord9  = lipgloss.Color("#81A1C1")
	nord10 = lipgloss.Color("#5E81AC")
	nord11 = lipgloss.Color("#BF616A")
	nord13 = lipgloss.Color("#EBCB8B")
	nord14 = lipgloss.Color("#A3BE8C")
	nord15 = lipgloss.Color("#B48EAD")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(nord6).
			Background(nord10).
			Padding(0, 2).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(nord3).
			Padding(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().Foreground(nord9).Bold(true)
	valueStyle = lipgloss.NewStyle().Foreground(nord4)

	successStyle = lipgloss.NewStyle().Foreground(nord14).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(nord11).Bold(true)
	pendingStyle = lipgloss.NewStyle().Foreground(nord13).Bold(true)

	logHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(nord7).
			MarginTop(1)
)

type tickMsg time.Time

type model struct {
	scheduler *scheduler.Scheduler
	stats     scheduler.SchedulerStats
	logs      []string
	viewport  viewport.Model
	spinner   spinner.Model
	ready     bool
	width     int
	height    int
}

func initialModel(s *scheduler.Scheduler) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(nord8)

	return model{
		scheduler: s,
		spinner:   sp,
		logs:      []string{"System ready. Awaiting tasks..."},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 12
		footerHeight := 2
		vMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-vMarginHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - vMarginHeight
		}

	case tickMsg:
		m.stats = m.scheduler.GetStats()
		cmds = append(cmds, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}))

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case string:
		m.logs = append(m.logs, m.formatLog(msg))
		if len(m.logs) > 500 {
			m.logs = m.logs[1:]
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Booting scheduler..."
	}

	header := headerStyle.Render("TASK SCHEDULER DASHBOARD") + m.spinner.View() + " System " + successStyle.Render("ACTIVE")

	globalStats := cardStyle.Width(30).Height(8).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("GLOBAL STATUS"),
			"",
			labelStyle.Render("Total Tasks:  ")+valueStyle.Render(fmt.Sprintf("%d", int(m.stats.TasksCompleted+m.stats.TasksFailed)+m.stats.QueueSize)),
			labelStyle.Render("Completed:    ")+successStyle.Render(fmt.Sprintf("%d", m.stats.TasksCompleted)),
			labelStyle.Render("Failed:       ")+errorStyle.Render(fmt.Sprintf("%d", m.stats.TasksFailed)),
		),
	)

	queueStats := cardStyle.Width(30).Height(8).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("QUEUE MONITOR"),
			"",
			labelStyle.Render("Queue Depth:  ")+valueStyle.Render(fmt.Sprintf("%d", m.stats.QueueSize)),
			labelStyle.Render("Wait Status:  ")+pendingStyle.Render(m.getQueueStatus()),
		),
	)

	workerStats := cardStyle.Width(40).Height(8).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			labelStyle.Render("WORKER CLUSTER"),
			"",
			m.renderWorkers(),
		),
	)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, globalStats, queueStats, workerStats)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		topRow,
		logHeaderStyle.Render("SYSTEM ACTIVITY LOGS"),
		boxStyle.BorderForeground(nord15).Render(m.viewport.View()),
		"\n "+lipgloss.NewStyle().Foreground(nord3).Render("press q to exit • nord theme • solid architecture"),
	)
}

func (m model) getQueueStatus() string {
	if m.stats.QueueSize > 10 {
		return "HIGH LOAD"
	} else if m.stats.QueueSize > 0 {
		return "ACTIVE"
	}
	return "IDLE"
}

func (m model) renderWorkers() string {
	var b strings.Builder
	for i, w := range m.stats.Workers {
		if i >= 4 {
			break
		}
		indicator := lipgloss.NewStyle().Foreground(nord3).Render("●")
		statusText := valueStyle.Render("Idle   ")
		if w.Status == "Executing" {
			indicator = m.spinner.View()
			id := w.CurrentTaskID
			if len(id) > 10 {
				id = id[:7] + "..."
			}
			statusText = lipgloss.NewStyle().Foreground(nord8).Render(fmt.Sprintf("Run [%-10s]", id))
		}
		fmt.Fprintf(&b, "%s W%d: %s\n", indicator, w.ID, statusText)
	}
	return b.String()
}
func (m model) formatLog(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}

	parts := strings.SplitN(msg, "] ", 2)
	if len(parts) < 2 {
		return msg
	}

	timestamp := lipgloss.NewStyle().Foreground(nord3).Render(parts[0] + "]")
	content := parts[1]

	if strings.Contains(content, "completed") {
		content = strings.Replace(content, "completed", successStyle.Render("completed"), 1)
	} else if strings.Contains(content, "failed") {
		content = errorStyle.Render(content)
	} else if strings.Contains(content, "Retrying") {
		content = pendingStyle.Render(content)
	} else if strings.Contains(content, "Rescheduled") {
		content = lipgloss.NewStyle().Foreground(nord15).Render(content)
	}

	if strings.HasPrefix(content, "Worker") {
		workerParts := strings.SplitN(content, ":", 2)
		if len(workerParts) == 2 {
			content = lipgloss.NewStyle().Foreground(nord8).Bold(true).Render(workerParts[0]+":") + workerParts[1]
		}
	}

	return timestamp + " " + content
}
