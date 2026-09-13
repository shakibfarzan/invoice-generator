package main

import "testing"

func TestSanitizeFilePart(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain number",
			in:   "10023",
			want: "10023",
		},
		{
			name: "keeps dots and dashes",
			in:   "INV-2026.001",
			want: "INV-2026.001",
		},
		{
			name: "spaces become dashes",
			in:   "INV 10023",
			want: "INV-10023",
		},
		{
			name: "slashes cannot escape the directory",
			in:   "../../etc/passwd",
			want: "etc-passwd",
		},
		{
			name: "dot dot is not a path",
			in:   "..",
			want: "unknown",
		},
		{
			name: "empty falls back",
			in:   "",
			want: "unknown",
		},
		{
			name: "only separators falls back",
			in:   "---",
			want: "unknown",
		},
		{
			name: "persian invoice number is kept",
			in:   "فاکتور-۱۰۰۲۳",
			want: "فاکتور-۱۰۰۲۳",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilePart(tt.in); got != tt.want {
				t.Errorf("sanitizeFilePart(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
