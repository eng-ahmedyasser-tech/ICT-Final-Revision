const API_BASE = "";  // سيتم التعامل مع المسارات النسبية تلقائياً

async function request(endpoint, options = {}) {
  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      headers: { "Content-Type": "application/json" },
      ...options,
    });
    if (!response.ok) {
      const errorText = await response.text();
      console.error(`HTTP ${response.status}: ${errorText}`);
      throw new Error(`HTTP ${response.status}: ${errorText}`);
    }
    return await response.json();
  } catch (error) {
    console.error(`API Error on ${endpoint}:`, error);
    return null;
  }
}

function normalizeCourseId(courseId) {
  if (!courseId) return "CS101";
  let clean = courseId.toString().trim().toUpperCase();
  if (clean === "1" || clean === "101" || clean === "CS1") return "CS101";
  if (clean === "2" || clean === "201" || clean === "CS2") return "CS201";
  if (clean === "3" || clean === "301" || clean === "CS3") return "CS301";
  return clean;
}

async function getCourses() {
  try {
    const result = await request("/api/courses", { method: "GET" });
    if (!result) return [];
    if (Array.isArray(result)) return result;
    if (result.courses && Array.isArray(result.courses)) return result.courses;
    return [];
  } catch (error) {
    console.error("Error fetching courses:", error);
    return [];
  }
}

async function getCurriculum(courseId) {
  const normalizedId = normalizeCourseId(courseId);
  console.log(`Fetching curriculum for: ${normalizedId}`);
  try {
    const result = await request(`/api/curriculum?courseId=${normalizedId}`, { method: "GET" });
    console.log("Curriculum API response:", result);
    if (!result) return [];
    if (Array.isArray(result)) {
      return result.map(item => ({
        topic: item.topic || item.Topic || "Untitled Topic",
        content: item.content || item.Content || "No content available"
      }));
    }
    if (result.data && Array.isArray(result.data)) {
      return result.data.map(item => ({
        topic: item.topic || item.Topic || "Untitled Topic",
        content: item.content || item.Content || "No content available"
      }));
    }
    return [];
  } catch (error) {
    console.error(`Error fetching curriculum for ${normalizedId}:`, error);
    return [];
  }
}

async function generateExam(courseId, count, difficulty) {
  const payload = {
    courseId: normalizeCourseId(courseId),
    count: parseInt(count) || 10,
    difficulty: difficulty || "easy"
  };
  console.log("Generating exam with payload:", payload);
  try {
    const result = await request("/api/generate-exam", { 
      method: "POST", 
      body: JSON.stringify(payload) 
    });
    if (!result) return [];
    if (Array.isArray(result)) return result;
    if (result.questions && Array.isArray(result.questions)) return result.questions;
    return [];
  } catch (error) {
    console.error("Error generating exam:", error);
    return [];
  }
}

async function submitExam(questions, answers) {
  const payload = {
    questions: questions || [],
    answers: answers || {}
  };
  console.log("Submitting exam with", Object.keys(answers).length, "answers");
  try {
    const result = await request("/api/submit-exam", { 
      method: "POST", 
      body: JSON.stringify(payload) 
    });
    if (!result) return { score: 0, review: [] };
    return {
      score: result.score || 0,
      review: result.review || [],
      percentage: result.percentage || 0,
      passed: result.passed || false
    };
  } catch (error) {
    console.error("Error submitting exam:", error);
    return { score: 0, review: [] };
  }
}

async function sendChatMessage(courseId, message, mode = "chat") {
  if (!message || message.trim() === "") return "Please enter a message.";
  const payload = {
    courseId: normalizeCourseId(courseId),
    message: message.trim(),
    mode: mode
  };
  console.log("Sending chat message:", { courseId: payload.courseId, messageLength: message.length, mode });
  try {
    const data = await request("/api/chat", { 
      method: "POST", 
      body: JSON.stringify(payload) 
    });
    if (data && data.response) return data.response;
    if (data && data.message) return data.message;
    return "I understand your question. Let me help you with that. Could you please provide more details about what you'd like to learn?";
  } catch (error) {
    console.error("Error sending chat message:", error);
    return "Connection error. Please check if the server is running and try again.";
  }
}

window.api = { 
  getCourses, 
  getCurriculum, 
  generateExam, 
  submitExam, 
  sendChatMessage 
};

window.apiDebug = {
  testGetCourses: async () => {
    const courses = await getCourses();
    console.log("Courses:", courses);
    return courses;
  },
  testGetCurriculum: async (courseId) => {
    const curriculum = await getCurriculum(courseId);
    console.log("Curriculum:", curriculum);
    return curriculum;
  }
};