package blocc

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
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle     = helpStyle.Copy().UnsetMargins()
	// durationStyle = dotStyle.Copy()
	appStyle = lipgloss.NewStyle().Margin(1, 2, 0, 2)
)

/*
type bloccResultMsg struct {
	duration  time.Duration
	food      string
	container string
}

func (r bloccResultMsg) String() string {
	if r.duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("🧗‍♀️ %s %s", r.food, durationStyle.Render(r.container))
	// durationStyle.Render(r.duration.String()))
}

type paasResultMsg struct {
	duration  time.Duration
	food      string
	container string
}

func (r paasResultMsg) String() string {
	if r.duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("🧗‍♀️ %s %s", r.food, durationStyle.Render(r.container))
	// durationStyle.Render(r.duration.String()))
}
*/

type ResultMsg struct {
	duration  time.Duration
	food      string
	container string
}

func (r ResultMsg) String() string {
	if r.duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("🧗‍♀️ %s", r.food)
	// durationStyle.Render(r.duration.String()))
}

type model struct {
	spinner spinner.Model
	keys    []string
	// bloccResults []bloccResultMsg
	// paasResults  []paasResultMsg
	results  map[string][]ResultMsg
	quitting bool
}

func newModel(keys []string) model {
	const numLastResults = 5
	s := spinner.New()
	s.Style = spinnerStyle
	r := make(map[string][]ResultMsg)
	for _, k := range keys {
		r[k] = make([]ResultMsg, numLastResults)
	}
	return model{
		spinner: s,
		keys:    keys,
		// bloccResults: make([]bloccResultMsg, numLastResults),
		// paasResults:  make([]paasResultMsg, numLastResults),
		results: r,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case ResultMsg:
		c := m.results[msg.container]
		m.results[msg.container] = append(c[1:], msg)
		return m, nil
	// case bloccResultMsg:
	// 	m.bloccResults = append(m.bloccResults[1:], msg)
	// 	return m, nil
	// case paasResultMsg:
	// 	m.paasResults = append(m.paasResults[1:], msg)
	// 	return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
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
		s += m.spinner.View() + " Tailing some logs..."
	}

	s += "\n\n"

	// for _, res := range m.bloccResults {
	// 	s += res.String() + "\n"
	// }

	// s += "\n\n"

	// for _, res := range m.paasResults {
	// 	s += res.String() + "\n"
	// }

	for _, k := range m.keys {
		for _, r := range m.results[k] {
			s += r.String() + "\n"
		}
		s += "\n\n"
	}

	if !m.quitting {
		s += helpStyle.Render("Press any key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	return appStyle.Render(s)
}

type roadRunner struct {
	PodName       string
	ContainerName string
	Namespace     string
}

func Tailing() {
	// rand.Seed(time.Now().UTC().UnixNano())

	p := tea.NewProgram(newModel([]string{"blocc-controller-manager-c584d6995-jps7v", "paas-controller-7b6988d9f6-lqh4s"}))

	load := []roadRunner{
		{
			PodName:       "blocc-controller-manager-c584d6995-jps7v",
			ContainerName: "manager",
			Namespace:     "blocc",
		},
		{
			PodName:       "paas-controller-7b6988d9f6-lqh4s",
			ContainerName: "paas-controller",
			Namespace:     "flow-system",
		},
	}

	// bloccNamespace := "blocc"
	// bloccPodname := "blocc-controller-manager-c584d6995-jps7v"

	// paasNamespace := "flow-system"
	// paasPodname := "paas-controller-7b6988d9f6-lqh4s"

	// Simulate activity
	// go bloccTailLogs(bloccPodname, bloccNamespace, p)
	// go paasTailLogs(paasPodname, paasNamespace, p)

	for _, rr := range load {
		go tailLogs(rr.PodName, rr.Namespace, rr.ContainerName, p)
	}

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// func bloccTailLogs(name, namespace string, p *tea.Program) {
// 	config, err := clientcmd.BuildConfigFromFlags("", "/Users/milosostojic/.kube/config")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	podLogOpts := corev1.PodLogOptions{
// 		Follow:    true,
// 		Container: "manager",
// 	}

// 	req, err := clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts).Stream(context.Background())
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer req.Close()

// 	scanner := bufio.NewScanner(req)
// 	for scanner.Scan() {
// 		time.Sleep(time.Second)

// 		p.Send(ResultMsg{food: scanner.Text(), duration: 1, container: name})
// 		if err := scanner.Err(); err != nil {
// 			break
// 		}
// 		// Send the Bubble Tea program a message from outside the
// 		// tea.Program. This will block until it is ready to receive
// 		// messages.
// 	}
// 	p.Send(tea.KeyMsg{})
// }

// func paasTailLogs(name, namespace string, p *tea.Program) {
// 	config, err := clientcmd.BuildConfigFromFlags("", "/Users/milosostojic/.kube/config")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	podLogOpts := corev1.PodLogOptions{
// 		Follow:    true,
// 		Container: "paas-controller",
// 	}

// 	req, err := clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts).Stream(context.Background())
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer req.Close()

// 	scanner := bufio.NewScanner(req)
// 	for scanner.Scan() {
// 		time.Sleep(time.Second)

// 		p.Send(ResultMsg{food: scanner.Text(), duration: 1, container: name})
// 		if err := scanner.Err(); err != nil {
// 			break
// 		}
// 		// Send the Bubble Tea program a message from outside the
// 		// tea.Program. This will block until it is ready to receive
// 		// messages.
// 	}
// 	p.Send(tea.KeyMsg{})
// }

func tailLogs(name, namespace, containerName string, p *tea.Program) {
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
		Container: containerName,
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

		p.Send(ResultMsg{food: scanner.Text(), duration: 1, container: name})
		if err := scanner.Err(); err != nil {
			break
		}
		// Send the Bubble Tea program a message from outside the
		// tea.Program. This will block until it is ready to receive
		// messages.
	}
	p.Send(tea.KeyMsg{})
}
