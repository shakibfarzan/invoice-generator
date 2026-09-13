package pdf

import "testing"

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
		want   string
	}{
		{
			name:   "zero",
			amount: 0,
			want:   "0",
		},
		{
			name:   "single group",
			amount: 500,
			want:   "500",
		},
		{
			name:   "exactly three digits",
			amount: 999,
			want:   "999",
		},
		{
			name:   "one separator",
			amount: 1000,
			want:   "1,000",
		},
		{
			name:   "two separators",
			amount: 1234500,
			want:   "1,234,500",
		},
		{
			name:   "three separators",
			amount: 1234567890,
			want:   "1,234,567,890",
		},
		{
			name:   "negative amount",
			amount: -1234500,
			want:   "-1,234,500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatMoney(tt.amount); got != tt.want {
				t.Errorf("FormatMoney(%d) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name    string
		percent float32
		want    string
	}{
		{
			name:    "zero",
			percent: 0,
			want:    "0٪",
		},
		{
			name:    "whole number",
			percent: 10,
			want:    "10٪",
		},
		{
			name:    "fraction",
			percent: 12.5,
			want:    "12.5٪",
		},
		{
			name:    "hundred",
			percent: 100,
			want:    "100٪",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPercent(tt.percent); got != tt.want {
				t.Errorf("FormatPercent(%v) = %q, want %q", tt.percent, got, tt.want)
			}
		})
	}
}
