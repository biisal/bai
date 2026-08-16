package commands

import (
	"context"
	"fmt"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	repo "github.com/biisal/bai/internal/db/sqlc"
	test_utils "github.com/biisal/bai/utils/tests"
)

func TestParseConversations(t *testing.T) {
	layout := "Monday, Jan 2, 2006 at 03:04 PM"
	mockTimeString := "Monday, Jan 2, 2006 at 03:04 PM"
	mockTime, err := time.Parse(layout, mockTimeString)
	if err != nil {
		t.Fatalf("failed to parse mock time: %v", err)
	}

	tests := []struct {
		name                 string
		getConversationsFunc func(ctx context.Context) ([]repo.Conversation, error)
		expected             []list.Item
	}{
		{
			name: "return conversations fetching from db",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return []repo.Conversation{
					{
						ID:        1,
						Title:     "Conversation 1",
						Directory: "/path/to/dir",
						CreatedAt: mockTime,
						UpdatedAt: mockTime,
					},
				}, nil
			},
			expected: []list.Item{
				ConversationItem{
					CreatedAt: mockTimeString,
					Conversation: repo.Conversation{
						ID:        1,
						Title:     "Conversation 1",
						Directory: "/path/to/dir",
						CreatedAt: mockTime,
						UpdatedAt: mockTime,
					},
				},
			},
		},
		{
			name: "returns empty slice when conversation returns error ",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return nil, fmt.Errorf("error")
			},
			expected: []list.Item{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got := parseConversations(ctx, tt.getConversationsFunc)
			test_utils.AssertSliceEqual(t, got, tt.expected)
		})
	}
}
