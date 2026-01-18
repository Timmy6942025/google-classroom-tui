package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// mockServer creates a mock Classroom API server.
func mockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/courses":
			courses := []*Course{
				{ID: "123", Name: "Test Course", Section: "A"},
				{ID: "456", Name: "Another Course", Section: "B"},
			}
			response := ListCoursesResponse{
				Courses: courses,
			}
			json.NewEncoder(w).Encode(response)
		case "/courses/123":
			course := &Course{ID: "123", Name: "Test Course", Section: "A"}
			json.NewEncoder(w).Encode(course)
		case "/courses/123/courseWork":
			coursework := []*CourseWork{
				{ID: "cw1", CourseID: "123", Title: "Assignment 1", WorkType: "ASSIGNMENT", MaxPoints: 100},
			}
			response := ListCourseWorkResponse{
				CourseWork: coursework,
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
}

// mockTokenSource creates a mock token source.
type mockTokenSource struct {
	token *oauth2.Token
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, nil
}

// TestNewClient tests creating a new API client.
func TestNewClient(t *testing.T) {
	server := mockServer()
	defer server.Close()

	token := &oauth2.Token{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	ts := &mockTokenSource{token: token}
	cfg := &Configuration{
		RateLimitBackoff: time.Second,
		MaxRetries:       3,
	}

	client, err := NewClient(context.Background(), ts, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}
}

// TestListCourses tests listing courses.
func TestListCourses(t *testing.T) {
	server := mockServer()
	defer server.Close()

	token := &oauth2.Token{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	ts := &mockTokenSource{token: token}
	cfg := &Configuration{}

	client, err := NewClient(context.Background(), ts, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	courses, err := client.ListCourses(context.Background())
	if err != nil {
		t.Fatalf("Failed to list courses: %v", err)
	}

	if len(courses) != 2 {
		t.Errorf("Expected 2 courses, got %d", len(courses))
	}
}

// TestGetCourse tests getting a single course.
func TestGetCourse(t *testing.T) {
	server := mockServer()
	defer server.Close()

	token := &oauth2.Token{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	ts := &mockTokenSource{token: token}
	cfg := &Configuration{}

	client, err := NewClient(context.Background(), ts, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	course, err := client.GetCourse(context.Background(), "123")
	if err != nil {
		t.Fatalf("Failed to get course: %v", err)
	}

	if course.ID != "123" {
		t.Errorf("Expected course ID '123', got '%s'", course.ID)
	}

	if course.Name != "Test Course" {
		t.Errorf("Expected course name 'Test Course', got '%s'", course.Name)
	}
}

// TestListCourseWork tests listing coursework.
func TestListCourseWork(t *testing.T) {
	server := mockServer()
	defer server.Close()

	token := &oauth2.Token{
		AccessToken:  "test_token",
		RefreshToken: "test_refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	ts := &mockTokenSource{token: token}
	cfg := &Configuration{}

	client, err := NewClient(context.Background(), ts, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	coursework, err := client.ListCourseWork(context.Background(), "123")
	if err != nil {
		t.Fatalf("Failed to list coursework: %v", err)
	}

	if len(coursework) != 1 {
		t.Errorf("Expected 1 coursework item, got %d", len(coursework))
	}

	if coursework[0].ID != "cw1" {
		t.Errorf("Expected coursework ID 'cw1', got '%s'", coursework[0].ID)
	}
}

// TestConvertCourse tests course conversion.
func TestConvertCourse(t *testing.T) {
	// This would test the internal conversion functions
	// For now, we'll skip detailed tests as they require mocking the actual API types
	t.Skip("Requires detailed API type mocking")
}
