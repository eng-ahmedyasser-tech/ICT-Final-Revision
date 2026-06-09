package main
 
import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
 
	"github.com/joho/godotenv"
)
 
// ============================================================
// TYPES
// ============================================================
 
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}
 
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}
 
type GeminiPart struct {
	Text string `json:"text"`
}
 
type GeminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
}
 
// QuestionTemplate for JSON bank
type QuestionTemplate struct {
	CourseID string   `json:"course_id"`
	Topic    string   `json:"topic"`
	Question string   `json:"question"`
	Correct  string   `json:"correct"`
	Wrong    []string `json:"wrong"`
}
 
var (
	aiEnabled    bool
	randGen      *rand.Rand
	questionBank []QuestionTemplate
)
 
// ============================================================
// INIT
// ============================================================
 
func InitAIService() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  .env file not found — using system env")
	}
 
	apiKey := os.Getenv("GEMINI_API_KEY")
 
	// DEBUG — هنشيله بعد ما نتأكد
	fmt.Printf("🔍 DEBUG: GEMINI_API_KEY length=%d\n", len(apiKey))
	if len(apiKey) > 4 {
		fmt.Printf("🔍 DEBUG: key starts with: %s...\n", apiKey[:4])
	}
 
	if apiKey != "" {
		CONFIG.AI_API_KEY = apiKey
	}
 
	aiEnabled = CONFIG.AI_API_KEY != "" && !strings.Contains(CONFIG.AI_API_KEY, "xxxxx")
	fmt.Printf("🔍 DEBUG: aiEnabled=%v, provider=%s\n", aiEnabled, CONFIG.AI_PROVIDER)
 
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	loadQuestionBank()
}
 
func loadQuestionBank() {
	data, err := os.ReadFile("questions.json")
	if err != nil {
		fmt.Println("⚠️  questions.json not found — will generate questions dynamically from curriculum")
		return
	}
	if err := json.Unmarshal(data, &questionBank); err != nil {
		fmt.Printf("⚠️  questions.json parse error: %v\n", err)
	}
	fmt.Printf("✅ Loaded %d questions from questions.json\n", len(questionBank))
}
 
// ============================================================
// CHAT
// ============================================================
 
func GenerateAIResponse(courseID, message, mode, context string) string {
	fmt.Printf("🔍 DEBUG: GenerateAIResponse called — aiEnabled=%v, provider=%s\n", aiEnabled, CONFIG.AI_PROVIDER)
 
	if aiEnabled && CONFIG.AI_PROVIDER == "gemini" {
		if resp := callGemini(courseID, message, mode, context); resp != "" {
			return resp
		}
	}
	return fallbackResponse(courseID, message, mode, context)
}
 
func callGemini(courseID, message, mode, context string) string {
	fmt.Println("🔍 DEBUG: Calling Gemini API...")
 
	prompt := fmt.Sprintf(`You are an elite ICT tutor.
Course: %s
Mode: %s
Curriculum Context:
%s
 
Student question: %s
 
Provide a detailed, step-by-step explanation with real-world examples.`,
		courseID, mode, truncateText(context, 3000), message)
 
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: prompt}}},
		},
	}
 
	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s",
		CONFIG.AI_API_KEY,
	)
 
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
 
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Gemini error: %v\n", err)
		return ""
	}
	defer resp.Body.Close()
 
	fmt.Printf("🔍 DEBUG: Gemini HTTP status: %d\n", resp.StatusCode)
 
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Gemini status: %d\n", resp.StatusCode)
		return ""
	}
 
	var result GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Gemini decode error: %v\n", err)
		return ""
	}
 
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		fmt.Println("✅ Gemini response received successfully")
		return result.Candidates[0].Content.Parts[0].Text
	}
 
	fmt.Println("❌ Gemini: empty candidates")
	return ""
}
 
func fallbackResponse(courseID, message, mode, context string) string {
	if context != "" {
		return fmt.Sprintf("📚 Based on the curriculum of %s:\n\n%s\n\n(Note: Configure GEMINI_API_KEY for enhanced AI responses.)", courseID, truncateText(context, 800))
	}
	return fmt.Sprintf("I'm your AI tutor for %s. Please ask a specific question about the course material.", courseID)
}
 
// ============================================================
// CONTEXT EXTRACTION
// ============================================================
 
func ExtractContext(curriculum []CurriculumItem, message string) string {
	msgLower := strings.ToLower(message)
	var relevant []string
	for _, item := range curriculum {
		if strings.Contains(msgLower, strings.ToLower(item.Topic)) ||
			strings.Contains(msgLower, strings.ToLower(strings.Split(item.Content, " ")[0])) {
			relevant = append(relevant, fmt.Sprintf("**%s**: %s", item.Topic, item.Content))
		}
	}
	if len(relevant) == 0 && len(curriculum) > 0 {
		for i := 0; i < minInt(3, len(curriculum)); i++ {
			relevant = append(relevant, fmt.Sprintf("**%s**: %s", curriculum[i].Topic, curriculum[i].Content))
		}
	}
	return strings.Join(relevant, "\n\n")
}
 
// ============================================================
// EXAM GENERATION
// ============================================================
 
func GenerateExam(courseID, difficulty string, count int, curriculum []CurriculumItem) Exam {
	requestedCount := count
	if requestedCount > 30 {
		fmt.Printf("⚠️ Requested %d questions, but maximum allowed is 30. Capping to 30.\n", requestedCount)
		requestedCount = 30
	}
	if requestedCount < 1 {
		requestedCount = 5
	}
 
	var questions []Question
 
	var available []QuestionTemplate
	for _, q := range questionBank {
		if q.CourseID == courseID {
			available = append(available, q)
		}
	}
 
	if len(available) > 0 {
		randGen.Shuffle(len(available), func(i, j int) { available[i], available[j] = available[j], available[i] })
		take := requestedCount
		if take > len(available) {
			take = len(available)
			fmt.Printf("⚠️ Only %d questions available in bank for course %s. Using all of them.\n", len(available), courseID)
		}
		for i := 0; i < take; i++ {
			questions = append(questions, questionFromTemplate(available[i], difficulty, i))
		}
		fmt.Printf("✅ Generated %d random questions from question bank (total bank: %d)\n", len(questions), len(available))
	}
 
	if len(questions) == 0 && len(curriculum) > 0 {
		fmt.Printf("⚠️ No questions in bank for course %s. Generating randomly from curriculum.\n", courseID)
		randGen.Shuffle(len(curriculum), func(i, j int) { curriculum[i], curriculum[j] = curriculum[j], curriculum[i] })
		take := requestedCount
		if take > len(curriculum) {
			take = len(curriculum)
			fmt.Printf("⚠️ Only %d curriculum topics available. Using all.\n", len(curriculum))
		}
		for i := 0; i < take; i++ {
			questions = append(questions, questionFromCurriculum(curriculum[i], difficulty, i))
		}
		fmt.Printf("✅ Generated %d random questions from curriculum\n", len(questions))
	}
 
	if len(questions) == 0 {
		fmt.Printf("⚠️ No curriculum or question bank found. Generating placeholder questions.\n")
		for i := 0; i < requestedCount; i++ {
			questions = append(questions, Question{
				ID:         fmt.Sprintf("placeholder_%d", i),
				Text:       "What is the primary goal of this course?",
				Options:    []string{"Learn ICT fundamentals", "Pass exams", "Get a certificate", "Have fun"},
				Correct:    0,
				Difficulty: difficulty,
				Topic:      "General",
			})
		}
	}
 
	return Exam{
		ID:         fmt.Sprintf("exam_%d", time.Now().UnixNano()),
		CourseID:   courseID,
		Difficulty: difficulty,
		Questions:  questions,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
}
 
func questionFromCurriculum(item CurriculumItem, difficulty string, id int) Question {
	questionText := fmt.Sprintf("Which of the following best describes '%s'?", item.Topic)
	correctAnswer := truncateText(item.Content, 120)
	wrongAnswers := []string{
		"This is a common misconception about the topic",
		"This describes a different concept in the same domain",
		"This is partially correct but missing key details",
	}
	if difficulty == "hard" {
		wrongAnswers = append(wrongAnswers, "This is a trick option closely related but incorrect")
	}
	options := append([]string{correctAnswer}, wrongAnswers...)
	randGen.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	correctIndex := 0
	for idx, opt := range options {
		if opt == correctAnswer {
			correctIndex = idx
			break
		}
	}
	return Question{
		ID:         fmt.Sprintf("q_%d_%d", time.Now().UnixNano(), id),
		Text:       questionText,
		Options:    options,
		Correct:    correctIndex,
		Difficulty: difficulty,
		Topic:      item.Topic,
	}
}
 
func questionFromTemplate(tmpl QuestionTemplate, difficulty string, id int) Question {
	correct := tmpl.Correct
	wrong := make([]string, len(tmpl.Wrong))
	copy(wrong, tmpl.Wrong)
	if difficulty == "easy" {
		correct = correct + " (correct)"
	} else if difficulty == "hard" {
		wrong = append(wrong, "This is a very tricky incorrect option")
	}
	options := append([]string{correct}, wrong...)
	randGen.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	correctIndex := 0
	cleanCorrect := strings.TrimSuffix(correct, " (correct)")
	for idx, opt := range options {
		if strings.TrimSuffix(opt, " (correct)") == cleanCorrect {
			correctIndex = idx
			break
		}
	}
	return Question{
		ID:         fmt.Sprintf("q_%d_%d", time.Now().UnixNano(), id),
		Text:       tmpl.Question,
		Options:    options,
		Correct:    correctIndex,
		Difficulty: difficulty,
		Topic:      tmpl.Topic,
	}
}
 
// ============================================================
// EXAM GRADING
// ============================================================
 
func GradeExam(questions []Question, answers map[string]string) ExamResult {
	score := 0
	var reviews []QuestionReview
	for idx, q := range questions {
		userAnswer, attempted := answers[fmt.Sprintf("%d", idx)]
		correctText := q.Options[q.Correct]
		cleanCorrect := strings.TrimSuffix(correctText, " (correct)")
		cleanCorrect = strings.TrimSuffix(cleanCorrect, " ✓")
		userClean := strings.TrimSuffix(userAnswer, " (correct)")
		userClean = strings.TrimSuffix(userClean, " ✓")
		isCorrect := attempted && userClean == cleanCorrect
		if isCorrect {
			score++
		}
		explanation := fmt.Sprintf("✅ Correct! %s", cleanCorrect)
		if !isCorrect {
			explanation = fmt.Sprintf("❌ Incorrect. The correct answer is: %s\n\n%s", cleanCorrect, getExplanationForTopic(q.Topic))
		}
		reviews = append(reviews, QuestionReview{
			QuestionID:        q.ID,
			QuestionText:      q.Text,
			YourAnswerText:    map[bool]string{true: userAnswer, false: "Not answered"}[attempted],
			CorrectAnswerText: cleanCorrect,
			IsCorrect:         isCorrect,
			Explanation:       explanation,
		})
	}
	percentage := float64(score) / float64(len(questions)) * 100
	return ExamResult{
		Score:      score,
		Total:      len(questions),
		Percentage: percentage,
		Passed:     percentage >= 60,
		Review:     reviews,
	}
}
 
func getExplanationForTopic(topic string) string {
	explanations := map[string]string{
		"C Programming Basics":       "Variables are fundamental to programming as they store data that can be modified during program execution.",
		"Control Flow in C":          "Control flow statements determine the order in which code executes based on conditions.",
		"Functions in C":             "Functions promote code reuse and modular design, making programs easier to maintain.",
		"Arrays and Pointers":        "Arrays store sequences of data while pointers provide memory address manipulation capabilities.",
		"Word Processing":            "Microsoft Word offers various tools for document creation and formatting automation.",
		"Spreadsheet Analysis":       "Excel provides powerful data analysis tools including formulas, pivot tables, and charts.",
		"Document Automation":        "Macros and VBA allow automation of repetitive tasks in Office applications.",
		"Cybersecurity Fundamentals": "The CIA triad forms the foundation of information security practices.",
		"Cryptographic Encryption":   "Encryption protects data confidentiality using mathematical algorithms.",
		"Network Security":           "Network security implements multiple layers of defense to protect data in transit.",
	}
	if exp, ok := explanations[topic]; ok {
		return exp
	}
	for key, exp := range explanations {
		if len(key) >= 10 && strings.Contains(strings.ToLower(topic), strings.ToLower(key[:10])) {
			return exp
		}
	}
	return "Review the course material to strengthen your understanding of this topic."
}
 
// ============================================================
// HELPERS
// ============================================================
 
func truncateText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
 
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
