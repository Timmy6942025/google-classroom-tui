package tea

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/google-classroom/internal/api"
)

// Tab definitions
type Tab int

const (
	TabCoursework Tab = iota
	TabStudents
	TabTeachers
	TabAnnouncements
)

func (t Tab) String() string {
	switch t {
	case TabCoursework:
		return "Coursework"
	case TabStudents:
		return "Students"
	case TabTeachers:
		return "Teachers"
	case TabAnnouncements:
		return "Announcements"
	default:
		return "Unknown"
	}
}

// CourseDetailModel represents the course detail TUI model.
type CourseDetailModel struct {
	course        *api.Course
	apiClient     *api.Client
	coursework    []*api.CourseWork
	students      []*api.Student
	teachers      []*api.Teacher
	announcements []*api.Announcement
	activeTab     Tab
	table         table.Model
	loading       bool
	err           error
	width         int
	height        int
}

// NewCourseDetailModel creates a new course detail model.
func NewCourseDetailModel(course *api.Course, apiClient *api.Client) *CourseDetailModel {
	// Create table with basic configuration
	t := table.New()
	t.SetHeight(20)

	return &CourseDetailModel{
		course:    course,
		apiClient: apiClient,
		activeTab: TabCoursework,
		table:     t,
		loading:   true,
	}
}

// Init initializes the model.
func (m *CourseDetailModel) Init() tea.Cmd {
	return m.loadData()
}

// Update handles messages.
func (m *CourseDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "b":
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case "left", "h":
			m.prevTab()
		case "right", "l":
			m.nextTab()
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadData()
		case "enter":
			return m, m.handleEnter()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 15)
		return m, nil

	case dataLoadedMsg:
		m.coursework = msg.coursework
		m.students = msg.students
		m.teachers = msg.teachers
		m.announcements = msg.announcements
		m.loading = false
		m.err = nil
		m.updateTable()
		return m, nil

	case dataLoadErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the model.
func (m *CourseDetailModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#ff79c6")).
						Bold(true).
						Render(m.course.Name),
					"",
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#bd93f9")).
						Render("Loading data..."),
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
						Render("Error loading data"),
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f8f8f2")).
						Render(m.err.Error()),
				),
			)
	}

	// Render header
	header := m.renderHeader()

	// Render tabs
	tabs := m.renderTabs()

	// Render table
	tableView := m.table.View()

	// Render footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render("←→/hl change tab | enter select | b back | r refresh | q quit")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				"",
				tabs,
				"",
				tableView,
				"",
				footer,
			),
		)
}

// renderHeader renders the course header.
func (m *CourseDetailModel) renderHeader() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff79c6")).
		Bold(true).
		Width(m.width - 4)

	lines := []string{m.course.Name}
	if m.course.Section != "" {
		lines = append(lines, fmt.Sprintf("Section: %s", m.course.Section))
	}
	if m.course.Room != "" {
		lines = append(lines, fmt.Sprintf("Room: %s", m.course.Room))
	}

	return style.Render(
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2")).
			Render(strings.Join(lines, "\n")),
	)
}

// renderTabs renders the tab bar.
func (m *CourseDetailModel) renderTabs() string {
	var tabs []string
	for i := Tab(0); i <= TabAnnouncements; i++ {
		if i == m.activeTab {
			tabs = append(tabs, lipgloss.NewStyle().
				Background(lipgloss.Color("#6272a4")).
				Foreground(lipgloss.Color("#f8f8f2")).
				Padding(0, 2).
				Render(" "+i.String()+" "))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272a4")).
				Padding(0, 2).
				Render(" "+i.String()+" "))
		}
	}

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				tabs...,
			),
		)
}

// loadData loads all course data.
func (m *CourseDetailModel) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		coursework, err := m.apiClient.ListCourseWork(ctx, m.course.ID)
		if err != nil {
			return dataLoadErrorMsg{err: err}
		}

		students, err := m.apiClient.ListStudents(ctx, m.course.ID)
		if err != nil {
			return dataLoadErrorMsg{err: err}
		}

		teachers, err := m.apiClient.ListTeachers(ctx, m.course.ID)
		if err != nil {
			return dataLoadErrorMsg{err: err}
		}

		announcements, err := m.apiClient.ListAnnouncements(ctx, m.course.ID)
		if err != nil {
			return dataLoadErrorMsg{err: err}
		}

		return dataLoadedMsg{
			coursework:    coursework,
			students:      students,
			teachers:      teachers,
			announcements: announcements,
		}
	}
}

// updateTable updates the table based on the active tab.
func (m *CourseDetailModel) updateTable() {
	var rows []table.Row
	var columns []table.Column

	switch m.activeTab {
	case TabCoursework:
		columns = []table.Column{
			{Title: "Title", Width: 40},
			{Title: "Type", Width: 15},
			{Title: "Due", Width: 15},
			{Title: "Points", Width: 10},
		}
		for _, cw := range m.coursework {
			dueDate := ""
			if cw.DueDate != "" {
				dueDate = cw.DueDate
			}
			rows = append(rows, table.Row{
				cw.Title,
				cw.WorkType,
				dueDate,
				fmt.Sprintf("%d", cw.MaxPoints),
			})
		}

	case TabStudents:
		columns = []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Email", Width: 40},
		}
		for _, s := range m.students {
			rows = append(rows, table.Row{
				s.Profile.Name,
				s.Profile.EmailAddress,
			})
		}

	case TabTeachers:
		columns = []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Email", Width: 40},
		}
		for _, t := range m.teachers {
			rows = append(rows, table.Row{
				t.Profile.Name,
				t.Profile.EmailAddress,
			})
		}

	case TabAnnouncements:
		columns = []table.Column{
			{Title: "Text", Width: 60},
			{Title: "Date", Width: 20},
		}
		for _, a := range m.announcements {
			preview := a.Text
			if len(preview) > 55 {
				preview = preview[:52] + "..."
			}
			rows = append(rows, table.Row{
				preview,
				a.CreateTime[:10],
			})
		}
	}

	m.table.SetColumns(columns)
	m.table.SetRows(rows)
}

// prevTab moves to the previous tab.
func (m *CourseDetailModel) prevTab() {
	if m.activeTab > 0 {
		m.activeTab--
		m.updateTable()
	}
}

// nextTab moves to the next tab.
func (m *CourseDetailModel) nextTab() {
	if m.activeTab < TabAnnouncements {
		m.activeTab++
		m.updateTable()
	}
}

// handleEnter handles enter key press.
func (m *CourseDetailModel) handleEnter() tea.Cmd {
	switch m.activeTab {
	case TabCoursework:
		if len(m.coursework) > 0 {
			selected := m.table.Cursor()
			if selected >= 0 && selected < len(m.coursework) {
				cw := m.coursework[selected]
				return func() tea.Msg {
					return CourseWorkSelectedMsg{
						Course:     m.course,
						CourseWork: cw,
					}
				}
			}
		}
	case TabAnnouncements:
		if len(m.announcements) > 0 {
			selected := m.table.Cursor()
			if selected >= 0 && selected < len(m.announcements) {
				a := m.announcements[selected]
				return func() tea.Msg {
					return AnnouncementSelectedMsg{
						Course:       m.course,
						Announcement: a,
					}
				}
			}
		}
	}
	return nil
}

// dataLoadedMsg is sent when data is loaded.
type dataLoadedMsg struct {
	coursework    []*api.CourseWork
	students      []*api.Student
	teachers      []*api.Teacher
	announcements []*api.Announcement
}

// dataLoadErrorMsg is sent when data fails to load.
type dataLoadErrorMsg struct {
	err error
}

// CourseWorkSelectedMsg is sent when coursework is selected.
type CourseWorkSelectedMsg struct {
	Course     *api.Course
	CourseWork *api.CourseWork
}

// AnnouncementSelectedMsg is sent when an announcement is selected.
type AnnouncementSelectedMsg struct {
	Course       *api.Course
	Announcement *api.Announcement
}

// NavigateBackMsg is sent when the user wants to go back.
type NavigateBackMsg struct{}
