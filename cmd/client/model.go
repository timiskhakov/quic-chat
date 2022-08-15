package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timiskhakov/quic-chat/internal/chat"
	"strings"
	"sync"
)

type errMsg error

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	err      error
	lines    []string
	send     func(text string) error
	messages <-chan chat.Message
	mutex    sync.Mutex
}

func initialModel(send func(text string) error, messages <-chan chat.Message) *model {
	vp := viewport.New(30, 10)

	ta := textarea.New()
	ta.Placeholder = "Message"
	ta.Focus()
	ta.SetWidth(30)
	ta.SetHeight(1)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &model{
		viewport: vp,
		textarea: ta,
		err:      nil,
		lines:    []string{},
		send:     send,
		messages: messages,
	}
}

func (m *model) Init() tea.Cmd {
	go func() {
		for message := range m.messages {
			m.mutex.Lock()
			m.lines = append(m.lines, fmt.Sprintf("[%s]: %s", message.Nickname, message.Text))
			m.mutex.Unlock()
		}
	}()

	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		taCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(msg)
	m.textarea, taCmd = m.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.err = m.send(m.textarea.Value())
			m.textarea.Reset()
			break
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.mutex.Lock()
	m.viewport.SetContent(strings.Join(m.lines, "\n"))
	m.mutex.Unlock()

	m.viewport.GotoBottom()

	return m, tea.Batch(vpCmd, taCmd)
}

func (m *model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
