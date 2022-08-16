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

type app struct {
	viewport viewport.Model
	textarea textarea.Model
	err      error
	lines    []string
	send     func(text string) error
	messages <-chan chat.Message
	mutex    sync.Mutex
}

func createApp(send func(text string) error, messages <-chan chat.Message) *app {
	vp := viewport.New(30, 10)

	ta := textarea.New()
	ta.Placeholder = "Message"
	ta.Focus()
	ta.SetWidth(30)
	ta.SetHeight(1)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return &app{
		viewport: vp,
		textarea: ta,
		err:      nil,
		lines:    []string{},
		send:     send,
		messages: messages,
	}
}

func (a *app) Init() tea.Cmd {
	go func() {
		for message := range a.messages {
			a.mutex.Lock()
			a.lines = append(a.lines, fmt.Sprintf("[%s]: %s", message.Nickname, message.Text))
			a.mutex.Unlock()
		}
	}()

	return textinput.Blink
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		taCmd tea.Cmd
	)

	a.viewport, vpCmd = a.viewport.Update(msg)
	a.textarea, taCmd = a.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return a, tea.Quit
		case tea.KeyEnter:
			a.err = a.send(a.textarea.Value())
			a.textarea.Reset()
			break
		}
	case errMsg:
		a.err = msg
		return a, nil
	}

	a.mutex.Lock()
	a.viewport.SetContent(strings.Join(a.lines, "\n"))
	a.mutex.Unlock()

	a.viewport.GotoBottom()

	return a, tea.Batch(vpCmd, taCmd)
}

func (a *app) View() string {
	return fmt.Sprintf("%s\n%s\n", a.viewport.View(), a.textarea.View())
}
