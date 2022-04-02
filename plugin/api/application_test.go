package api

import "testing"

func TestValidateApplicationName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"valid app name app", args{name: "app"}, "app"},
		{"valid app name 123", args{name: "123"}, "123"},
		{"valid app name app_12", args{name: "app_12"}, "app_12"},
		{"valid app name empty", args{name: ""}, DefaultApplicationName},
		{"valid app name +++", args{name: "+++"}, DefaultApplicationName},
		{"valid app name +a+a+", args{name: "+a+a+"}, "aa"},
		{"valid app name b+a+a+b", args{name: "b+a+a+b"}, "baab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeSpecialCharsForValidApplicationName(tt.args.name); got != tt.want {
				t.Errorf("ValidateApplicationName() = %v, want %v", got, tt.want)
			}
		})
	}
}
