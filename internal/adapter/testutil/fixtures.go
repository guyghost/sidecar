package testutil

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// ClaudeCodeMessage represents a JSONL message in Claude Code format.
type ClaudeCodeMessage struct {
	Type      string                `json:"type"`
	UUID      string                `json:"uuid"`
	SessionID string                `json:"sessionId"`
	Timestamp time.Time             `json:"timestamp"`
	Message   *ClaudeCodeMsgContent `json:"message,omitempty"`
	CWD       string                `json:"cwd,omitempty"`
	Version   string                `json:"version,omitempty"`
}

// ClaudeCodeMsgContent holds the actual message content.
type ClaudeCodeMsgContent struct {
	Role    string           `json:"role"`
	Content json.RawMessage  `json:"content"`
	Model   string           `json:"model,omitempty"`
	Usage   *ClaudeCodeUsage `json:"usage,omitempty"`
}

// ClaudeCodeUsage tracks token usage.
type ClaudeCodeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// GenerateClaudeCodeSessionFile creates a JSONL file with realistic message content.
// messageCount determines the number of message pairs (user + assistant).
// avgMessageSize is the approximate size of message content in bytes.
func GenerateClaudeCodeSessionFile(path string, messageCount int, avgMessageSize int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	sessionID := "bench-session-001"
	enc := json.NewEncoder(f)
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Write summary line
	summary := ClaudeCodeMessage{
		Type:      "summary",
		UUID:      "summary-001",
		SessionID: sessionID,
		Timestamp: baseTime,
		CWD:       "/home/user/project",
		Version:   "0.2.61",
	}
	if err := enc.Encode(summary); err != nil {
		return err
	}

	// Generate message pairs
	for i := 0; i < messageCount; i++ {
		ts := baseTime.Add(time.Duration(i*2) * time.Second)

		// User message
		userContent := generateTextContent(avgMessageSize/2, "user", i)
		user := ClaudeCodeMessage{
			Type:      "user",
			UUID:      fmt.Sprintf("msg-user-%06d", i),
			SessionID: sessionID,
			Timestamp: ts,
			Message: &ClaudeCodeMsgContent{
				Role:    "user",
				Content: userContent,
			},
		}
		if err := enc.Encode(user); err != nil {
			return err
		}

		// Assistant message with tool use and thinking
		assistantContent := generateAssistantContent(avgMessageSize/2, i)
		assistant := ClaudeCodeMessage{
			Type:      "assistant",
			UUID:      fmt.Sprintf("msg-asst-%06d", i),
			SessionID: sessionID,
			Timestamp: ts.Add(time.Second),
			Message: &ClaudeCodeMsgContent{
				Role:    "assistant",
				Content: assistantContent,
				Model:   "claude-sonnet-4-20250514",
				Usage: &ClaudeCodeUsage{
					InputTokens:  500 + (i % 100),
					OutputTokens: 200 + (i % 50),
				},
			},
		}
		if err := enc.Encode(assistant); err != nil {
			return err
		}
	}

	return nil
}

// generateTextContent creates a content block with text of approximately the given size.
func generateTextContent(size int, role string, index int) json.RawMessage {
	text := fmt.Sprintf("%s message #%d: ", role, index)
	padding := make([]byte, size-len(text))
	for i := range padding {
		padding[i] = 'x'
	}
	text += string(padding)

	blocks := []map[string]any{
		{"type": "text", "text": text},
	}
	data, _ := json.Marshal(blocks)
	return data
}

// generateAssistantContent creates assistant content with tool use and thinking blocks.
func generateAssistantContent(size int, index int) json.RawMessage {
	textSize := size / 2
	thinkingSize := size / 4
	toolInputSize := size / 4

	// Create padding
	textPadding := make([]byte, textSize)
	for i := range textPadding {
		textPadding[i] = 'a'
	}

	thinkingPadding := make([]byte, thinkingSize)
	for i := range thinkingPadding {
		thinkingPadding[i] = 't'
	}

	toolInput := make([]byte, toolInputSize)
	for i := range toolInput {
		toolInput[i] = 'i'
	}

	blocks := []map[string]any{
		{"type": "thinking", "thinking": fmt.Sprintf("Thinking about message %d: %s", index, string(thinkingPadding))},
		{"type": "text", "text": fmt.Sprintf("Response %d: %s", index, string(textPadding))},
	}

	// Add tool use every 5 messages
	if index%5 == 0 {
		blocks = append(blocks, map[string]any{
			"type":  "tool_use",
			"id":    fmt.Sprintf("toolu_%06d", index),
			"name":  "Read",
			"input": map[string]any{"file_path": fmt.Sprintf("/path/to/file_%d.go", index), "content": string(toolInput)},
		})
	}

	data, _ := json.Marshal(blocks)
	return data
}

// CodexMessage represents a JSONL message in Codex format.
type CodexMessage struct {
	Type      string    `json:"type"`
	SessionID string    `json:"session_id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Role      string    `json:"role,omitempty"`
	Content   string    `json:"content,omitempty"`
	Model     string    `json:"model,omitempty"`
	CallID    string    `json:"call_id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Arguments string    `json:"arguments,omitempty"`
	Output    string    `json:"output,omitempty"`
	Input     int       `json:"input,omitempty"`
	Output_   int       `json:"output_,omitempty"` // for usage
}

// GenerateCodexSessionFile creates a JSONL file in Codex format.
func GenerateCodexSessionFile(path string, messageCount int, avgMessageSize int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	sessionID := "codex-bench-001"
	enc := json.NewEncoder(f)
	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Write session start
	start := CodexMessage{
		Type:      "session_start",
		SessionID: sessionID,
		Timestamp: baseTime,
	}
	if err := enc.Encode(start); err != nil {
		return err
	}

	// Generate messages
	for i := 0; i < messageCount; i++ {
		ts := baseTime.Add(time.Duration(i*2) * time.Second)

		// User message
		userContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("User message %d: ", i))
		user := CodexMessage{
			Type:      "message",
			Role:      "user",
			Content:   userContent,
			Timestamp: ts,
		}
		if err := enc.Encode(user); err != nil {
			return err
		}

		// Assistant message
		assistantContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("Assistant response %d: ", i))
		assistant := CodexMessage{
			Type:      "message",
			Role:      "assistant",
			Content:   assistantContent,
			Model:     "gpt-4o",
			Timestamp: ts.Add(time.Second),
		}
		if err := enc.Encode(assistant); err != nil {
			return err
		}

		// Add tool call every 5 messages
		if i%5 == 0 {
			toolCall := CodexMessage{
				Type:      "function_call",
				CallID:    fmt.Sprintf("call_%06d", i),
				Name:      "shell",
				Arguments: fmt.Sprintf(`{"cmd": "echo test %d"}`, i),
				Timestamp: ts.Add(500 * time.Millisecond),
			}
			if err := enc.Encode(toolCall); err != nil {
				return err
			}

			toolOutput := CodexMessage{
				Type:      "function_output",
				CallID:    fmt.Sprintf("call_%06d", i),
				Output:    fmt.Sprintf("test %d\n", i),
				Timestamp: ts.Add(600 * time.Millisecond),
			}
			if err := enc.Encode(toolOutput); err != nil {
				return err
			}
		}
	}

	return nil
}

// generatePaddedString creates a string of approximately the given size with a prefix.
func generatePaddedString(size int, prefix string) string {
	if size <= len(prefix) {
		return prefix
	}
	padding := make([]byte, size-len(prefix))
	for i := range padding {
		padding[i] = 'x'
	}
	return prefix + string(padding)
}

// ApproximateMessageCount returns the message count to generate a file of approximately the given size.
// size is in bytes, avgMessageSize is the target size per message pair.
func ApproximateMessageCount(targetSize int64, avgMessageSize int) int {
	// Each message pair is approximately 2 * avgMessageSize + overhead (~100 bytes)
	pairSize := 2*avgMessageSize + 100
	return int(targetSize) / pairSize
}

// GeminiCLISession represents a session file in Gemini CLI format.
type GeminiCLISession struct {
	SessionID   string             `json:"sessionId"`
	ProjectHash string             `json:"projectHash"`
	StartTime   string             `json:"startTime"`
	LastUpdated string             `json:"lastUpdated"`
	Messages    []GeminiCLIMessage `json:"messages"`
}

// GeminiCLIMessage represents a message in Gemini CLI format.
type GeminiCLIMessage struct {
	Type     string             `json:"type"`
	Content  string             `json:"content"`
	Model    string             `json:"model,omitempty"`
	Tokens   *GeminiCLITokens   `json:"tokens,omitempty"`
	Thoughts []GeminiCLIThought `json:"thoughts,omitempty"`
}

// GeminiCLIThought represents a thought in Gemini CLI format.
type GeminiCLIThought struct {
	Subject     string `json:"subject"`
	Description string `json:"description,omitempty"`
}

// GeminiCLITokens tracks token usage.
type GeminiCLITokens struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// GenerateGeminiCLISessionFile creates a JSON session file in Gemini CLI format.
func GenerateGeminiCLISessionFile(path string, messageCount int, avgMessageSize int) error {
	session := GeminiCLISession{
		SessionID:   "session-001",
		ProjectHash: "abc123",
		StartTime:   "2024-01-15T10:00:00Z",
		LastUpdated: "2024-01-15T12:00:00Z",
		Messages:    make([]GeminiCLIMessage, 0, messageCount*2),
	}

	for i := 0; i < messageCount; i++ {
		userContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("User message %d: ", i))
		session.Messages = append(session.Messages, GeminiCLIMessage{
			Type:    "user",
			Content: userContent,
			Tokens:  &GeminiCLITokens{Input: 0, Output: 0},
		})

		assistantContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("Assistant response %d: ", i))
		thoughts := []GeminiCLIThought{}
		if i%5 == 0 {
			thoughts = append(thoughts, GeminiCLIThought{
				Subject: "Thinking",
			})
		}
		session.Messages = append(session.Messages, GeminiCLIMessage{
			Type:     "gemini",
			Content:  assistantContent,
			Model:    "gemini-2.5-pro",
			Tokens:   &GeminiCLITokens{Input: 100 + i%10, Output: 50 + i%5},
			Thoughts: thoughts,
		})
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(session)
}

// AmpMessage represents a message in Amp format.
type AmpMessage struct {
	Role      string       `json:"role"`
	MessageID int          `json:"messageId"`
	Content   []AmpContent `json:"content"`
	Usage     *AmpUsage    `json:"usage,omitempty"`
	Meta      AmpMeta      `json:"meta"`
}

// AmpContent represents content blocks in Amp format.
type AmpContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// AmpUsage tracks token usage in Amp format.
type AmpUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AmpMeta contains metadata for Amp messages.
type AmpMeta struct {
	SentAt int64 `json:"sentAt"`
}

// AmpEnv represents environment info in Amp format.
type AmpEnv struct {
	Initial AmpInitial `json:"initial"`
}

// AmpInitial represents initial environment state.
type AmpInitial struct {
	Trees []AmpTree `json:"trees"`
}

// AmpTree represents a project tree URI.
type AmpTree struct {
	URI string `json:"uri"`
}

// AmpThread represents a thread file in Amp format.
type AmpThread struct {
	V        int          `json:"v"`
	ID       string       `json:"id"`
	Created  int64        `json:"created"`
	Messages []AmpMessage `json:"messages"`
	Env      AmpEnv       `json:"env"`
}

// GenerateAmpThreadFile creates a JSON thread file in Amp format.
// projectURI should be like "file:///path/to/project".
func GenerateAmpThreadFile(path string, messageCount int, avgMessageSize int, projectURI string) error {
	thread := AmpThread{
		V:        1,
		ID:       "T-abc123",
		Created:  1705312800000,
		Messages: make([]AmpMessage, 0, messageCount*2),
		Env: AmpEnv{
			Initial: AmpInitial{
				Trees: []AmpTree{{URI: projectURI}},
			},
		},
	}

	baseTime := int64(1705312800000)

	for i := 0; i < messageCount; i++ {
		ts := baseTime + int64(i*2000)

		userContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("User message %d: ", i))
		thread.Messages = append(thread.Messages, AmpMessage{
			Role:      "user",
			MessageID: i * 2,
			Content:   []AmpContent{{Type: "text", Text: userContent}},
			Meta:      AmpMeta{SentAt: ts},
		})

		assistantContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("Assistant response %d: ", i))
		content := []AmpContent{{Type: "text", Text: assistantContent}}
		if i%5 == 0 {
			content = append(content, AmpContent{Type: "thinking", Text: "Let me think about this..."})
		}

		ts2 := ts + 1000
		thread.Messages = append(thread.Messages, AmpMessage{
			Role:      "assistant",
			MessageID: i*2 + 1,
			Content:   content,
			Usage:     &AmpUsage{InputTokens: 100 + i%10, OutputTokens: 50 + i%5},
			Meta:      AmpMeta{SentAt: ts2},
		})
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(thread)
}

// KiroHistoryEntry represents a conversation turn in Kiro format.
type KiroHistoryEntry struct {
	User      KiroUserMessage      `json:"user"`
	Assistant KiroAssistantMessage `json:"assistant"`
}

// KiroUserMessage represents a user message in Kiro format.
type KiroUserMessage struct {
	Prompt KiroPrompt `json:"Prompt"`
}

// KiroPrompt represents a prompt with content and context.
type KiroPrompt struct {
	Content      string   `json:"content"`
	ContextFiles []string `json:"context_files"`
}

// KiroAssistantMessage represents an assistant message in Kiro format.
type KiroAssistantMessage struct {
	Response KiroResponse `json:"Response"`
}

// KiroResponse represents an assistant response.
type KiroResponse struct {
	Content []KiroContentBlock `json:"content"`
}

// KiroContentBlock represents a content block in Kiro format.
type KiroContentBlock struct {
	Text string `json:"text"`
}

// KiroConversation represents the full conversation JSON for Kiro.
type KiroConversation struct {
	Model        string             `json:"model"`
	SystemPrompt string             `json:"system_prompt"`
	History      []KiroHistoryEntry `json:"history"`
}

// GenerateKiroConversationJSON creates a JSON blob for Kiro's conversations_v2.value column.
func GenerateKiroConversationJSON(messageCount int, avgMessageSize int) ([]byte, error) {
	conv := KiroConversation{
		Model:        "claude-sonnet-4-20250514",
		SystemPrompt: "You are a helpful assistant.",
		History:      make([]KiroHistoryEntry, 0, messageCount),
	}

	for i := 0; i < messageCount; i++ {
		userContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("User message %d: ", i))
		assistantContent := generatePaddedString(avgMessageSize/2, fmt.Sprintf("Assistant response %d: ", i))

		conv.History = append(conv.History, KiroHistoryEntry{
			User: KiroUserMessage{
				Prompt: KiroPrompt{
					Content:      userContent,
					ContextFiles: []string{},
				},
			},
			Assistant: KiroAssistantMessage{
				Response: KiroResponse{
					Content: []KiroContentBlock{{Text: assistantContent}},
				},
			},
		})
	}

	return json.Marshal(conv)
}

// OpenCodeProject represents a project file in OpenCode format.
type OpenCodeProject struct {
	ID       string `json:"id"`
	Worktree string `json:"worktree"`
}

// OpenCodeSession represents a session file in OpenCode format.
type OpenCodeSession struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	ParentID string       `json:"parentID"`
	Time     OpenCodeTime `json:"time"`
}

// OpenCodeTime holds timestamp information for OpenCode sessions/messages.
type OpenCodeTime struct {
	Created int64 `json:"created"`
	Updated int64 `json:"updated,omitempty"`
}

// OpenCodeMessage represents a message file in OpenCode format.
type OpenCodeMessage struct {
	ID        string          `json:"id"`
	SessionID string          `json:"sessionID"`
	Role      string          `json:"role"`
	Time      OpenCodeTime    `json:"time"`
	Tokens    *OpenCodeTokens `json:"tokens,omitempty"`
}

// OpenCodeTokens tracks token usage in OpenCode format.
type OpenCodeTokens struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// OpenCodePart represents a part file in OpenCode format.
type OpenCodePart struct {
	ID        string `json:"id"`
	MessageID string `json:"messageID"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

// GenerateOpenCodeStorage creates the multi-file directory structure for OpenCode.
// Returns the storage root directory path.
func GenerateOpenCodeStorage(baseDir string, projectPath string, sessionCount int, messagesPerSession int, avgMessageSize int) error {
	// Create directory structure
	dirs := []string{
		baseDir + "/project",
		baseDir + "/session",
		baseDir + "/message",
		baseDir + "/part",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create project file
	projectID := "proj-001"
	project := OpenCodeProject{
		ID:       projectID,
		Worktree: projectPath,
	}
	projectData, err := json.Marshal(project)
	if err != nil {
		return err
	}
	if err := os.WriteFile(baseDir+"/project/"+projectID+".json", projectData, 0644); err != nil {
		return err
	}

	baseTime := int64(1705312800)

	// Create session files and nested message/part files
	for i := 0; i < sessionCount; i++ {
		sessionID := fmt.Sprintf("sess-%03d", i)
		sessionTime := OpenCodeTime{
			Created: baseTime + int64(i*1000),
			Updated: baseTime + int64(i*1000+100),
		}
		session := OpenCodeSession{
			ID:       sessionID,
			Title:    fmt.Sprintf("Session %d", i+1),
			ParentID: projectID,
			Time:     sessionTime,
		}
		sessionData, err := json.Marshal(session)
		if err != nil {
			return err
		}

		// Create session directory and file
		sessionDir := baseDir + "/session/" + projectID
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(sessionDir+"/"+sessionID+".json", sessionData, 0644); err != nil {
			return err
		}

		// Create messages for this session
		for j := 0; j < messagesPerSession; j++ {
			msgID := fmt.Sprintf("msg-%06d", i*1000+j)
			role := "user"
			if j%2 == 1 {
				role = "assistant"
			}

			msgTime := OpenCodeTime{
				Created: baseTime + int64(i*1000+j*10),
				Updated: baseTime + int64(i*1000+j*10+5),
			}
			msg := OpenCodeMessage{
				ID:        msgID,
				SessionID: sessionID,
				Role:      role,
				Time:      msgTime,
				Tokens:    &OpenCodeTokens{Input: 100 + j%10, Output: 50 + j%5},
			}

			msgData, err := json.Marshal(msg)
			if err != nil {
				return err
			}

			// Create message directory and file
			msgDir := baseDir + "/message/" + sessionID
			if err := os.MkdirAll(msgDir, 0755); err != nil {
				return err
			}
			if err := os.WriteFile(msgDir+"/"+msgID+".json", msgData, 0644); err != nil {
				return err
			}

			// Create part for this message
			partID := fmt.Sprintf("part-%06d", i*1000+j)
			content := generatePaddedString(avgMessageSize/2, fmt.Sprintf("%s message %d-%d: ", role, i, j))
			part := OpenCodePart{
				ID:        partID,
				MessageID: msgID,
				Type:      "text",
				Content:   content,
			}

			partData, err := json.Marshal(part)
			if err != nil {
				return err
			}

			// Create part directory and file
			partDir := baseDir + "/part/" + msgID
			if err := os.MkdirAll(partDir, 0755); err != nil {
				return err
			}
			if err := os.WriteFile(partDir+"/"+partID+".json", partData, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateCursorDB creates a SQLite database with Cursor's schema.
// Returns the path to the created store.db file.
func GenerateCursorDB(dir string, messageCount int, avgMessageSize int) (string, error) {
	dbPath := dir + "/store.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Create tables
	if _, err := db.Exec(`
		CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT);
		CREATE TABLE blobs (id TEXT PRIMARY KEY, data TEXT);
	`); err != nil {
		return "", err
	}

	// Insert meta entry with simplified schema
	metaJSON := map[string]any{
		"version":    "1.0.0",
		"created_at": time.Now().Unix(),
	}
	metaBytes, _ := json.Marshal(metaJSON)
	if _, err := db.Exec("INSERT INTO meta (key, value) VALUES (?, ?)", "0", fmt.Sprintf("%x", metaBytes)); err != nil {
		return "", err
	}

	// Insert a session blob
	sessionID := "bench-session-001"
	session := map[string]any{
		"id":         sessionID,
		"created_at": time.Now().Unix(),
		"updated_at": time.Now().Unix(),
		"messages":   messageCount * 2,
	}
	sessionBytes, _ := json.Marshal(session)
	if _, err := db.Exec("INSERT INTO blobs (id, data) VALUES (?, ?)", sessionID, string(sessionBytes)); err != nil {
		return "", err
	}

	return dbPath, nil
}

// GenerateWarpDB creates a SQLite database with Warp's schema.
// Populates ai_queries and blocks tables.
func GenerateWarpDB(path string, conversationCount int, messagesPerConversation int, projectDir string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create tables
	if _, err := db.Exec(`
		CREATE TABLE ai_queries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			exchange_id TEXT NOT NULL,
			input TEXT,
			model_id TEXT,
			start_ts TEXT,
			working_directory TEXT,
			conversation_id TEXT
		);
		CREATE TABLE blocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			stylized_command TEXT,
			stylized_output TEXT,
			exit_code INTEGER,
			start_ts TEXT,
			ai_metadata TEXT
		);
		CREATE TABLE agent_conversations (
			conversation_id TEXT PRIMARY KEY,
			conversation_data TEXT
		);
	`); err != nil {
		return err
	}

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Populate data
	for i := 0; i < conversationCount; i++ {
		uuid := fmt.Sprintf("warp-conv-%06d", i)
		exchangeID := fmt.Sprintf("exchange-%06d", i)
		conversationID := fmt.Sprintf("conv-%06d", i)
		inputJSON := fmt.Sprintf(`[{"Query": {"text": "Conversation %d prompt with some content"}}]`, i)

		// Insert ai_query
		_, err := db.Exec(`
			INSERT INTO ai_queries (uuid, exchange_id, input, model_id, start_ts, working_directory, conversation_id)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, uuid, exchangeID, inputJSON, "gpt-4", baseTime.Format("2006-01-02 15:04:05"), projectDir, conversationID)
		if err != nil {
			return err
		}

		// Add blocks for messages
		for j := 0; j < messagesPerConversation; j++ {
			blockTime := baseTime.Add(time.Duration(j) * time.Second)
			aiMeta := fmt.Sprintf(`{"conversation_id": "%s", "action_id": "action-%06d-%06d", "type": "command"}`, conversationID, i, j)

			if _, err := db.Exec(`
				INSERT INTO blocks (stylized_command, stylized_output, exit_code, start_ts, ai_metadata)
				VALUES (?, ?, ?, ?, ?)
			`, fmt.Sprintf("echo %d", j), fmt.Sprintf("output %d", j), 0, blockTime.Format("2006-01-02 15:04:05"), aiMeta); err != nil {
				return err
			}
		}

		// Insert agent_conversations entry
		convDataJSON := `{"usage_metadata": {"token_usage": [{"warp_tokens": 1000, "byok_tokens": 0, "credits_spent": 10}], "credits_spent": 10}`
		if _, err := db.Exec("INSERT INTO agent_conversations (conversation_id, conversation_data) VALUES (?, ?)", conversationID, convDataJSON); err != nil {
			return err
		}

		baseTime = baseTime.Add(time.Hour)
	}

	return nil
}

// GenerateKiroDB creates a SQLite database with Kiro's schema.
// Populates conversations_v2 table.
func GenerateKiroDB(path string, projectPath string, conversationCount int, messagesPerConversation int, avgMessageSize int) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create table
	if _, err := db.Exec(`
		CREATE TABLE conversations_v2 (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL,
			conversation_id TEXT NOT NULL,
			project_path TEXT NOT NULL,
			value TEXT NOT NULL,
			created_at INTEGER,
			updated_at INTEGER
		);
	`); err != nil {
		return err
	}

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC).Unix()

	// Populate conversations
	for i := 0; i < conversationCount; i++ {
		convID := fmt.Sprintf("conv-%06d", i)
		convJSON, err := GenerateKiroConversationJSON(messagesPerConversation, avgMessageSize)
		if err != nil {
			return err
		}

		if _, err := db.Exec(`
			INSERT INTO conversations_v2 (key, conversation_id, project_path, value, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, projectPath, convID, projectPath, string(convJSON), baseTime+int64(i*3600), baseTime+int64(i*3600+600)); err != nil {
			return err
		}
	}

	return nil
}
