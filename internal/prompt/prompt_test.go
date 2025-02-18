package prompt

import (
	"testing"
)

// TestNew tests the creation of a new Prompt object
func TestNew(t *testing.T) {
	p := New("testCommand", "testRest")

	if p.Command != "testCommand" {
		t.Errorf("Expected Command to be 'testCommand', got %s", p.Command)
	}

	if p.Rest != "testRest" {
		t.Errorf("Expected Rest to be 'testRest', got %s", p.Rest)
	}

	if len(p.Questions) != 0 {
		t.Errorf("Expected Questions length to be 0, got %d", len(p.Questions))
	}
}

// TestAsk ensures that questions can be added correctly
func TestAsk(t *testing.T) {
	p := New("testCommand", "testRest")

	q := p.Ask("Are you ready?", []string{"Yes", "No"}, "Yes")

	if q.Question != "Are you ready?" {
		t.Errorf("Expected question text to match, got %s", q.Question)
	}

	if len(q.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(q.Options))
	}

	if q.DefaultResponse != "Yes" {
		t.Errorf("Expected default response to be 'Yes', got %s", q.DefaultResponse)
	}

	if len(p.Questions) != 1 {
		t.Errorf("Expected Questions length to be 1, got %d", len(p.Questions))
	}
}

// TestGetNextQuestion ensures that the next pending question is returned correctly
func TestGetNextQuestion(t *testing.T) {
	p := New("testCommand", "testRest")

	p.Ask("Are you ready?", []string{"Yes", "No"}, "Yes")

	nextQuestion := p.GetNextQuestion()
	if nextQuestion == nil || nextQuestion.Question != "Are you ready?" {
		t.Errorf("Expected next question to be 'Are you ready?', got %v", nextQuestion)
	}

	nextQuestion.Done = true

	if p.GetNextQuestion() != nil {
		t.Errorf("Expected no next question, but got one")
	}
}

// TestAnswer ensures the Answer method works correctly
func TestAnswer(t *testing.T) {
	q := &Question{
		Question:        "What's your favorite color?",
		Options:         []string{"Red", "Blue", "Green"},
		DefaultResponse: "Red",
	}

	q.Answer("Blue")

	if q.Response != "Blue" {
		t.Errorf("Expected response to be 'Blue', got %s", q.Response)
	}

	if !q.Done {
		t.Errorf("Expected Done to be true, got false")
	}

	q.Answer("")
	if q.Response != "Red" {
		t.Errorf("Expected response to default to 'Red', got %s", q.Response)
	}
}

// TestRejectResponse ensures the RejectResponse method works correctly
func TestRejectResponse(t *testing.T) {
	q := &Question{
		Question: "What's your favorite color?",
		Response: "Red",
		Done:     true,
	}

	q.RejectResponse()

	if q.Response != "" {
		t.Errorf("Expected Response to be empty, got %s", q.Response)
	}

	if q.Done {
		t.Errorf("Expected Done to be false, got true")
	}
}

// TestReset ensures the Reset method works correctly
func TestReset(t *testing.T) {
	q := &Question{
		Done: true,
	}

	q.Reset()

	if q.Done {
		t.Errorf("Expected Done to be false, got true")
	}
}
