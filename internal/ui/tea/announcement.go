package tea

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/google-classroom/internal/api"
)

// AnnouncementItem represents an announcement item in the list.
type AnnouncementItem struct {
	announcement *api.Announcement
}

// Title returns the title of the announcement item.
func (i AnnouncementItem) Title() string {
	preview := i.announcement.Text
	if len(preview) > 50 {
		preview = preview[:47] + "..."
	}
	return preview
}

// Description returns the description of the announcement item.
func (i AnnouncementItem) Description() string {
	return fmt.Sprintf("%s | %s", i.announcement.CreatorUserID, i.announcement.CreateTime[:10])
}

// FilterValue returns the filter value for the announcement item.
func (i AnnouncementItem) FilterValue() string {
	return i.announcement.Text
}

// AnnouncementModel represents the announcement TUI model.
type AnnouncementModel struct {
	course        *api.Course
	apiClient     *api.Client
	announcements []*api.Announcement
	list          list.Model
	spinner       spinner.Model
	paginator     paginator.Model
	loading       bool
	err           error
	width         int
	height        int
	selectedAnn   *api.Announcement
	fullView      bool
}

// NewAnnouncementModel creates a new announcement model.
func NewAnnouncementModel(course *api.Course, apiClient *api.Client) *AnnouncementModel {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// Create paginator
	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff79c6")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4")).Render("•")

	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Announcements"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff79c6")).
		Bold(true)
	l.SetShowStatusBar(false)

	return &AnnouncementModel{
		course:    course,
		apiClient: apiClient,
		list:      l,
		spinner:   s,
		paginator: p,
		loading:   true,
		fullView:  false,
	}
}

// Init initializes the model.
func (m *AnnouncementModel) Init() tea.Cmd {
	return m.loadAnnouncements()
}

// Update handles messages.
func (m *AnnouncementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "b":
			if m.fullView {
				m.fullView = false
				return m, nil
			}
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case "enter":
			if m.fullView {
				m.fullView = false
				return m, nil
			}
			if i := m.list.SelectedItem(); i != nil {
				if item, ok := i.(AnnouncementItem); ok {
					m.selectedAnn = item.announcement
					m.fullView = true
				}
			}
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadAnnouncements()
		case "/":
			// TODO: Implement search
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-10)
		return m, nil

	case announcementsLoadedMsg:
		m.announcements = msg.announcements
		m.loading = false
		m.err = nil
		m.updateList()
		return m, nil

	case announcementsLoadErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the model.
func (m *AnnouncementModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					m.spinner.View(),
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f8f8f2")).
						Render("Loading announcements..."),
				),
			)
	}

	if m.err != nil {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#ff5555")).
						Bold(true).
						Render("Error loading announcements"),
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f8f8f2")).
						Render(m.err.Error()),
				),
			)
	}

	if m.fullView {
		return m.renderFullView()
	}

	// Render list
	listView := m.list.View()

	// Render footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render("↑↓ navigate | enter view | r refresh | b back | q quit")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				listView,
				"",
				footer,
			),
		)
}

// renderFullView renders the full announcement view.
func (m *AnnouncementModel) renderFullView() string {
	if m.selectedAnn == nil {
		return "No announcement selected"
	}

	// Format the announcement text with wrapping
	lines := wrapText(m.selectedAnn.Text, m.width-4)
	content := strings.Join(lines, "\n")

	// Render header
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff79c6")).
		Bold(true).
		Render("From: " + m.selectedAnn.CreatorUserID)

	// Render date
	date := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render(m.selectedAnn.CreateTime[:19])

	// Render content
	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8f8f2")).
		Width(m.width - 4).
		Render(content)

	// Render footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render("Press enter or esc to go back")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				date,
				"",
				body,
				"",
				footer,
			),
		)
}

// loadAnnouncements loads announcements from the API.
func (m *AnnouncementModel) loadAnnouncements() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		announcements, err := m.apiClient.ListAnnouncements(ctx, m.course.ID)
		if err != nil {
			return announcementsLoadErrorMsg{err: err}
		}
		return announcementsLoadedMsg{announcements: announcements}
	}
}

// updateList updates the list with announcements.
func (m *AnnouncementModel) updateList() {
	items := make([]list.Item, len(m.announcements))
	for i, a := range m.announcements {
		items[i] = AnnouncementItem{announcement: a}
	}
	m.list.SetItems(items)
}

// wrapText wraps text to fit within the specified width.
func wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// announcementsLoadedMsg is sent when announcements are loaded.
type announcementsLoadedMsg struct {
	announcements []*api.Announcement
}

// announcementsLoadErrorMsg is sent when announcements fail to load.
type announcementsLoadErrorMsg struct {
	err error
}
