package utils

import "testing"

func TestGetAvailableUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "github",
			args: args{
				url: "https://github.com",
			},
			want: "https://github.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAvailableUrl(tt.args.url); got != tt.want {
				t.Errorf("GetAvailableUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
