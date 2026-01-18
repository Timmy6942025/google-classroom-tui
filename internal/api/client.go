// Package api provides the Google Classroom API client wrapper.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

// Client wraps the Google Classroom API with additional functionality.
type Client struct {
	service    *classroom.Service
	httpClient *http.Client
}

// Configuration holds API client configuration.
type Configuration struct {
	RateLimitBackoff time.Duration
	MaxRetries       int
}

// DefaultConfiguration returns the default client configuration.
func DefaultConfiguration() *Configuration {
	return &Configuration{
		RateLimitBackoff: 1 * time.Second,
		MaxRetries:       3,
	}
}

// NewClient creates a new Google Classroom API client.
func NewClient(ctx context.Context, ts oauth2.TokenSource, cfg *Configuration) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfiguration()
	}

	// Create HTTP client with OAuth token source
	httpClient := oauth2.NewClient(ctx, ts)

	// Create Classroom service
	service, err := classroom.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create classroom service: %w", err)
	}

	return &Client{
		service:    service,
		httpClient: httpClient,
	}, nil
}

// Course represents a Google Classroom course.
type Course struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Section        string `json:"section"`
	Description    string `json:"descriptionHeading"`
	Room           string `json:"room"`
	OwnerID        string `json:"ownerId"`
	EnrollmentCode string `json:"enrollmentCode"`
	CourseState    string `json:"courseState"`
	TimeCreated    string `json:"timeCreated"`
	UpdateTime     string `json:"updateTime"`
}

// CourseWork represents an assignment or material in a course.
type CourseWork struct {
	ID            string `json:"id"`
	CourseID      string `json:"courseId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	WorkType      string `json:"workType"`
	State         string `json:"state"`
	DueDate       string `json:"dueDate"`
	DueTime       string `json:"dueTime"`
	MaxPoints     int    `json:"maxPoints"`
	CreatorUserID string `json:"creatorUserId"`
	UpdateTime    string `json:"updateTime"`
}

// StudentSubmission represents a student's submission for coursework.
type StudentSubmission struct {
	ID            string `json:"id"`
	CourseID      string `json:"courseId"`
	CourseWorkID  string `json:"courseWorkId"`
	UserID        string `json:"userId"`
	State         string `json:"state"`
	AssignedGrade int    `json:"assignedGrade"`
	DraftGrade    int    `json:"draftGrade"`
	Late          bool   `json:"late"`
	CreateTime    string `json:"createTime"`
	UpdateTime    string `json:"updateTime"`
}

// Announcement represents a course announcement.
type Announcement struct {
	ID            string `json:"id"`
	CourseID      string `json:"courseId"`
	Text          string `json:"text"`
	State         string `json:"state"`
	CreatorUserID string `json:"creatorUserId"`
	CreateTime    string `json:"createTime"`
	UpdateTime    string `json:"updateTime"`
}

// Student represents a course student.
type Student struct {
	UserID   string      `json:"userId"`
	Profile  UserProfile `json:"profile"`
	CourseID string      `json:"courseId"`
}

// Teacher represents a course teacher.
type Teacher struct {
	UserID   string      `json:"userId"`
	Profile  UserProfile `json:"profile"`
	CourseID string      `json:"courseId"`
}

// UserProfile represents a user's profile information.
type UserProfile struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	PhotoURL     string `json:"photoUrl"`
}

// ListCoursesResponse represents the response from listing courses.
type ListCoursesResponse struct {
	Courses       []*Course `json:"courses"`
	NextPageToken string    `json:"nextPageToken"`
}

// ListCourseWorkResponse represents the response from listing coursework.
type ListCourseWorkResponse struct {
	CourseWork    []*CourseWork `json:"courseWork"`
	NextPageToken string        `json:"nextPageToken"`
}

// ListStudentSubmissionsResponse represents the response from listing submissions.
type ListStudentSubmissionsResponse struct {
	StudentSubmissions []*StudentSubmission `json:"studentSubmissions"`
	NextPageToken      string               `json:"nextPageToken"`
}

// ListAnnouncementsResponse represents the response from listing announcements.
type ListAnnouncementsResponse struct {
	Announcements []*Announcement `json:"announcements"`
	NextPageToken string          `json:"nextPageToken"`
}

// ListStudentsResponse represents the response from listing students.
type ListStudentsResponse struct {
	Students      []*Student `json:"students"`
	NextPageToken string     `json:"nextPageToken"`
}

// ListTeachersResponse represents the response from listing teachers.
type ListTeachersResponse struct {
	Teachers      []*Teacher `json:"teachers"`
	NextPageToken string     `json:"nextPageToken"`
}

// ListCourses retrieves all courses the user has access to.
func (c *Client) ListCourses(ctx context.Context) ([]*Course, error) {
	var courses []*Course
	pageToken := ""

	for {
		req := c.service.Courses.List()
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		resp, err := c.executeWithRetry(ctx, func() (*classroom.ListCoursesResponse, error) {
			return req.Do()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list courses: %w", err)
		}

		for _, course := range resp.Courses {
			courses = append(courses, convertCourse(course))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return courses, nil
}

// GetCourse retrieves a specific course by ID.
func (c *Client) GetCourse(ctx context.Context, courseID string) (*Course, error) {
	resp, err := c.executeWithRetry(ctx, func() (*classroom.Course, error) {
		return c.service.Courses.Get(courseID).Do()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get course %s: %w", courseID, err)
	}

	return convertCourse(resp), nil
}

// ListCourseWork retrieves all coursework for a course.
func (c *Client) ListCourseWork(ctx context.Context, courseID string) ([]*CourseWork, error) {
	var coursework []*CourseWork
	pageToken := ""

	for {
		req := c.service.Courses.CourseWork.List(courseID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		resp, err := c.executeWithRetry(ctx, func() (*classroom.ListCourseWorkResponse, error) {
			return req.Do()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list coursework: %w", err)
		}

		for _, cw := range resp.CourseWork {
			coursework = append(coursework, convertCourseWork(cw))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return coursework, nil
}

// GetCourseWork retrieves specific coursework by ID.
func (c *Client) GetCourseWork(ctx context.Context, courseID, courseWorkID string) (*CourseWork, error) {
	resp, err := c.executeWithRetry(ctx, func() (*classroom.CourseWork, error) {
		return c.service.Courses.CourseWork.Get(courseID, courseWorkID).Do()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get coursework %s: %w", courseWorkID, err)
	}

	return convertCourseWork(resp), nil
}

// ListStudentSubmissions retrieves all submissions for coursework.
func (c *Client) ListStudentSubmissions(ctx context.Context, courseID, courseWorkID string) ([]*StudentSubmission, error) {
	var submissions []*StudentSubmission
	pageToken := ""

	for {
		req := c.service.Courses.CourseWork.StudentSubmissions.List(courseID, courseWorkID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		resp, err := c.executeWithRetry(ctx, func() (*classroom.ListStudentSubmissionsResponse, error) {
			return req.Do()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list submissions: %w", err)
		}

		for _, sub := range resp.StudentSubmissions {
			submissions = append(submissions, convertSubmission(sub))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return submissions, nil
}

// GetStudentSubmission retrieves a specific submission.
func (c *Client) GetStudentSubmission(ctx context.Context, courseID, courseWorkID, submissionID string) (*StudentSubmission, error) {
	resp, err := c.executeWithRetry(ctx, func() (*classroom.StudentSubmission, error) {
		return c.service.Courses.CourseWork.StudentSubmissions.Get(courseID, courseWorkID, submissionID).Do()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get submission %s: %w", submissionID, err)
	}

	return convertSubmission(resp), nil
}

// TurnIn turns in a student's submission.
func (c *Client) TurnIn(ctx context.Context, courseID, courseWorkID, submissionID string) error {
	_, err := c.executeWithRetry(ctx, func() (*classroom.Empty, error) {
		return c.service.Courses.CourseWork.StudentSubmissions.TurnIn(courseID, courseWorkID, submissionID, &classroom.TurnInStudentSubmissionRequest{}).Do()
	})
	if err != nil {
		return fmt.Errorf("failed to turn in submission: %w", err)
	}

	return nil
}

// ListAnnouncements retrieves all announcements for a course.
func (c *Client) ListAnnouncements(ctx context.Context, courseID string) ([]*Announcement, error) {
	var announcements []*Announcement
	pageToken := ""

	for {
		req := c.service.Courses.Announcements.List(courseID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		resp, err := c.executeWithRetry(ctx, func() (*classroom.ListAnnouncementsResponse, error) {
			return req.Do()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list announcements: %w", err)
		}

		for _, ann := range resp.Announcements {
			announcements = append(announcements, convertAnnouncement(ann))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return announcements, nil
}

// ListStudents retrieves all students for a course.
func (c *Client) ListStudents(ctx context.Context, courseID string) ([]*Student, error) {
	var students []*Student
	pageToken := ""

	for {
		req := c.service.Courses.Students.List(courseID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		resp, err := c.executeWithRetry(ctx, func() (*classroom.ListStudentsResponse, error) {
			return req.Do()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list students: %w", err)
		}

		for _, s := range resp.Students {
			students = append(students, convertStudent(s))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return students, nil
}

// ListTeachers retrieves all teachers for a course.
func (c *Client) ListTeachers(ctx context.Context, courseID string) ([]*Teacher, error) {
	var teachers []*Teacher
	pageToken := ""

	for {
		req := c.service.Courses.Teachers.List(courseID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		resp, err := c.executeWithRetry(ctx, func() (*classroom.ListTeachersResponse, error) {
			return req.Do()
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list teachers: %w", err)
		}

		for _, t := range resp.Teachers {
			teachers = append(teachers, convertTeacher(t))
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return teachers, nil
}

// executeWithRetry executes a function with exponential backoff on rate limit errors.
func (c *Client) executeWithRetry(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	backoff := time.Second

	for attempt := 0; attempt < 3; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := fn()
		if err == nil {
			return resp, nil
		}

		// Check for rate limit error (429)
		if isRateLimitError(err) {
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// Check for other API errors
		if isAPIError(err) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("after %d attempts: %w", 3, lastErr)
}

// isRateLimitError checks if the error is a rate limit error.
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit")
}

// isAPIError checks if the error is an API error that should not be retried.
func isAPIError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// 403 (forbidden), 404 (not found), 401 (unauthorized) should not be retried
	return strings.Contains(errStr, "403") || strings.Contains(errStr, "404") || strings.Contains(errStr, "401")
}

// convertCourse converts a Classroom Course to our Course type.
func convertCourse(c *classroom.Course) *Course {
	return &Course{
		ID:             c.Id,
		Name:           c.Name,
		Section:        c.Section,
		Description:    c.DescriptionHeading,
		Room:           c.Room,
		OwnerID:        c.OwnerId,
		EnrollmentCode: c.EnrollmentCode,
		CourseState:    c.CourseState,
		TimeCreated:    c.CreationTime,
		UpdateTime:     c.UpdateTime,
	}
}

// convertCourseWork converts a Classroom CourseWork to our CourseWork type.
func convertCourseWork(cw *classroom.CourseWork) *CourseWork {
	return &CourseWork{
		ID:            cw.Id,
		CourseID:      cw.CourseId,
		Title:         cw.Title,
		Description:   cw.Description,
		WorkType:      cw.WorkType,
		State:         cw.State,
		DueDate:       formatDate(cw.DueDate),
		DueTime:       formatTime(cw.DueTime),
		MaxPoints:     int(cw.MaxPoints),
		CreatorUserID: cw.CreatorUserId,
		UpdateTime:    cw.UpdateTime,
	}
}

// convertSubmission converts a Classroom StudentSubmission to our type.
func convertSubmission(s *classroom.StudentSubmission) *StudentSubmission {
	return &StudentSubmission{
		ID:            s.Id,
		CourseID:      s.CourseId,
		CourseWorkID:  s.CourseWorkId,
		UserID:        s.UserId,
		State:         s.State,
		AssignedGrade: int(s.AssignedGrade),
		DraftGrade:    int(s.DraftGrade),
		Late:          s.Late,
		CreateTime:    s.CreationTime,
		UpdateTime:    s.UpdateTime,
	}
}

// convertAnnouncement converts a Classroom Announcement to our type.
func convertAnnouncement(a *classroom.Announcement) *Announcement {
	return &Announcement{
		ID:            a.Id,
		CourseID:      a.CourseId,
		Text:          a.Text,
		State:         a.State,
		CreatorUserID: a.CreatorUserId,
		CreateTime:    a.CreationTime,
		UpdateTime:    a.UpdateTime,
	}
}

// convertStudent converts a Classroom Student to our type.
func convertStudent(s *classroom.Student) *Student {
	return &Student{
		UserID:   s.UserId,
		Profile:  convertProfile(s.Profile),
		CourseID: s.CourseId,
	}
}

// convertTeacher converts a Classroom Teacher to our type.
func convertTeacher(t *classroom.Teacher) *Teacher {
	return &Teacher{
		UserID:   t.UserId,
		Profile:  convertProfile(t.Profile),
		CourseID: t.CourseId,
	}
}

// convertProfile converts a Classroom UserProfile to our type.
func convertProfile(p *classroom.UserProfile) UserProfile {
	if p == nil {
		return UserProfile{}
	}
	return UserProfile{
		ID:           p.Id,
		Name:         p.Name.FullName,
		EmailAddress: p.EmailAddress,
		PhotoURL:     p.PhotoUrl,
	}
}

// formatDate formats a Classroom Date as a string.
func formatDate(d *classroom.Date) string {
	if d == nil {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// formatTime formats a Classroom TimeOfDay as a string.
func formatTime(t *classroom.TimeOfDay) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", t.Hours, t.Minutes)
}

// PrettyPrint prints a value as JSON for debugging.
func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}
