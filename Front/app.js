const CONFIG = {
  API_BASE: "http://localhost:8000",
  DEFAULT_MODE: "home",
  DEFAULT_QUESTION_COUNT: 10,
  DEFAULT_DIFFICULTY: "easy"
};

// Theme Management
function initTheme() {
  const savedTheme = localStorage.getItem('theme') || 'dark';
  document.documentElement.setAttribute('data-theme', savedTheme);
}

function toggleTheme() {
  const currentTheme = document.documentElement.getAttribute('data-theme');
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', newTheme);
  localStorage.setItem('theme', newTheme);
  const themeBtn = document.getElementById("themeToggleBtn");
  if (themeBtn) {
    themeBtn.innerHTML = newTheme === 'dark' ? '☀️' : '🌙';
  }
}

// Enhanced State Management
const state = {
  mode: "home",
  course: null,
  exam: { questions: [], answers: {}, score: null, review: [] },
  view: "home",
  analysis: { loading: false, content: null, currentTopic: null },
  loading: false,
  error: null
};

let currentExamConfig = { count: CONFIG.DEFAULT_QUESTION_COUNT, difficulty: CONFIG.DEFAULT_DIFFICULTY };
let timerInterval = null;
let secondsRemaining = 0;
let currentQuestionIndex = 0;
let activeSelectedTopicContent = null;
let activeSelectedTopicName = null;

window.addEventListener("DOMContentLoaded", () => {
  initTheme();
  render();
  attachGlobalChat();
});

function setView(view, data = {}) {
  state.view = view;
  state.error = null;
  
  if (view === "workspace" && data.course) {
    state.course = data.course;
    activeSelectedTopicContent = null;
    activeSelectedTopicName = null;
    state.analysis.content = null;
    state.analysis.currentTopic = null;
  }
  if (view === "exam-config" && data.course) state.course = data.course;
  if (view === "exam-run") {
    state.exam.answers = {};
    state.exam.questions = [];
    state.exam.score = null;
    state.exam.review = [];
    currentQuestionIndex = 0;
    if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
  }
  if (view === "home") {
    state.mode = "home";
    state.course = null;
    state.exam = { questions: [], answers: {}, score: null, review: [] };
    if (timerInterval) { clearInterval(timerInterval); timerInterval = null; }
  }
  render();
}

function showLoading(message = "Loading...") {
  state.loading = true;
  render();
}

function hideLoading() {
  state.loading = false;
  render();
}

function showError(message) {
  state.error = message;
  render();
  setTimeout(() => {
    if (state.error === message) {
      state.error = null;
      render();
    }
  }, 5000);
}

function escapeHtml(text) {
  if (!text) return '';
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

async function render() {
  const root = document.getElementById("root");
  if (!root) return;

  const header = `
    <div class="app-header">
      <div class="header-left">
        <div class="logo" onclick="window.setView && setView('home')">📘 ICT Revision Hub</div>
        <div class="nav-buttons">
          <button class="nav-btn ${state.view === "home" ? "active" : ""}" data-nav="home">🏠 Home</button>
          <button class="nav-btn ${state.view === "courses" ? "active" : ""}" data-nav="courses">📚 Courses</button>
        </div>
      </div>
      <div class="header-right">
        ${state.course ? `<span class="current-course">📖 ${escapeHtml(state.course.name)}</span>` : ''}
        <button class="theme-toggle" id="themeToggleBtn">${document.documentElement.getAttribute('data-theme') === 'dark' ? '☀️' : '🌙'}</button>
      </div>
    </div>
  `;

  let mainContent = "";
  
  if (state.loading) {
    mainContent = '<div class="loading-overlay"><div class="loading-spinner-large"></div><p>Loading...</p></div>';
  } else if (state.error) {
    mainContent = `<div class="error-message">⚠️ ${escapeHtml(state.error)}</div>`;
  } else if (state.view === "home") mainContent = renderHome();
  else if (state.view === "courses") mainContent = await renderCourses();
  else if (state.view === "workspace") mainContent = await renderWorkspace();
  else if (state.view === "exam-config") mainContent = renderExamConfig();
  else if (state.view === "exam-run") mainContent = await renderExamRun();
  else if (state.view === "results") mainContent = renderResults();

  root.innerHTML = header + `<div class="container">${mainContent}</div>`;

  document.getElementById("themeToggleBtn")?.addEventListener("click", toggleTheme);
  attachNavEvents();
  attachHomeEvents();
  attachCourseEvents();
  attachExamEvents();
  attachTopicEvents();
}

function renderHome() {
  return `
    <div class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">Welcome to ICT Revision Hub</h1>
        <p class="hero-subtitle">Your comprehensive platform for ICT exam preparation</p>
        <div class="hero-buttons">
          <button class="btn-primary btn-large" data-mode="study">📖 Study Mode</button>
          <button class="btn-secondary btn-large" data-mode="exam">📝 Exam Mode</button>
        </div>
      </div>
      <div class="features-grid">
        <div class="feature-card">
          <div class="feature-icon">🎯</div>
          <h3>Personalized Learning</h3>
          <p>AI-powered explanations tailored to your course</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">📊</div>
          <h3>Practice Exams</h3>
          <p>Generate custom exams with varying difficulty</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">💬</div>
          <h3>AI Assistant</h3>
          <p>24/7 intelligent tutoring support</p>
        </div>
      </div>
    </div>
  `;
}

async function renderCourses() {
  try {
    const courses = await window.api.getCourses();
    if (!courses || courses.length === 0) {
      return '<div class="empty-state">No courses available</div>';
    }
    
    return `
      <div class="page-header">
        <h2>Available Courses</h2>
        <p class="mode-badge">${state.mode.toUpperCase()} Mode</p>
      </div>
      <div class="cards-grid">
        ${courses.map(c => `
          <div class="course-card" data-course-id="${c.id}" data-course-name="${escapeHtml(c.name)}">
            <div class="course-icon">📘</div>
            <div class="course-info">
              <h3>${escapeHtml(c.id)}: ${escapeHtml(c.name)}</h3>
              <p>${escapeHtml(c.description) || 'Comprehensive course covering essential topics'}</p>
              <button class="btn-primary btn-small">${state.mode === 'exam' ? 'Take Exam' : 'Start Studying'}</button>
            </div>
          </div>
        `).join("")}
      </div>
    `;
  } catch (error) {
    console.error("Error loading courses:", error);
    showError("Failed to load courses. Please check your connection.");
    return '<div class="error-message">Failed to load courses</div>';
  }
}

async function renderWorkspace() {
  try {
    const curriculum = await window.api.getCurriculum(state.course.id);
    
    // Debug: Log what we received
    console.log("Curriculum received:", curriculum);
    
    // Check if curriculum is valid
    if (!curriculum || !Array.isArray(curriculum) || curriculum.length === 0) {
      return `
        <div class="workspace-header">
          <h2>${escapeHtml(state.course.name)}</h2>
          <button class="btn-secondary" onclick="setView('courses')">← Change Course</button>
        </div>
        <div class="empty-state">
          <p>No curriculum data available for this course.</p>
          <p>Please check back later or select a different course.</p>
        </div>
      `;
    }
    
    let analysisContent = '';
    
    if (state.analysis.loading) {
      analysisContent = '<div class="loading-spinner">🤖 AI is analyzing this topic...</div>';
    } else if (state.analysis.content) {
      let formatted = state.analysis.content.replace(/\n/g, '<br>').replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
      analysisContent = `
        <div class="analysis-result">
          <div class="analysis-header">
            <h4>🎯 AI Analysis: ${escapeHtml(state.analysis.currentTopic || 'Topic')}</h4>
          </div>
          <div class="analysis-body">${formatted}</div>
        </div>
      `;
    } else {
      analysisContent = '<div class="placeholder-text">💡 Click any topic to get AI-powered analysis and explanations</div>';
    }

    const activeNotesBody = activeSelectedTopicContent && activeSelectedTopicName
      ? `
        <div class="topic-content">
          <div class="topic-header">
            <h3>📖 ${escapeHtml(activeSelectedTopicName)}</h3>
          </div>
          <div class="content-body">${escapeHtml(activeSelectedTopicContent).replace(/\n/g, '<br>')}</div>
        </div>
      `
      : `
        <div class="welcome-message">
          <h4>Welcome to ${escapeHtml(state.course.name)}</h4>
          <p>Select a topic from the syllabus to begin learning</p>
        </div>
      `;

    return `
      <div class="workspace-header">
        <h2>${escapeHtml(state.course.name)}</h2>
        <button class="btn-secondary" onclick="setView('courses')">← Change Course</button>
      </div>
      <div class="workspace">
        <div class="panel syllabus-panel">
          <h3>📚 Syllabus Topics</h3>
          <ul class="topic-list">
            ${curriculum.map(item => {
              // Handle both possible field names
              const topicName = item.topic || item.Topic || "Untitled Topic";
              const topicContent = item.content || item.Content || "No content available";
              const isActive = activeSelectedTopicName === topicName ? "active" : "";
              return `<li class="topic-item ${isActive}" data-topic="${escapeHtml(topicName)}" data-content="${encodeURIComponent(topicContent)}">
                <span class="topic-icon">📄</span>
                <span class="topic-name">${escapeHtml(topicName)}</span>
              </li>`;
            }).join("")}
          </ul>
        </div>
        <div class="panel notes-panel">
          <h3>📝 Study Notes</h3>
          <div id="notesContent">${activeNotesBody}</div>
        </div>
        <div class="panel analysis-panel">
          <h3>🤖 AI Tutor</h3>
          <div id="analysisContent">${analysisContent}</div>
        </div>
      </div>
    `;
  } catch (error) {
    console.error("Error rendering workspace:", error);
    showError("Failed to load course content. Please try again.");
    return `
      <div class="workspace-header">
        <h2>${state.course ? escapeHtml(state.course.name) : 'Course'}</h2>
        <button class="btn-secondary" onclick="setView('courses')">← Change Course</button>
      </div>
      <div class="error-message">
        Failed to load curriculum data. Please check your connection and try again.
      </div>
    `;
  }
}

function renderExamConfig() {
  return `
    <div class="exam-config-container">
      <div class="config-header">
        <h2>Exam Setup</h2>
        <p>${escapeHtml(state.course.name)}</p>
      </div>
      <div class="config-card">
        <div class="config-group">
          <label>📊 Number of Questions</label>
          <div class="button-group">
    <button class="config-btn ${currentExamConfig.count === 5 ? 'active' : ''}" data-count="5">5</button>
    <button class="config-btn ${currentExamConfig.count === 10 ? 'active' : ''}" data-count="10">10</button>
    <button class="config-btn ${currentExamConfig.count === 20 ? 'active' : ''}" data-count="20">20</button>
    <button class="config-btn ${currentExamConfig.count === 30 ? 'active' : ''}" data-count="30">30</button>

          </div>
        
        <div class="config-group">
          <label>⚡ Difficulty Level</label>
          <div class="difficulty-selector">
            <label class="difficulty-option ${currentExamConfig.difficulty === 'easy' ? 'selected' : ''}">
              <input type="radio" name="difficulty" value="easy" ${currentExamConfig.difficulty === "easy" ? "checked" : ""}>
              <span>Easy</span>
            </label>
            <label class="difficulty-option ${currentExamConfig.difficulty === 'medium' ? 'selected' : ''}">
              <input type="radio" name="difficulty" value="medium" ${currentExamConfig.difficulty === "medium" ? "checked" : ""}>
              <span>Medium</span>
            </label>
            <label class="difficulty-option ${currentExamConfig.difficulty === 'hard' ? 'selected' : ''}">
              <input type="radio" name="difficulty" value="hard" ${currentExamConfig.difficulty === "hard" ? "checked" : ""}>
              <span>Hard</span>
            </label>
          </div>
        </div>
        
        <button id="startExamBtn" class="btn-primary btn-large">🚀 Start Exam</button>
      </div>
    </div>
  `;
}

async function renderExamRun() {
  if (state.exam.questions.length === 0) {
    showLoading("Generating your exam...");
    try {
      const questions = await window.api.generateExam(state.course.id, currentExamConfig.count, currentExamConfig.difficulty);
      if (!questions || questions.length === 0) {
        throw new Error("Failed to generate questions");
      }
      state.exam.questions = questions;
      secondsRemaining = state.exam.questions.length * 60;
      
      if (timerInterval) clearInterval(timerInterval);
      timerInterval = setInterval(() => {
        if (secondsRemaining <= 0) {
          clearInterval(timerInterval);
          triggerExamSubmission();
        } else {
          secondsRemaining--;
          const timerEl = document.getElementById("examTimer");
          if (timerEl) timerEl.innerText = formatTime(secondsRemaining);
          
          // Warning when 5 minutes left
          if (secondsRemaining === 300) {
            showError("⚠️ 5 minutes remaining!");
          }
        }
      }, 1000);
    } catch (error) {
      console.error("Error generating exam:", error);
      showError("Failed to generate exam. Please try again.");
      setView("exam-config", { course: state.course });
      return "";
    } finally {
      hideLoading();
    }
  }

  if (!state.exam.questions || state.exam.questions.length === 0) {
    return '<div class="error-message">No questions available</div>';
  }
  
  const q = state.exam.questions[currentQuestionIndex];
  if (!q) return '<div class="error-message">Question not found</div>';
  
  const savedAnswer = state.exam.answers[currentQuestionIndex] || "";
  const isLastQuestion = currentQuestionIndex === state.exam.questions.length - 1;
  const progress = ((currentQuestionIndex + 1) / state.exam.questions.length) * 100;

  return `
    <div class="exam-container">
      <div class="exam-header">
        <div id="examTimer" class="timer">${formatTime(secondsRemaining)}</div>
        <div class="progress-bar">
          <div class="progress-fill" style="width: ${progress}%"></div>
        </div>
        <div class="question-counter">Question ${currentQuestionIndex + 1} of ${state.exam.questions.length}</div>
      </div>
      
      <div class="question-card">
        <h3 class="question-text">${escapeHtml(q.text)}</h3>
        <div class="options">
          ${q.options.map((opt, idx) => {
            const isSelected = savedAnswer === opt ? "selected" : "";
            return `
              <label class="option ${isSelected}">
                <input type="radio" name="question" value="${escapeHtml(opt)}" ${savedAnswer === opt ? "checked" : ""}>
                <span class="option-letter">${String.fromCharCode(65+idx)}.</span>
                <span class="option-text">${escapeHtml(opt)}</span>
              </label>
            `;
          }).join("")}
        </div>
        
        <div class="exam-navigation">
          <button id="prevQ" class="btn-secondary" ${currentQuestionIndex === 0 ? "disabled" : ""}>← Previous</button>
          ${isLastQuestion ? 
            `<button id="submitExamBtn" class="btn-success">✅ Submit Exam</button>` : 
            `<button id="nextQ" class="btn-primary">Next →</button>`
          }
        </div>
      </div>
    </div>
  `;
}

function renderResults() {
  if (!state.exam.questions || state.exam.questions.length === 0) {
    return '<div class="error-message">No exam data available</div>';
  }
  
  const percentage = (state.exam.score / state.exam.questions.length) * 100;
  let gradeClass = '';
  let gradeMessage = '';
  
  if (percentage >= 80) {
    gradeClass = 'grade-excellent';
    gradeMessage = 'Excellent! Outstanding performance! 🎉';
  } else if (percentage >= 60) {
    gradeClass = 'grade-good';
    gradeMessage = 'Good job! You passed! 👍';
  } else {
    gradeClass = 'grade-needs-improvement';
    gradeMessage = 'Keep practicing! Review the material and try again. 💪';
  }
  
  return `
    <div class="results-container">
      <div class="results-header">
        <h2>📊 Exam Results</h2>
        <div class="score-card ${gradeClass}">
          <div class="score-circle">
            <span class="score-number">${state.exam.score}</span>
            <span class="score-total">/${state.exam.questions.length}</span>
          </div>
          <div class="score-percentage">${Math.round(percentage)}%</div>
          <div class="score-message">${gradeMessage}</div>
        </div>
      </div>
      
      <div class="review-section">
        <h3>Detailed Review</h3>
        <div class="review-list">
          ${state.exam.review.map((r, idx) => `
            <div class="review-item ${r.is_correct ? 'correct' : 'incorrect'}">
              <div class="review-header">
                <span class="question-number">Q${idx + 1}</span>
                <span class="result-badge ${r.is_correct ? 'correct' : 'incorrect'}">
                  ${r.is_correct ? '✓ Correct' : '✗ Incorrect'}
                </span>
              </div>
              <p class="review-question">${escapeHtml(r.question_text)}</p>
              <div class="review-answers">
                <div class="your-answer">Your answer: <strong>${escapeHtml(r.your_answer_text)}</strong></div>
                ${!r.is_correct ? `<div class="correct-answer">Correct answer: <strong>${escapeHtml(r.correct_answer_text)}</strong></div>` : ''}
              </div>
              <div class="review-explanation">${escapeHtml(r.explanation)}</div>
            </div>
          `).join("")}
        </div>
      </div>
      
      <div class="results-actions">
        <button onclick="setView('exam-config', {course: state.course})" class="btn-primary">🔄 Retake Exam</button>
        <button onclick="setView('home')" class="btn-secondary">🏠 Back to Home</button>
      </div>
    </div>
  `;
}

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

function attachNavEvents() {
  document.querySelectorAll("[data-nav]").forEach(btn => {
    btn.onclick = () => setView(btn.dataset.nav);
  });
}

function attachHomeEvents() {
  document.querySelectorAll("[data-mode]").forEach(btn => {
    btn.onclick = () => {
      state.mode = btn.dataset.mode;
      setView("courses");
    };
  });
}

function attachCourseEvents() {
  document.querySelectorAll(".course-card").forEach(card => {
    card.onclick = () => {
      const course = { 
        id: card.dataset.courseId, 
        name: card.dataset.courseName 
      };
      state.course = course;
      setView(state.mode === "exam" ? "exam-config" : "workspace", { course });
    };
  });

  // Config button handlers
  document.querySelectorAll("[data-count]")?.forEach(btn => {
    btn.onclick = () => {
      currentExamConfig.count = parseInt(btn.dataset.count);
      render();
    };
  });
  
  document.querySelectorAll('input[name="difficulty"]')?.forEach(radio => {
    radio.onchange = () => {
      if (radio.checked) currentExamConfig.difficulty = radio.value;
    };
  });

  document.getElementById("startExamBtn")?.addEventListener("click", () => {
    setView("exam-run");
  });
}

function attachExamEvents() {
  document.querySelectorAll('input[name="question"]').forEach(input => {
    input.onchange = () => {
      state.exam.answers[currentQuestionIndex] = input.value;
      // Auto-save progress
      localStorage.setItem('exam_progress', JSON.stringify({
        answers: state.exam.answers,
        currentIndex: currentQuestionIndex
      }));
    };
  });

  document.getElementById("nextQ")?.addEventListener("click", () => {
    if (currentQuestionIndex < state.exam.questions.length - 1) {
      currentQuestionIndex++;
      render();
    }
  });
  
  document.getElementById("prevQ")?.addEventListener("click", () => {
    if (currentQuestionIndex > 0) {
      currentQuestionIndex--;
      render();
    }
  });
  
  document.getElementById("submitExamBtn")?.addEventListener("click", triggerExamSubmission);
}

async function triggerExamSubmission() {
  if (timerInterval) { 
    clearInterval(timerInterval); 
    timerInterval = null; 
  }
  
  // Check if all questions answered
  const unanswered = state.exam.questions.length - Object.keys(state.exam.answers).length;
  if (unanswered > 0) {
    if (!confirm(`You have ${unanswered} unanswered question(s). Submit anyway?`)) {
      return;
    }
  }
  
  showLoading("Submitting your exam...");
  try {
    const result = await window.api.submitExam(state.exam.questions, state.exam.answers);
    state.exam.score = result.score;
    state.exam.review = result.review;
    localStorage.removeItem('exam_progress');
    setView("results");
  } catch (error) {
    console.error("Error submitting exam:", error);
    showError("Failed to submit exam. Please try again.");
  } finally {
    hideLoading();
  }
}

async function attachTopicEvents() {
  document.querySelectorAll(".topic-item").forEach(item => {
    item.onclick = async () => {
      activeSelectedTopicName = item.dataset.topic;
      const encodedContent = item.dataset.content;
      activeSelectedTopicContent = encodedContent ? decodeURIComponent(encodedContent) : "No content available";
      state.analysis.currentTopic = activeSelectedTopicName;
      state.analysis.loading = true;
      render();

      try {
        const analysis = await window.api.sendChatMessage(
          state.course.id, 
          `Please analyze this topic for exam preparation. Provide key points and important concepts to remember:\n\nTopic: ${activeSelectedTopicName}\n\nContent: ${activeSelectedTopicContent}`, 
          "chat"
        );
        state.analysis.content = analysis || "No analysis available for this topic.";
      } catch (error) {
        console.error("Error getting analysis:", error);
        state.analysis.content = "⚠️ AI analysis temporarily unavailable. Please try again later.";
      } finally {
        state.analysis.loading = false;
        render();
      }
    };
  });
}

function attachGlobalChat() {
  if (document.querySelector(".global-chat")) return;

  const chatContainer = document.createElement("div");
  chatContainer.className = "global-chat collapsed";
  chatContainer.innerHTML = `
    <div class="chat-header" id="chatHeader">
      <div class="chat-header-content">
        <span>🤖 AI Learning Assistant</span>
        <span class="chat-status">● Online</span>
      </div>
      <button class="chat-toggle-btn" id="chatToggle">▲</button>
    </div>
    <div class="chat-body">
      <div class="chat-messages" id="chatMessages">
        <div class="message system">
          👋 Welcome! I'm your AI tutor. Ask me anything about:
          <ul>
            <li>Course concepts and topics</li>
            <li>Exam preparation strategies</li>
            <li>Practice questions and explanations</li>
          </ul>
        </div>
      </div>
      <div class="chat-input-area">
        <input type="text" id="chatInput" placeholder="Ask a question..." autocomplete="off">
        <button id="sendChatBtn">Send</button>
      </div>
    </div>
  `;

  document.body.appendChild(chatContainer);

  const chatHeader = document.getElementById("chatHeader");
  if (chatHeader) {
    chatHeader.onclick = () => {
      chatContainer.classList.toggle("collapsed");
      const toggleBtn = document.getElementById("chatToggle");
      if (toggleBtn) {
        toggleBtn.innerText = chatContainer.classList.contains("collapsed") ? "▲" : "▼";
      }
    };
  }

  let isSending = false;
  const sendBtn = document.getElementById("sendChatBtn");
  const chatInput = document.getElementById("chatInput");
  const messagesContainer = document.getElementById("chatMessages");

  const handleSend = async () => {
    if (isSending) return;
    const text = chatInput.value.trim();
    if (!text) return;

    isSending = true;
    sendBtn.disabled = true;
    
    const userMsg = document.createElement("div");
    userMsg.className = "message user";
    userMsg.innerText = text;
    messagesContainer.appendChild(userMsg);
    chatInput.value = "";
    messagesContainer.scrollTop = messagesContainer.scrollHeight;

    const thinkingMsg = document.createElement("div");
    thinkingMsg.className = "message ai thinking";
    thinkingMsg.innerText = "🤔 Thinking...";
    messagesContainer.appendChild(thinkingMsg);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;

    try {
      const activeCourseId = state.course ? state.course.id : "CS101";
      const aiResponse = await window.api.sendChatMessage(activeCourseId, text, "chat");
      
      thinkingMsg.remove();
      const aiMsg = document.createElement("div");
      aiMsg.className = "message ai";
      aiMsg.innerText = aiResponse || "I couldn't process that request. Please try again.";
      messagesContainer.appendChild(aiMsg);
    } catch (error) {
      console.error("Chat error:", error);
      thinkingMsg.remove();
      const errorMsg = document.createElement("div");
      errorMsg.className = "message error";
      errorMsg.innerText = "Sorry, I'm having trouble connecting. Please try again.";
      messagesContainer.appendChild(errorMsg);
    } finally {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
      isSending = false;
      sendBtn.disabled = false;
      chatInput.focus();
    }
  };

  if (sendBtn) sendBtn.onclick = handleSend;
  if (chatInput) chatInput.onkeypress = (e) => { if (e.key === "Enter") handleSend(); };
}

// Make functions available globally
window.setView = setView;
window.state = state;