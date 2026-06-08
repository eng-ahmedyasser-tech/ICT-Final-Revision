package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type OpenRouterRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

var (
	aiEnabled bool
	randGen   *rand.Rand
)

func InitAIService() {
	aiEnabled = CONFIG.AI_API_KEY != "" && !strings.Contains(CONFIG.AI_API_KEY, "xxxxx")
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateAIResponse(courseID, message, mode, context string) string {
	if aiEnabled && CONFIG.AI_PROVIDER == "openrouter" {
		if resp := callOpenRouter(courseID, message, mode, context); resp != "" {
			return resp
		}
	}
	return fallbackResponse(courseID, message, mode, context)
}

func callOpenRouter(courseID, message, mode, context string) string {
	systemPrompt := fmt.Sprintf(`You are an elite ICT tutor and senior cybersecurity analyst. Your teaching philosophy is "Deep Learning through Practical Application".

Course: %s
Mode: %s

Curriculum Context (Your primary source of truth):
%s

Mandatory Guidelines for every response:
1. DEEP DIVE: Never provide surface-level definitions. Explain the "why" and "how" behind every concept.
2. CONTEXTUAL EXAMPLES: Every technical explanation MUST be accompanied by a concrete, real-world example.
3. STEP-BY-STEP BREAKDOWN: Break complex logic into smaller, manageable parts using numbered steps.
4. PRACTICAL PERSPECTIVE: Relate concepts to real-world system architecture, cybersecurity defense, or practical IT operations.
5. CLARITY & FORMATTING: Use bold headers, bullet points, and code blocks for technical syntax. Ensure the tone is academic, professional, and encouraging.
6. STRICT ADHERENCE: If the concept is not covered in the provided curriculum, explicitly state: "This is beyond the current scope," then provide a brief high-level overview if appropriate.`, courseID, mode, truncateText(context, 3000))

	reqBody := OpenRouterRequest{
		Model: "openai/gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: message},
		},
		MaxTokens: 800,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+CONFIG.AI_API_KEY)
	req.Header.Set("HTTP-Referer", "http://localhost:8000")
	req.Header.Set("X-Title", "ICT Revision Hub")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("OpenRouter error: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("OpenRouter status: %d\n", resp.StatusCode)
		return ""
	}

	var result OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return ""
	}
	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content
	}
	return ""
}

func fallbackResponse(courseID, message, mode, context string) string {
	return fmt.Sprintf("📚 **%s - %s Mode**\n\nBased on the curriculum:\n%s\n\n💡 Tip: Configure OpenRouter API key for enhanced AI responses.", courseID, mode, truncateText(context, 500))
}

func ExtractContext(curriculum []CurriculumItem, message string) string {
	msgLower := strings.ToLower(message)
	var relevant []string

	// Check for topic matches
	for _, item := range curriculum {
		if strings.Contains(msgLower, strings.ToLower(item.Topic)) ||
			strings.Contains(msgLower, strings.ToLower(strings.Split(item.Content, " ")[0])) {
			relevant = append(relevant, fmt.Sprintf("**%s**: %s", item.Topic, item.Content))
		}
	}

	// If no specific matches, return top topics
	if len(relevant) == 0 && len(curriculum) > 0 {
		for i := 0; i < minInt(3, len(curriculum)); i++ {
			relevant = append(relevant, fmt.Sprintf("**%s**: %s", curriculum[i].Topic, curriculum[i].Content))
		}
	}

	return strings.Join(relevant, "\n\n")
}

func GenerateExam(courseID, difficulty string, count int, curriculum []CurriculumItem) Exam {
	exam := Exam{
		ID:         fmt.Sprintf("exam_%d", time.Now().UnixNano()),
		CourseID:   courseID,
		Difficulty: difficulty,
		Questions:  []Question{},
	}

	if len(curriculum) == 0 {
		fmt.Println("Warning: No curriculum available for exam generation")
		return exam
	}

	// Don't limit by topics - allow up to the requested count
	// but ensure we don't exceed available topics * 3
	maxQuestions := len(curriculum) * 3
	if count > maxQuestions {
		fmt.Printf("Warning: Requested %d questions but only %d available. Generating %d questions.\n", count, maxQuestions, maxQuestions)
		count = maxQuestions
	}

	usedQuestions := make(map[string]bool)
	generatedQuestions := 0
	maxAttempts := count * 10 // Increased attempts

	for generatedQuestions < count && len(exam.Questions) < count && maxAttempts > 0 {
		maxAttempts--

		// Cycle through topics systematically instead of randomly
		topicIndex := generatedQuestions % len(curriculum)
		topic := curriculum[topicIndex]

		// Generate a question for this topic
		q := generateMCQ(topic, difficulty, generatedQuestions)

		// Check for duplicate question text
		if usedQuestions[q.Text] {
			// Try a different variation
			q.Text = fmt.Sprintf("%s (Version %d)", strings.TrimSuffix(q.Text, fmt.Sprintf(" (Version %d)", 0)), generatedQuestions+1)
		}

		exam.Questions = append(exam.Questions, q)
		usedQuestions[q.Text] = true
		generatedQuestions++
	}

	fmt.Printf("Generated %d questions for exam (requested: %d)\n", len(exam.Questions), count)
	return exam
}
func generateMCQ(topic CurriculumItem, difficulty string, id int) Question {
	content := topic.Content
	topicLower := strings.ToLower(topic.Topic)

	var questionText string
	var correctOption string
	var wrongOptions []string

	// Generate diverse questions based on topic
	switch {
	case strings.Contains(topicLower, "variable") || strings.Contains(topicLower, "basics"):
		questionText = "What is the primary purpose of a variable in programming?"
		correctOption = "To store and manipulate data in memory with changeable values"
		wrongOptions = []string{
			"To define constant values that never change",
			"To execute loops repeatedly",
			"To declare functions and their parameters",
		}

	case strings.Contains(topicLower, "control flow") || strings.Contains(topicLower, "if-else"):
		questionText = "Which control flow statement allows a program to make decisions based on conditions?"
		correctOption = "if-else statement"
		wrongOptions = []string{
			"for loop",
			"while loop",
			"do-while loop",
		}

	case strings.Contains(topicLower, "function"):
		questionText = "What is the primary benefit of using functions in programming?"
		correctOption = "Code reusability and modular organization"
		wrongOptions = []string{
			"Increased execution speed",
			"Reduced memory usage",
			"Automatic error handling",
		}

	case strings.Contains(topicLower, "array") || strings.Contains(topicLower, "pointer"):
		questionText = "What distinguishes an array from a pointer in C?"
		correctOption = "Arrays allocate contiguous memory blocks; pointers store memory addresses"
		wrongOptions = []string{
			"Arrays can only store integers",
			"Pointers cannot be used with arrays",
			"There is no difference between them",
		}

	case strings.Contains(topicLower, "word processing") || strings.Contains(topicLower, "microsoft word"):
		questionText = "Which feature in Microsoft Word allows creating multiple documents with consistent formatting?"
		correctOption = "Templates"
		wrongOptions = []string{
			"Styles",
			"Macros",
			"Themes",
		}

	case strings.Contains(topicLower, "spreadsheet") || strings.Contains(topicLower, "excel"):
		questionText = "What is the primary function of a Pivot Table in Excel?"
		correctOption = "To summarize and analyze large datasets dynamically"
		wrongOptions = []string{
			"To create charts and graphs",
			"To perform mathematical calculations",
			"To sort and filter data only",
		}

	case strings.Contains(topicLower, "cybersecurity") || strings.Contains(topicLower, "cia"):
		questionText = "What are the three core principles of the CIA triad in cybersecurity?"
		correctOption = "Confidentiality, Integrity, Availability"
		wrongOptions = []string{
			"Confidentiality, Identity, Authentication",
			"Compliance, Investigation, Assessment",
			"Cryptography, Intrusion, Access",
		}

	case strings.Contains(topicLower, "cryptographic") || strings.Contains(topicLower, "encryption"):
		questionText = "What is the key difference between symmetric and asymmetric encryption?"
		correctOption = "Symmetric uses one key; asymmetric uses a public-private key pair"
		wrongOptions = []string{
			"Symmetric is slower but more secure",
			"Asymmetric uses only one key",
			"There is no meaningful difference",
		}

	case strings.Contains(topicLower, "network security") || strings.Contains(topicLower, "firewall"):
		questionText = "What is the primary purpose of a firewall in network security?"
		correctOption = "To monitor and filter incoming/outgoing network traffic based on security rules"
		wrongOptions = []string{
			"To encrypt all network communications",
			"To create backup copies of network data",
			"To manage user passwords",
		}

	default:
		// Generate a question based on the topic content
		questionText = fmt.Sprintf("Regarding '%s', which statement is most accurate?", topic.Topic)
		correctOption = truncateText(content, 120)
		wrongOptions = []string{
			"This describes a different concept in the same domain",
			"This is partially correct but misses key details",
			"This represents a common misunderstanding of the topic",
		}
	}

	// Adjust difficulty - make questions harder based on difficulty level
	if difficulty == "hard" {
		// Make wrong options more convincing
		if len(wrongOptions) >= 3 {
			wrongOptions = []string{
				wrongOptions[0] + " (common misconception)",
				wrongOptions[1],
				"This is true for a different programming concept",
			}
		}
	} else if difficulty == "easy" {
		// Make correct option more obvious
		correctOption = correctOption + " ✓"
	}

	// Shuffle options
	allOptions := append([]string{correctOption}, wrongOptions...)
	randGen.Shuffle(len(allOptions), func(i, j int) {
		allOptions[i], allOptions[j] = allOptions[j], allOptions[i]
	})

	// Find correct index
	correctIndex := 0
	for i, opt := range allOptions {
		// Remove the checkmark for comparison
		optClean := strings.TrimSuffix(opt, " ✓")
		correctClean := strings.TrimSuffix(correctOption, " ✓")
		if optClean == correctClean {
			correctIndex = i
			break
		}
	}

	return Question{
		ID:         fmt.Sprintf("q_%d_%d_%d", time.Now().UnixNano(), randGen.Intn(10000), id),
		Text:       questionText,
		Options:    allOptions,
		Correct:    correctIndex,
		Difficulty: difficulty,
		Topic:      topic.Topic,
	}
}

func GradeExam(questions []Question, answers map[string]string) ExamResult {
	score := 0
	var reviews []QuestionReview

	for idx, q := range questions {
		userAnswer, attempted := answers[fmt.Sprintf("%d", idx)]
		correctText := q.Options[q.Correct]
		// Remove any checkmark from comparison
		correctTextClean := strings.TrimSuffix(correctText, " ✓")
		userAnswerClean := strings.TrimSuffix(userAnswer, " ✓")

		isCorrect := attempted && userAnswerClean == correctTextClean

		if isCorrect {
			score++
		}

		explanation := fmt.Sprintf("✅ Correct! %s", correctTextClean)
		if !isCorrect {
			explanation = fmt.Sprintf("❌ Incorrect. The correct answer is: %s\n\n%s", correctTextClean, getExplanationForTopic(q.Topic))
		}

		reviews = append(reviews, QuestionReview{
			QuestionID:        q.ID,
			QuestionText:      q.Text,
			YourAnswerText:    map[bool]string{true: userAnswer, false: "Not answered"}[attempted],
			CorrectAnswerText: correctTextClean,
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

	// Try exact match
	if explanation, exists := explanations[topic]; exists {
		return explanation
	}

	// Try partial match
	for key, explanation := range explanations {
		if len(key) >= 10 && strings.Contains(strings.ToLower(topic), strings.ToLower(key[:10])) {
			return explanation
		}
	}

	return "Review the course material to strengthen your understanding of this topic."
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
