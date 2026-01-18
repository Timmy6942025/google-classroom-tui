package tea

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/google-classroom/internal/api"
)

// SubmissionModel represents the submission TUI model.
type SubmissionModel struct {
	course      *api.Course
	courseWork  *api.CourseWork
	apiClient   *api.Client
	submissions []*api.StudentSubmission
	table       table.Model
	loading     bool
	err         error
	width       int
	height      int
}

// NewSubmissionModel creates a new submission model.
func NewSubmissionModel(course *api.Course, courseWork *api.CourseWork, apiClient *api.Client) *SubmissionModel {
	t := table.New()
	t.SetHeight(15)

	return &SubmissionModel{
		course:     course,
		courseWork: courseWork,
		apiClient:  apiClient,
		table:      t,
		loading:    true,
	}
}

// Init initializes the model.
func (m *SubmissionModel) Init() tea.Cmd {
	return m.loadSubmissions()
}

// Update handles messages.
func (m *SubmissionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "b":
			return m, func() tea.Msg { return NavigateBackMsg{} }
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadSubmissions()
		case "t":
			return m, m.handleTurnIn()
		case "enter":
			return m, m.handleViewSubmission()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 15)
		return m, nil

	case submissionsLoadedMsg:
		m.submissions = msg.submissions
		m.loading = false
		m.err = nil
		m.updateTable()
		return m, nil

	case submissionsLoadErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case submissionUpdatedMsg:
		m.loading = true
		m.err = nil
		return m, m.loadSubmissions()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the model.
func (m *SubmissionModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Center,
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#bd93f9")).
						Render("Loading submissions..."),
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
						Render("Error loading submissions"),
					lipgloss.NewStyle().
						Foreground(lipgloss.Color("#f8f8f2")).
						Render(m.err.Error()),
				),
			)
	}

	// Render header
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff79c6")).
		Bold(true).
		Render(m.courseWork.Title)

	// Render table
	tableView := m.table.View()

	// Render footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		Render("↑↓ navigate | enter view | t turn in | r refresh | b back | q quit")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				"",
				tableView,
				"",
				footer,
			),
		)
}

// loadSubmissions loads submissions from the API.
func (m *SubmissionModel) loadSubmissions() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		submissions, err := m.apiClient.ListStudentSubmissions(ctx, m.course.ID, m.courseWork.ID)
		if err != nil {
			return submissionsLoadErrorMsg{err: err}
		}
		return submissionsLoadedMsg{submissions: submissions}
	}
}

// updateTable updates the table with submission data.
func (m *SubmissionModel) updateTable() {
	columns := []table.Column{
		{Title: "State", Width: 15},
		{Title: "Grade", Width: 10},
		{Title: "Late", Width: 10},
		{Title: "Updated", Width: 20},
	}

	rows := make([]table.Row, len(m.submissions))
	for i, s := range m.submissions {
		grade := "Not graded"
		if s.AssignedGrade > 0 {
			grade = fmt.Sprintf("%d/%d", s.AssignedGrade, m.courseWork.MaxPoints)
		}
		late := "No"
		if s.Late {
			late = "Yes"
		}
		rows[i] = table.Row{
			s.State,
			grade,
			late,
			s.UpdateTime[:19],
		}
	}

	m.table.SetColumns(columns)
	m.table.SetRows(rows)
}

// handleTurnIn handles the turn-in action.
func (m *SubmissionModel) handleTurnIn() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Find the current user's submission
		// For simplicity, we'll turn in the first submission in the list
		if len(m.submissions) == 0 {
			return errorMsg{err: fmt.Errorf("no submissions found")}
		}

		sub := m.submissions[0]
		if sub.State != "NEW" && sub.State != "CREATED" {
			return errorMsg{err: fmt.Errorf("submission cannot be turned in")}
		}

		err := m.apiClient.TurnIn(ctx, m.course.ID, m.courseWork.ID, sub.ID)
		if err != nil {
			return errorMsg{err: err}
		}

		return submissionUpdatedMsg{}
	}
}

// handleViewSubmission handles viewing submission details.
func (m *SubmissionModel) handleViewSubmission() tea.Cmd {
	if len(m.submissions) == 0 {
		return nil
	}

	selected := m.table.Cursor()
	if selected >= 0 && selected < len(m.submissions) {
		sub := m.submissions[selected]
		return func() tea.Msg {
			return SubmissionDetailMsg{
				Course:     m.course,
				CourseWork: m.courseWork,
				Submission: sub,
			}
		}
	}
	return nil
}

// submissionsLoadedMsg is sent when submissions are loaded.
type submissionsLoadedMsg struct {
	submissions []*api.StudentSubmission
}

// submissionsLoadErrorMsg is sent when submissions fail to load.
type submissionsLoadErrorMsg struct {
	err error
}

// submissionUpdatedMsg is sent when a submission is updated.
type submissionUpdatedMsg struct{}

// SubmissionDetailMsg is sent when a submission is selected.
type SubmissionDetailMsg struct {
	Course     *api.Course
	CourseWork *api.CourseWork
	Submission *api.StudentSubmission
}

// errorMsg is sent when an error occurs.
type errorMsg struct {
	err error
}
