
package main
 
import (
	"os"
)
 
type Course struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
 
type CurriculumItem struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
}
 
type Exam struct {
	ID         string     `json:"id"`
	CourseID   string     `json:"course_id"`
	Difficulty string     `json:"difficulty"`
	Questions  []Question `json:"questions"`
	CreatedAt  string     `json:"created_at"`
}
 
type Question struct {
	ID         string   `json:"id"`
	Text       string   `json:"text"`
	Options    []string `json:"options"`
	Correct    int      `json:"correct"`
	Difficulty string   `json:"difficulty"`
	Topic      string   `json:"topic"`
}
 
type ExamResult struct {
	ExamID      string           `json:"exam_id"`
	CourseID    string           `json:"course_id"`
	Score       int              `json:"score"`
	Total       int              `json:"total"`
	Percentage  float64          `json:"percentage"`
	Passed      bool             `json:"passed"`
	Review      []QuestionReview `json:"review"`
	SubmittedAt string           `json:"submitted_at"`
}
 
type QuestionReview struct {
	QuestionID        string `json:"question_id"`
	QuestionText      string `json:"question_text"`
	YourAnswer        int    `json:"your_answer"`
	YourAnswerText    string `json:"your_answer_text"`
	CorrectAnswer     int    `json:"correct_answer"`
	CorrectAnswerText string `json:"correct_answer_text"`
	IsCorrect         bool   `json:"is_correct"`
	Explanation       string `json:"explanation"`
}
 
type ChatRequest struct {
	CourseID string `json:"courseId"`
	Message  string `json:"message"`
	Mode     string `json:"mode"`
}
 
type GenerateExamRequest struct {
	CourseID   string `json:"courseId"`
	Difficulty string `json:"difficulty"`
	Count      int    `json:"count"`
}
 
type SubmitExamRequest struct {
	ExamID    string            `json:"exam_id"`
	Questions []Question        `json:"questions"`
	Answers   map[string]string `json:"answers"`
}
 
type ChatResponse struct {
	Response string `json:"response"`
}
 
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
 
type GroqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
 
var CONFIG = struct {
	AI_API_KEY  string
	AI_PROVIDER string
}{}
 
func InitConfig() {
	CONFIG.AI_API_KEY = os.Getenv("GROQ_API_KEY")
	CONFIG.AI_PROVIDER = "groq"
}
