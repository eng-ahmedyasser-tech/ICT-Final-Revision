package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// تعريف QuestionTemplate هنا مؤقتاً (سيكون في models.go)
//但如果已经在了 models.go 中定义，请确保导入正确

var (
	aiEnabled    bool
	randGen      *rand.Rand
	questionBank []QuestionTemplate
	
	// Rate limiting variables
	lastGroqCall    time.Time
	groqMutex       sync.Mutex
	groqCallCount   int
	groqWindowStart time.Time
)

func InitAIService() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  .env file not found — using system env")
	}

	apiKey := os.Getenv("GROQ_API_KEY")

	fmt.Printf("🔍 DEBUG: GROQ_API_KEY length=%d\n", len(apiKey))
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
	
	// إحصائيات عن مستويات الصعوبة
	stats := make(map[string]int)
	for _, q := range questionBank {
		if q.Difficulty == "" {
			stats["unknown"]++
		} else {
			stats[q.Difficulty]++
		}
	}
	fmt.Printf("✅ Loaded %d questions from questions.json\n", len(questionBank))
	fmt.Printf("   Difficulty breakdown: Easy=%d, Medium=%d, Hard=%d, Unknown=%d\n", 
		stats["easy"], stats["medium"], stats["hard"], stats["unknown"])
}

func callGroqWithRateLimit(courseID, message, mode, context string) string {
	groqMutex.Lock()
	defer groqMutex.Unlock()

	now := time.Now()

	// Reset counter every minute
	if groqWindowStart.IsZero() || now.Sub(groqWindowStart) > 60*time.Second {
		groqWindowStart = now
		groqCallCount = 0
	}

	// Check if we exceeded 20 requests per minute (leaving buffer for 30 limit)
	if groqCallCount >= 20 {
		waitTime := 60*time.Second - now.Sub(groqWindowStart)
		if waitTime > 0 {
			fmt.Printf("⏳ Rate limit: %d requests in last minute. Waiting %v...\n", groqCallCount, waitTime)
			time.Sleep(waitTime)
		}
		// Reset after waiting
		groqWindowStart = time.Now()
		groqCallCount = 0
	}

	// Minimum 2 seconds between calls
	if !lastGroqCall.IsZero() && now.Sub(lastGroqCall) < 2*time.Second {
		sleepTime := 2*time.Second - now.Sub(lastGroqCall)
		fmt.Printf("⏳ Too fast! Waiting %v between calls...\n", sleepTime)
		time.Sleep(sleepTime)
	}

	lastGroqCall = time.Now()
	groqCallCount++

	fmt.Printf("🔍 Groq call #%d in current minute\n", groqCallCount)

	return callGroq(courseID, message, mode, context)
}

func callGroq(courseID, message, mode, context string) string {
	fmt.Println("🔍 DEBUG: Calling Groq API...")

	systemPrompt := fmt.Sprintf(`You are an elite ICT tutor.
Course: %s
Mode: %s
Curriculum Context:
%s

Provide a detailed, step-by-step explanation with real-world examples.`,
		courseID, mode, truncateText(context, 3000))

	reqBody := map[string]any{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": message},
		},
		"max_tokens": 800,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+CONFIG.AI_API_KEY)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Groq error: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	fmt.Printf("🔍 DEBUG: Groq HTTP status: %d\n", resp.StatusCode)

	if resp.StatusCode == 429 {
		fmt.Printf("❌ Rate limit exceeded (429). Consider increasing delay or upgrading plan.\n")
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Groq status: %d\n", resp.StatusCode)
		return ""
	}

	var result GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Groq decode error: %v\n", err)
		return ""
	}

	if len(result.Choices) > 0 {
		fmt.Println("✅ Groq response received successfully")
		return result.Choices[0].Message.Content
	}

	fmt.Println("❌ Groq: empty response")
	return ""
}

func GenerateAIResponse(courseID, message, mode, context string) string {
	fmt.Printf("🔍 DEBUG: GenerateAIResponse called — aiEnabled=%v, provider=%s\n", aiEnabled, CONFIG.AI_PROVIDER)

	if aiEnabled && CONFIG.AI_PROVIDER == "groq" {
		if resp := callGroqWithRateLimit(courseID, message, mode, context); resp != "" {
			return resp
		}
	}
	return fallbackResponse(courseID, message, mode, context)
}

func fallbackResponse(courseID, message, mode, context string) string {
	if context != "" {
		return fmt.Sprintf("📚 Based on the curriculum of %s:\n\n%s\n\n(Note: Configure GROQ_API_KEY for enhanced AI responses.)", courseID, truncateText(context, 800))
	}
	return fmt.Sprintf("I'm your AI tutor for %s. Please ask a specific question about the course material.", courseID)
}

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

// ============================================
// لوجيك مستوى الصعوبة الجديد
// ============================================

func normalizeDifficulty(difficulty string) string {
	d := strings.ToLower(difficulty)
	if d == "easy" || d == "medium" || d == "hard" {
		return d
	}
	return "medium"
}

// الدالة الرئيسية لتوليد الامتحان مع لوجيك صعوبة حقيقي
func GenerateExam(courseID, difficulty string, count int, curriculum []CurriculumItem) Exam {
	requestedCount := count
	if requestedCount > 30 {
		fmt.Printf("⚠️ Requested %d questions, capping to 30\n", requestedCount)
		requestedCount = 30
	}
	if requestedCount < 1 {
		requestedCount = 5
	}

	// تطبيع مستوى الصعوبة
	difficulty = normalizeDifficulty(difficulty)
	fmt.Printf("🎯 Generating exam for %s | Difficulty: %s | Count: %d\n", 
		courseID, difficulty, requestedCount)

	var questions []Question

	// 1. محاولة جلب أسئلة من بنك الأسئلة بنفس مستوى الصعوبة
	var available []QuestionTemplate
	for _, q := range questionBank {
		if q.CourseID == courseID && q.Difficulty == difficulty {
			available = append(available, q)
		}
	}

	fmt.Printf("📚 Found %d questions in bank with difficulty '%s' for course %s\n", 
		len(available), difficulty, courseID)

	if len(available) >= requestedCount {
		// عشوائية كافية
		randGen.Shuffle(len(available), func(i, j int) { available[i], available[j] = available[j], available[i] })
		for i := 0; i < requestedCount; i++ {
			questions = append(questions, questionFromTemplateClean(available[i]))
		}
		fmt.Printf("✅ Generated %d questions from bank (exact difficulty match)\n", len(questions))
	} else if len(available) > 0 {
		// أسئلة أقل من المطلوب - نأخذ المتاح
		for _, q := range available {
			questions = append(questions, questionFromTemplateClean(q))
		}
		fmt.Printf("⚠️ Only %d questions available for '%s' difficulty. Need %d more.\n", 
			len(available), difficulty, requestedCount-len(available))
		
		// نكمل من مستويات صعوبة مجاورة أو من المنهج
		remaining := requestedCount - len(questions)
		questions = append(questions, generateFallbackQuestions(courseID, difficulty, remaining, curriculum)...)
	} else {
		// لا توجد أسئلة على الإطلاق - نولد من المنهج
		fmt.Printf("⚠️ No questions in bank for %s with difficulty '%s'. Generating from curriculum.\n", 
			courseID, difficulty)
		questions = generateFallbackQuestions(courseID, difficulty, requestedCount, curriculum)
	}

	// ضمان عدم وجود أسئلة فارغة
	if len(questions) == 0 {
		fmt.Printf("❌ CRITICAL: No questions generated! Using placeholder questions.\n")
		questions = generatePlaceholderQuestions(courseID, requestedCount)
	}

	return Exam{
		ID:         fmt.Sprintf("exam_%d", time.Now().UnixNano()),
		CourseID:   courseID,
		Difficulty: difficulty,
		Questions:  questions,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
}

// تحويل قالب السؤال إلى سؤال (بدون تلاعب نصي في الإجابات)
func questionFromTemplateClean(tmpl QuestionTemplate) Question {
	correct := tmpl.Correct
	wrong := make([]string, len(tmpl.Wrong))
	copy(wrong, tmpl.Wrong)
	
	// خلط الخيارات
	options := append([]string{correct}, wrong...)
	randGen.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	
	// إيجاد الفهرس الصحيح بعد الخلط
	correctIndex := 0
	for idx, opt := range options {
		if opt == correct {
			correctIndex = idx
			break
		}
	}
	
	difficulty := tmpl.Difficulty
	if difficulty == "" {
		difficulty = "medium"
	}
	
	return Question{
		ID:         fmt.Sprintf("q_%d_%d", time.Now().UnixNano(), randGen.Intn(10000)),
		Text:       tmpl.Question,
		Options:    options,
		Correct:    correctIndex,
		Difficulty: difficulty,
		Topic:      tmpl.Topic,
	}
}

// توليد أسئلة احتياطية من المنهج مع تحكم بمستوى الصعوبة
func generateFallbackQuestions(courseID, targetDifficulty string, count int, curriculum []CurriculumItem) []Question {
	var questions []Question
	
	if len(curriculum) == 0 {
		return questions
	}
	
	// خلط المنهج
	randGen.Shuffle(len(curriculum), func(i, j int) { curriculum[i], curriculum[j] = curriculum[j], curriculum[i] })
	
	for i := 0; i < count && i < len(curriculum); i++ {
		questions = append(questions, questionFromCurriculumWithDifficulty(curriculum[i], targetDifficulty, i))
	}
	
	fmt.Printf("🔄 Generated %d fallback questions from curriculum (difficulty: %s)\n", 
		len(questions), targetDifficulty)
	
	return questions
}

// توليد سؤال من المنهج بمستوى صعوبة محدد
func questionFromCurriculumWithDifficulty(item CurriculumItem, difficulty string, id int) Question {
	var questionText, correctAnswer string
	var wrongAnswers []string
	
	content := item.Content
	topic := item.Topic
	
	switch difficulty {
	case "easy":
		questionText = fmt.Sprintf("What is the main concept of '%s'?", topic)
		correctAnswer = truncateText(content, 100)
		wrongAnswers = []string{
			"This is unrelated to the topic",
			"A common misunderstanding of this concept",
			"This describes a different subject entirely",
		}
		
	case "medium":
		questionText = fmt.Sprintf("Which of the following best describes the key principles of '%s'?", topic)
		correctAnswer = truncateText(content, 150)
		wrongAnswers = []string{
			"A simplified but incorrect explanation",
			"A concept from a different field",
			"An outdated or incomplete definition",
		}
		
	case "hard":
		questionText = fmt.Sprintf("Analyze and explain the following about '%s': %s\n\nProvide a detailed technical explanation.", 
			topic, truncateText(content, 80))
		correctAnswer = truncateText(content, 200) + " This represents the complete technical understanding of the concept."
		wrongAnswers = []string{
			"A surface-level understanding missing key details",
			"A common misconception that contradicts the material",
			"An unrelated technical concept",
			"An oversimplified version that loses important nuance",
		}
		
	default:
		questionText = fmt.Sprintf("Which of the following best describes '%s'?", topic)
		correctAnswer = truncateText(content, 120)
		wrongAnswers = []string{"Incorrect definition", "Related but wrong", "Common error"}
	}
	
	options := append([]string{correctAnswer}, wrongAnswers...)
	randGen.Shuffle(len(options), func(i, int2 int) { options[i], options[int2] = options[int2], options[i] })
	
	correctIndex := 0
	for idx, opt := range options {
		if opt == correctAnswer {
			correctIndex = idx
			break
		}
	}
	
	return Question{
		ID:         fmt.Sprintf("fallback_%d_%d", time.Now().UnixNano(), id),
		Text:       questionText,
		Options:    options,
		Correct:    correctIndex,
		Difficulty: difficulty,
		Topic:      topic,
	}
}

// أسئلة احتياطية نهائية
func generatePlaceholderQuestions(courseID string, count int) []Question {
	var questions []Question
	for i := 0; i < count; i++ {
		questions = append(questions, Question{
			ID:         fmt.Sprintf("ph_%d", i),
			Text:       "What is the most important concept in " + courseID + "?",
			Options:    []string{"Understanding fundamentals", "Memorizing facts", "Practical application", "Theory only"},
			Correct:    0,
			Difficulty: "medium",
			Topic:      "General",
		})
	}
	return questions
}

// ============================================
// تقييم الامتحان
// ============================================

func GradeExam(questions []Question, answers map[string]string) ExamResult {
	score := 0
	var reviews []QuestionReview
	for idx, q := range questions {
		userAnswer, attempted := answers[fmt.Sprintf("%d", idx)]
		correctText := q.Options[q.Correct]
		userClean := strings.TrimSuffix(userAnswer, " (correct)")
		userClean = strings.TrimSuffix(userClean, " ✓")
		correctClean := strings.TrimSuffix(correctText, " (correct)")
		correctClean = strings.TrimSuffix(correctClean, " ✓")
		
		isCorrect := attempted && userClean == correctClean
		if isCorrect {
			score++
		}
		explanation := getExplanationForTopicWithDifficulty(q.Topic, q.Difficulty)
		if !isCorrect {
			explanation = fmt.Sprintf("❌ Incorrect. The correct answer is: %s\n\n%s", correctClean, explanation)
		} else {
			explanation = fmt.Sprintf("✅ Correct! %s", explanation)
		}
		reviews = append(reviews, QuestionReview{
			QuestionID:        q.ID,
			QuestionText:      q.Text,
			YourAnswerText:    map[bool]string{true: userAnswer, false: "Not answered"}[attempted],
			CorrectAnswerText: correctClean,
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

func getExplanationForTopicWithDifficulty(topic string, difficulty string) string {
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
		"Caesar Cipher":              "The Caesar cipher shifts each letter by a fixed number. Encryption: E(x) = (x + n) mod 26, Decryption: D(x) = (x - n) mod 26.",
		"Vigenère Cipher":            "A polyalphabetic cipher using a repeating keyword. More secure than Caesar because different letters get shifted differently.",
	}
	
	baseExp := explanations[topic]
	if baseExp == "" {
		// البحث عن تطابق جزئي
		for key, exp := range explanations {
			if len(key) >= 10 && strings.Contains(strings.ToLower(topic), strings.ToLower(key[:10])) {
				baseExp = exp
				break
			}
		}
	}
	if baseExp == "" {
		baseExp = "Review the course material to strengthen your understanding of this topic."
	}
	
	// إضافة تفاصيل إضافية حسب مستوى الصعوبة
	if difficulty == "hard" {
		baseExp += "\n\n🔍 Advanced: Consider how this concept applies in real-world scenarios and its limitations."
	} else if difficulty == "easy" {
		baseExp += "\n\n💡 Tip: Focus on remembering the key definition for exam success."
	}
	
	return baseExp
}

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