package web

import (
	"encoding/json"
	"strings"
	"time"
)

type fileResult struct {
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

type settingsProvider struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Configured bool            `json:"configured"`
	Models     []settingsModel `json:"models"`
}

type settingsModel struct {
	ID        string `json:"id"`
	Context   int    `json:"context"`
	MaxOutput int    `json:"maxOutput"`
}

type stateResponse struct {
	Directory string        `json:"directory"`
	Version   string        `json:"version"`
	Provider  string        `json:"provider"`
	Model     string        `json:"model"`
	Running   *runningState `json:"running"`
}

type runningState struct {
	ConversationID int64 `json:"conversationId"`
}

type modelsResponse struct {
	Providers []providerModels `json:"providers"`
	Provider  string           `json:"provider"`
	Model     string           `json:"model"`
}

type providerModels struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Models []string `json:"models"`
}

type setModelRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type conversationItem struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

type renameConversationRequest struct {
	Title string `json:"title"`
}

type renameConversationResponse struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type settingsResponse struct {
	Models []settingsProvider `json:"models"`
	Active settingsActive     `json:"active"`
	Git    settingsGit        `json:"git"`
	About  settingsAbout      `json:"about"`
}

type settingsActive struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type settingsGit struct {
	Initialized bool   `json:"initialized"`
	Dirty       bool   `json:"dirty"`
	Dir         string `json:"dir"`
	Hint        string `json:"hint"`
}

type settingsAbout struct {
	Version      string `json:"version"`
	Directory    string `json:"directory"`
	ConfigPath   string `json:"configPath"`
	LogPath      string `json:"logPath"`
	LatestUpdate string `json:"latestUpdate"`
}

type skillItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// pluginItem shares the shape of skillItem, like the web app's types.
type pluginItem = skillItem

type diffResponse struct {
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated"`
}

type chatRequest struct {
	ID      json.RawMessage `json:"id"`
	Message uiMessage       `json:"message"`
}

type uiMessage struct {
	ID    string   `json:"id"`
	Role  string   `json:"role"`
	Text  string   `json:"text"`
	Parts []uiPart `json:"parts"`
}

type uiPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (m uiMessage) text() string {
	if len(m.Parts) == 0 {
		return m.Text
	}
	var b strings.Builder
	for _, p := range m.Parts {
		if p.Type == "text" || p.Type == "" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

type errorResponse struct {
	Error string `json:"error"`
}
