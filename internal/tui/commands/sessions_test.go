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

func TestGetConversations(t *testing.T) {
	layout := "Monday, Jan 2, 2006 at 03:04 PM"

	mockTimeString := "Monday, Jan 2, 2006 at 03:04 PM"

	mockTime, err := time.Parse(layout, mockTimeString)
	if err != nil {
		t.Fatalf("failed to parse mock time: %v", err)
	}
	updatedTime, err := time.Parse(layout, "Tuesday, Jan 3, 2006 at 04:05 PM")
	if err != nil {
		t.Fatalf("failed to parse updated time: %v", err)
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
			name: "return multiple conversations",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return []repo.Conversation{
					{
						ID:        1,
						Title:     "Conversation 1",
						Directory: "/path/to/dir1",
						CreatedAt: mockTime,
						UpdatedAt: mockTime,
					},
					{
						ID:        2,
						Title:     "Conversation 2",
						Directory: "/path/to/dir2",
						CreatedAt: updatedTime,
						UpdatedAt: updatedTime,
					},
				}, nil
			},
			expected: []list.Item{
				ConversationItem{
					CreatedAt: mockTimeString,
					Conversation: repo.Conversation{
						ID:        1,
						Title:     "Conversation 1",
						Directory: "/path/to/dir1",
						CreatedAt: mockTime,
						UpdatedAt: mockTime,
					},
				},
				ConversationItem{
					CreatedAt: "Tuesday, Jan 3, 2006 at 04:05 PM",
					Conversation: repo.Conversation{
						ID:        2,
						Title:     "Conversation 2",
						Directory: "/path/to/dir2",
						CreatedAt: updatedTime,
						UpdatedAt: updatedTime,
					},
				},
			},
		},
		{
			name: "return empty slice when db returns no conversations",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return []repo.Conversation{}, nil
			},
			expected: []list.Item{},
		},
		{
			name: "return empty slice when db returns nil",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return nil, nil
			},
			expected: []list.Item{},
		},
		{
			name: "return empty slice when conversation returns error",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return nil, fmt.Errorf("database error")
			},
			expected: []list.Item{},
		},
		{
			name: "return conversations with different timestamps",
			getConversationsFunc: func(ctx context.Context) ([]repo.Conversation, error) {
				return []repo.Conversation{
					{
						ID:        42,
						Title:     "Test Conversation",
						Directory: "/tmp/test",
						CreatedAt: updatedTime,
						UpdatedAt: mockTime,
					},
				}, nil
			},
			expected: []list.Item{
				ConversationItem{
					CreatedAt: "Tuesday, Jan 3, 2006 at 04:05 PM",
					Conversation: repo.Conversation{
						ID:        42,
						Title:     "Test Conversation",
						Directory: "/tmp/test",
						CreatedAt: updatedTime,
						UpdatedAt: mockTime,
					},
				},
			},
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
