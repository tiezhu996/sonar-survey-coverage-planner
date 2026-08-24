package model

import "testing"

func TestUserActiveDefaultAndCanLogin(t *testing.T) {
	user := User{Active: true}
	if !user.CanLogin() {
		t.Fatal("active user should be able to login")
	}
	disabled := User{Active: false}
	if disabled.CanLogin() {
		t.Fatal("disabled user should not be able to login")
	}
}
