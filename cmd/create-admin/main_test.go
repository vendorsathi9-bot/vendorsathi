package main

import "testing"

func TestValidateBootstrapInput(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		password string
		email    string
		wantErr  bool
	}{
		{
			name:     "valid input",
			phone:    "+919876543210",
			password: "A-strong-unique-password-42",
			email:    "admin@example.com",
		},
		{
			name:     "valid input without email",
			phone:    "+919876543210",
			password: "A-strong-unique-password-42",
		},
		{
			name:     "missing phone",
			password: "A-strong-unique-password-42",
			wantErr:  true,
		},
		{
			name:     "phone without country code",
			phone:    "9876543210",
			password: "A-strong-unique-password-42",
			wantErr:  true,
		},
		{
			name:     "short password",
			phone:    "+919876543210",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "documented demo password",
			phone:    "+919876543210",
			password: "AdminPassword123!",
			wantErr:  true,
		},
		{
			name:     "invalid email",
			phone:    "+919876543210",
			password: "A-strong-unique-password-42",
			email:    "not-an-email",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBootstrapInput(tt.phone, tt.password, tt.email)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateBootstrapInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
