package main

import (
	"fmt"
	"net/http"
        "os"
)

func main() {
    InitConfig() // استدعِ هذه الدالة هنا أولاً
    InitStorage()
    InitAIService()
    

    // API Routes (أضف "/" قبل المسارات لضمان توافقها)
    http.HandleFunc("/api/courses", handleCourses)
    http.HandleFunc("/api/curriculum", handleCurriculum)
    http.HandleFunc("/api/chat", handleChat)
    http.HandleFunc("/api/generate-exam", handleGenerateExam)
    http.HandleFunc("/api/submit-exam", handleSubmitExam)

    // Serve frontend
    fs := http.FileServer(http.Dir("./Front"))
    http.Handle("/", fs)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Println("🚀 Server running on port " + port)
    http.ListenAndServe(":"+port, nil)
}