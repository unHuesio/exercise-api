package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBuildAuthPayloadIncludesAdminFlag(t *testing.T) {
	payload := buildAuthPayload("abc123", []string{"member", "admin"})

	if payload["token"] != "abc123" {
		t.Fatalf("expected token to be abc123, got %v", payload["token"])
	}

	roles, ok := payload["roles"].([]string)
	if !ok {
		t.Fatalf("expected roles []string, got %T", payload["roles"])
	}
	if len(roles) != 2 || roles[0] != "member" || roles[1] != "admin" {
		t.Fatalf("unexpected roles payload: %#v", roles)
	}

	isAdmin, ok := payload["isAdmin"].(bool)
	if !ok {
		t.Fatalf("expected isAdmin bool, got %T", payload["isAdmin"])
	}
	if !isAdmin {
		t.Fatal("expected admin flag to be true when admin role is present")
	}
}

func TestBuildAuthPayloadDefaultsToMemberRole(t *testing.T) {
	payload := buildAuthPayload("def456", nil)

	roles, ok := payload["roles"].([]string)
	if !ok {
		t.Fatalf("expected roles []string, got %T", payload["roles"])
	}
	if len(roles) != 1 || roles[0] != "member" {
		t.Fatalf("expected default member role, got %#v", roles)
	}

	isAdmin, ok := payload["isAdmin"].(bool)
	if !ok {
		t.Fatalf("expected isAdmin bool, got %T", payload["isAdmin"])
	}
	if isAdmin {
		t.Fatal("expected admin flag to be false when no admin role is present")
	}
}

func TestBuildAuthPayloadUsesGinMap(t *testing.T) {
	payload := buildAuthPayload("token", []string{"admin"})
	if _, ok := payload["token"]; !ok {
		t.Fatal("expected token key to exist")
	}
	if _, ok := payload["roles"]; !ok {
		t.Fatal("expected roles key to exist")
	}
	if _, ok := payload["isAdmin"]; !ok {
		t.Fatal("expected isAdmin key to exist")
	}
	_ = gin.H{}
}
