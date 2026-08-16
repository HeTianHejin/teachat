package dao

import "testing"

func TestUserNameHelpers(t *testing.T) {
	user := User{FamilyName: "李", GivenName: "四", AliasName: "阿四"}

	if got := user.DisplayName(); got != "阿四" {
		t.Fatalf("DisplayName() = %q, want %q", got, "阿四")
	}

	if got := user.FullName(); got != "李四" {
		t.Fatalf("FullName() = %q, want %q", got, "李四")
	}

	user.AliasName = ""
	if got := user.DisplayName(); got != "李四" {
		t.Fatalf("DisplayName() fallback = %q, want %q", got, "李四")
	}

	user.FamilyName = ""
	user.GivenName = ""
	user.Name = "旧名"
	if got := user.FullName(); got != "旧名" {
		t.Fatalf("FullName() fallback = %q, want %q", got, "旧名")
	}
}
