package adapterutil

import "testing"

func TestShortID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "long ID truncated to 8 chars",
			id:   "msg_abc123456789def",
			want: "msg_abc1",
		},
		{
			name: "ID exactly 8 chars",
			id:   "msg_1234",
			want: "msg_1234",
		},
		{
			name: "ID shorter than 8 chars",
			id:   "msg_1",
			want: "msg_1",
		},
		{
			name: "empty ID",
			id:   "",
			want: "",
		},
		{
			name: "ID with UUID",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			want: "550e8400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortID(tt.id); got != tt.want {
				t.Errorf("ShortID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{
			name:   "short title not truncated",
			s:      "Hello",
			maxLen: 20,
			want:   "Hello",
		},
		{
			name:   "title exactly at limit",
			s:      "Hello World",
			maxLen: 11,
			want:   "Hello World",
		},
		{
			name:   "title truncated with ellipsis",
			s:      "Hello World This Is A Long Title",
			maxLen: 20,
			want:   "Hello World This ...",
		},
		{
			name:   "title with newlines replaced",
			s:      "Line1\nLine2\nLine3",
			maxLen: 20,
			want:   "Line1 Line2 Line3",
		},
		{
			name:   "title with carriage returns removed",
			s:      "Line1\r\nLine2",
			maxLen: 20,
			want:   "Line1 Line2",
		},
		{
			name:   "title with whitespace trimmed",
			s:      "  Hello World  ",
			maxLen: 20,
			want:   "Hello World",
		},
		{
			name:   "mixed newlines and whitespace",
			s:      "  \n Hello \n World  \n ",
			maxLen: 20,
			want:   "Hello   World", // spaces preserved, just trimmed
		},
		{
			name:   "empty string",
			s:      "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "maxLen less than 3",
			s:      "Hello World",
			maxLen: 2,
			want:   "He", // truncated without ellipsis since maxLen < 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateTitle(tt.s, tt.maxLen); got != tt.want {
				t.Errorf("TruncateTitle(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}
