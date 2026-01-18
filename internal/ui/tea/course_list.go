package tea

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/google-classroom/internal/api"
)

// CourseListModel represents the course list TUI model.
type CourseListModel struct {
	list            list.Model
	spinner         spinner.Model
	apiClient       *api.Client
	courses         []*api.Course
	filteredCourses []*api.Course
	searchQuery     string
	searchInput     textinput.Model
	loading         bool
	err             error
	width           int
	height          int
	selectedCourse  *api.Course
}

// CourseItem represents a course item in the list.
type CourseItem struct {
	course *api.Course
}

// Title returns the title of the course item.
func (i CourseItem) Title() string {
	return i.course.Name
}

// Description returns the description of the course item.
func (i CourseItem) Description() string {
	section := ""
	if i.course.Section != "" {
		section = fmt.Sprintf(" | %s", i.course.Section)
	}
	return fmt.Sprintf("%s%s", i.course.CourseState, section)
}

// FilterValue returns the filter value for the course item.
func (i CourseItem) FilterValue() string {
	return i.course.Name + " " + i.course.Section
}

// NewCourseListModel creates a new course list model.
func NewCourseListModel(apiClient *api.Client) *CourseListModel {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Search courses..."
	ti.Prompt = "/"
	ti.Width = 30
	ti.Focus()

	// Create list
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Your Courses"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff79c6")).
		Bold(true)
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4"))
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4"))

	return &CourseListModel{
		list:        l,
		spinner:     s,
		apiClient:   apiClient,
		searchInput: ti,
		loading:     true,
	}
}

// Init initializes the model.
func (m *CourseListModel) Init() tea.Cmd {
	return m.loadCourses()
}

// Update handles messages.
func (m *CourseListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			m.searchInput.Focus()
			return m, textinput.Blink
		case "enter":
			if i := m.list.SelectedItem(); i != nil {
				if item, ok := i.(CourseItem); ok {
					m.selectedCourse = item.course
					return m, func() tea.Msg { return CourseSelectedMsg{Course: item.course} }
				}
			}
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadCourses()
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

	case coursesLoadedMsg:
		m.courses = msg.courses
		m.filteredCourses = msg.courses
		m.loading = false
		m.err = nil
		m.updateList()
		return m, nil

	case coursesLoadErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	// Update search input if focused
	if m.searchInput.Focused() {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		if cmd != nil {
			go m.handleSearch()
		}
		return m, cmd
	}

	// Update list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the model.
func (m *CourseListModel) View() string {
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
						Render("Loading courses..."),
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
						Render("Error loading courses"),
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f8f8f2")).
						Render(m.err.Error()),
					"",
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#6272a4")).
						Render("Press 'r' to retry"),
				),
			)
	}

	// Render search input
	searchView := ""
	if m.searchInput.Focused() {
		searchView = m.searchInput.View()
	} else {
		searchView = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Render("/ to search")
	}

	// Render list
	listView := m.list.View()

	// Render footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render("↑↓ navigate | enter select | / search | r refresh | q quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		searchView,
		"",
		listView,
		"",
		footer,
	)
}

// loadCourses loads courses from the API.
func (m *CourseListModel) loadCourses() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		courses, err := m.apiClient.ListCourses(ctx)
		if err != nil {
			return coursesLoadErrorMsg{err: err}
		}
		return coursesLoadedMsg{courses: courses}
	}
}

// updateList updates the list with filtered courses.
func (m *CourseListModel) updateList() {
	items := make([]list.Item, len(m.filteredCourses))
	for i, course := range m.filteredCourses {
		items[i] = CourseItem{course: course}
	}
	m.list.SetItems(items)
}

// handleSearch handles search input changes.
func (m *CourseListModel) handleSearch() {
	query := strings.ToLower(m.searchInput.Value())

	if query == "" {
		m.filteredCourses = m.courses
	} else {
		m.filteredCourses = make([]*api.Course, 0)
		for _, course := range m.courses {
			if strings.Contains(strings.ToLower(course.Name), query) ||
				strings.Contains(strings.ToLower(course.Section), query) {
				m.filteredCourses = append(m.filteredCourses, course)
			}
		}
	}

	m.updateList()
}

// SelectedCourse returns the currently selected course.
func (m *CourseListModel) SelectedCourse() *api.Course {
	return m.selectedCourse
}

// coursesLoadedMsg is sent when courses are loaded.
type coursesLoadedMsg struct {
	courses []*api.Course
}

// coursesLoadErrorMsg is sent when courses fail to load.
type coursesLoadErrorMsg struct {
	err error
}

// CourseSelectedMsg is sent when a course is selected.
type CourseSelectedMsg struct {
	Course *api.Course
}
