package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	InitConfig()
	InitStorage()
	InitAIService()

	http.HandleFunc("/api/courses", handleCourses)
	http.HandleFunc("/api/curriculum", handleCurriculum)
	http.HandleFunc("/api/chat", handleChat)
	http.HandleFunc("/api/generate-exam", handleGenerateExam)
	http.HandleFunc("/api/submit-exam", handleSubmitExam)

	// Serve static files from current directory (where index.html resides)
	fs := http.FileServer(http.Dir("./Front"))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("🚀 Server running on port " + port)
	http.ListenAndServe(":"+port, nil)
}