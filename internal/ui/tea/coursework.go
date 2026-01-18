package tea

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/google-classroom/internal/api"
)

// Filter type for coursework
type CourseworkFilter int

const (
	FilterAll CourseworkFilter = iota
	FilterAssignments
	FilterMaterials
	FilterAnnouncements
)

func (f CourseworkFilter) String() string {
	switch f {
	case FilterAll:
		return "All"
	case FilterAssignments:
		return "Assignments"
	case FilterMaterials:
		return "Materials"
	case FilterAnnouncements:
		return "Announcements"
	default:
		return "Unknown"
	}
}

// CourseworkItem represents a coursework item in the list.
type CourseworkItem struct {
	coursework *api.CourseWork
	filter     CourseworkFilter
}

// Title returns the title of the coursework item.
func (i CourseworkItem) Title() string {
	return i.coursework.Title
}

// Description returns the description of the coursework item.
func (i CourseworkItem) Description() string {
	status := ""
	if i.coursework.DueDate != "" {
		status = fmt.Sprintf("Due: %s", i.coursework.DueDate)
	}
	if i.coursework.MaxPoints > 0 {
		if status != "" {
			status += " | "
		}
		status += fmt.Sprintf("%d pts", i.coursework.MaxPoints)
	}
	return status
}

// FilterValue returns the filter value for the coursework item.
func (i CourseworkItem) FilterValue() string {
	return i.coursework.Title
}

// CourseworkModel represents the coursework TUI model.
type CourseworkModel struct {
	course     *api.Course
	apiClient  *api.Client
	coursework []*api.CourseWork
	filteredCW []*api.CourseWork
	filter     CourseworkFilter
	list       list.Model
	spinner    spinner.Model
	loading    bool
	err        error
	width      int
	height     int
	selectedCW *api.CourseWork
}

// NewCourseworkModel creates a new coursework model.
func NewCourseworkModel(course *api.Course, apiClient *api.Client) *CourseworkModel {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Coursework"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff79c6")).
		Bold(true)

	return &CourseworkModel{
		course:    course,
		apiClient: apiClient,
		filter:    FilterAll,
		list:      l,
		spinner:   s,
		loading:   true,
	}
}

// Init initializes the model.
func (m *CourseworkModel) Init() tea.Cmd {
	return m.loadCoursework()
}

// Update handles messages.
func (m *CourseworkModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "b":
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case "a":
			m.filter = FilterAssignments
			m.updateList()
		case "m":
			m.filter = FilterMaterials
			m.updateList()
		case "n":
			m.filter = FilterAnnouncements
			m.updateList()
		case "all", "A":
			m.filter = FilterAll
			m.updateList()
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadCoursework()
		case "enter":
			if i := m.list.SelectedItem(); i != nil {
				if item, ok := i.(CourseworkItem); ok {
					m.selectedCW = item.coursework
					return m, func() tea.Msg {
						return SubmissionListMsg{
							Course:     m.course,
							CourseWork: item.coursework,
						}
					}
				}
			}
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

	case courseworkLoadedMsg:
		m.coursework = msg.coursework
		m.filteredCW = msg.coursework
		m.loading = false
		m.err = nil
		m.updateList()
		return m, nil

	case courseworkLoadErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the model.
func (m *CourseworkModel) View() string {
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
						Render("Loading coursework..."),
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
						Render("Error loading coursework"),
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f8f8f2")).
						Render(m.err.Error()),
				),
			)
	}

	// Render filter status
	filterInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#bd93f9")).
		Render(fmt.Sprintf("Filter: %s (press a/m/n/all)", m.filter))

	// Render list
	listView := m.list.View()

	// Render footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render("↑↓ navigate | enter select | a/m/n filter | r refresh | b back | q quit")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				filterInfo,
				"",
				listView,
				"",
				footer,
			),
		)
}

// loadCoursework loads coursework from the API.
func (m *CourseworkModel) loadCoursework() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		coursework, err := m.apiClient.ListCourseWork(ctx, m.course.ID)
		if err != nil {
			return courseworkLoadErrorMsg{err: err}
		}
		return courseworkLoadedMsg{coursework: coursework}
	}
}

// updateList updates the list with filtered coursework.
func (m *CourseworkModel) updateList() {
	// Filter based on filter type
	if m.filter == FilterAll {
		m.filteredCW = m.coursework
	} else {
		m.filteredCW = make([]*api.CourseWork, 0)
		for _, cw := range m.coursework {
			if m.filter == FilterAssignments && cw.WorkType == "ASSIGNMENT" {
				m.filteredCW = append(m.filteredCW, cw)
			} else if m.filter == FilterMaterials && cw.WorkType == "MATERIAL" {
				m.filteredCW = append(m.filteredCW, cw)
			} else if m.filter == FilterAnnouncements && cw.WorkType == "SHORT_ANSWER_QUESTION" {
				m.filteredCW = append(m.filteredCW, cw)
			}
		}
	}

	// Create list items
	items := make([]list.Item, len(m.filteredCW))
	for i, cw := range m.filteredCW {
		items[i] = CourseworkItem{coursework: cw, filter: m.filter}
	}
	m.list.SetItems(items)
}

// SelectedCourseWork returns the currently selected coursework.
func (m *CourseworkModel) SelectedCourseWork() *api.CourseWork {
	return m.selectedCW
}

// courseworkLoadedMsg is sent when coursework is loaded.
type courseworkLoadedMsg struct {
	coursework []*api.CourseWork
}

// courseworkLoadErrorMsg is sent when coursework fails to load.
type courseworkLoadErrorMsg struct {
	err error
}

// SubmissionListMsg is sent when coursework is selected.
type SubmissionListMsg struct {
	Course     *api.Course
	CourseWork *api.CourseWork
}
