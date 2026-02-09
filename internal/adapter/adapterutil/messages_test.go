package adapterutil

import (
	"testing"

	"github.com/marcus/sidecar/internal/adapter"
)

func TestCopyMessages(t *testing.T) {
	tests := []struct {
		name string
		msgs []adapter.Message
		want []adapter.Message
		deep bool // check that deep copy works
	}{
		{
			name: "nil slice",
			msgs: nil,
			want: nil,
		},
		{
			name: "empty slice",
			msgs: []adapter.Message{},
			want: []adapter.Message{},
		},
		{
			name: "single message without slices",
			msgs: []adapter.Message{
				{ID: "msg1", Role: "user", Content: "test"},
			},
			want: []adapter.Message{
				{ID: "msg1", Role: "user", Content: "test"},
			},
			deep: false,
		},
		{
			name: "message with tool uses",
			msgs: []adapter.Message{
				{
					ID:   "msg1",
					Role: "assistant",
					ToolUses: []adapter.ToolUse{
						{ID: "tool1", Name: "bash", Input: `{"cmd": "echo"}`},
					},
				},
			},
			want: []adapter.Message{
				{
					ID:   "msg1",
					Role: "assistant",
					ToolUses: []adapter.ToolUse{
						{ID: "tool1", Name: "bash", Input: `{"cmd": "echo"}`},
					},
				},
			},
			deep: true,
		},
		{
			name: "message with thinking blocks",
			msgs: []adapter.Message{
				{
					ID:   "msg1",
					Role: "assistant",
					ThinkingBlocks: []adapter.ThinkingBlock{
						{Content: "thinking...", TokenCount: 3},
					},
				},
			},
			want: []adapter.Message{
				{
					ID:   "msg1",
					Role: "assistant",
					ThinkingBlocks: []adapter.ThinkingBlock{
						{Content: "thinking...", TokenCount: 3},
					},
				},
			},
			deep: true,
		},
		{
			name: "message with content blocks",
			msgs: []adapter.Message{
				{
					ID:   "msg1",
					Role: "assistant",
					ContentBlocks: []adapter.ContentBlock{
						{Type: "text", Text: "content"},
					},
				},
			},
			want: []adapter.Message{
				{
					ID:   "msg1",
					Role: "assistant",
					ContentBlocks: []adapter.ContentBlock{
						{Type: "text", Text: "content"},
					},
				},
			},
			deep: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CopyMessages(tt.msgs)

			// Check basic equality
			if len(got) != len(tt.want) {
				t.Errorf("CopyMessages() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].ID != tt.want[i].ID {
					t.Errorf("CopyMessages()[%d].ID = %v, want %v", i, got[i].ID, tt.want[i].ID)
				}
				if got[i].Role != tt.want[i].Role {
					t.Errorf("CopyMessages()[%d].Role = %v, want %v", i, got[i].Role, tt.want[i].Role)
				}
				if got[i].Content != tt.want[i].Content {
					t.Errorf("CopyMessages()[%d].Content = %v, want %v", i, got[i].Content, tt.want[i].Content)
				}
			}

			// Check deep copy if required
			if tt.deep && len(tt.msgs) > 0 {
				// Modify original and verify copy is unchanged
				origMsg := &tt.msgs[0]
				cpyMsg := &got[0]

				// Modify ToolUses
				if origMsg.ToolUses != nil && cpyMsg.ToolUses != nil {
					origMsg.ToolUses[0].Name = "modified"
					if cpyMsg.ToolUses[0].Name == "modified" {
						t.Error("CopyMessages() did not create deep copy of ToolUses")
					}
					// Restore for other checks
					origMsg.ToolUses[0].Name = tt.want[0].ToolUses[0].Name
				}

				// Modify ThinkingBlocks
				if origMsg.ThinkingBlocks != nil && cpyMsg.ThinkingBlocks != nil {
					origMsg.ThinkingBlocks[0].Content = "modified"
					if cpyMsg.ThinkingBlocks[0].Content == "modified" {
						t.Error("CopyMessages() did not create deep copy of ThinkingBlocks")
					}
					// Restore
					origMsg.ThinkingBlocks[0].Content = tt.want[0].ThinkingBlocks[0].Content
				}

				// Modify ContentBlocks
				if origMsg.ContentBlocks != nil && cpyMsg.ContentBlocks != nil {
					origMsg.ContentBlocks[0].Text = "modified"
					if cpyMsg.ContentBlocks[0].Text == "modified" {
						t.Error("CopyMessages() did not create deep copy of ContentBlocks")
					}
					// Restore
					origMsg.ContentBlocks[0].Text = tt.want[0].ContentBlocks[0].Text
				}
			}
		})
	}
}
