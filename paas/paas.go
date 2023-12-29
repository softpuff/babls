package paas

// A simple example that shows how to send messages to a Bubble Tea program
// from outside the program using Program.Send(Msg).

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle      = helpStyle.Copy().UnsetMargins()
	durationStyle = dotStyle.Copy()
	appStyle      = lipgloss.NewStyle().Margin(1, 2, 0, 2)
)

type resultMsg struct {
	duration  time.Duration
	food      string
	container string
}

func (r resultMsg) String() string {
	if r.duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("🧗‍♀️ %s %s", r.food, durationStyle.Render(r.container))
	// durationStyle.Render(r.duration.String()))
}

type model struct {
	bloccSpinner spinner.Model
	paasSpinner  spinner.Model
	results      []resultMsg
	quitting     bool
}

func newModel() model {
	const numLastResults = 5
	s := spinner.New()
	s.Style = spinnerStyle
	ps := spinner.New()
	ps.Style = spinnerStyle
	return model{
		bloccSpinner: s,
		paasSpinner:  ps,
		results:      make([]resultMsg, numLastResults),
	}
}

func (m model) Init() tea.Cmd {
	return m.bloccSpinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case resultMsg:
		m.results = append(m.results[1:], msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.paasSpinner.Update(msg)
		m.bloccSpinner, cmd = m.bloccSpinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	var s string

	if m.quitting {
		s += "That’s all for today!"
	} else {
		s += m.bloccSpinner.View() + " Tailing some logs..."
		s += "\n\n"
		s += m.paasSpinner.View() + " Tailing some logs..."
	}

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if !m.quitting {
		s += helpStyle.Render("Press any key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	return appStyle.Render(s)
}

func Tailing() {
	// rand.Seed(time.Now().UTC().UnixNano())

	p := tea.NewProgram(newModel())
	namespace := "flow-system"
	podname := "paas-controller-7b6988d9f6-lqh4s"

	// Simulate activity
	go TailLogs(podname, namespace, p)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func TailLogs(name, namespace string, p *tea.Program) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/milosostojic/.kube/config")
	if err != nil {
		fmt.Println(err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	podLogOpts := corev1.PodLogOptions{
		Follow:    true,
		Container: "paas-controller",
	}

	req, err := clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts).Stream(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer req.Close()

	scanner := bufio.NewScanner(req)
	for scanner.Scan() {
		time.Sleep(time.Second)

		p.Send(resultMsg{food: scanner.Text(), duration: 1, container: name})
		if err := scanner.Err(); err != nil {
			break
		}
		// Send the Bubble Tea program a message from outside the
		// tea.Program. This will block until it is ready to receive
		// messages.
	}
	p.Send(tea.KeyMsg{})
}
