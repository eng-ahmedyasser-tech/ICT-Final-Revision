package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleCourses(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	courses, err := GetAllCourses()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func handleCurriculum(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	courseID := r.URL.Query().Get("courseId")
	if courseID == "" {
		courseID = r.URL.Query().Get("course_id")
	}
	
	fmt.Printf("Fetching curriculum for course: %s\n", courseID)
	
	curriculum, err := GetCurriculum(courseID)
	if err != nil {
		fmt.Printf("Error getting curriculum: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CurriculumItem{})
		return
	}
	
	fmt.Printf("Found %d curriculum items\n", len(curriculum))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(curriculum)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	curriculum, _ := GetCurriculum(req.CourseID)
	context := ExtractContext(curriculum, req.Message)
	response := GenerateAIResponse(req.CourseID, req.Message, req.Mode, context)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Response: response})
}

func handleGenerateExam(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	var req GenerateExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Generating exam for course: %s, count: %d, difficulty: %s\n", req.CourseID, req.Count, req.Difficulty)

	curriculum, err := GetCurriculum(req.CourseID)
	if err != nil {
		fmt.Printf("Error getting curriculum: %v\n", err)
		http.Error(w, "Failed to get curriculum", http.StatusInternalServerError)
		return
	}
	
	fmt.Printf("Curriculum has %d topics\n", len(curriculum))
	
	exam := GenerateExam(req.CourseID, req.Difficulty, req.Count, curriculum)
	StoreExam(exam)

	fmt.Printf("Generated %d questions\n", len(exam.Questions))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exam.Questions)
}

func handleSubmitExam(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		return
	}
	var req SubmitExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := GradeExam(req.Questions, req.Answers)
	result.ExamID = req.ExamID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}