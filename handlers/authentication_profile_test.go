package handlers

import (
	"testing"

	"gym-api/m/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildUserProfileResponseIncludesAdminFlag(t *testing.T) {
	user := models.User{ID: primitive.NewObjectID(), Email: "admin@example.com", Provider: "google"}
	profile := buildUserProfileResponse(user, []string{"member", "admin"})

	if profile["email"] != "admin@example.com" {
		t.Fatalf("expected email admin@example.com, got %v", profile["email"])
	}

	roles, ok := profile["roles"].([]string)
	if !ok {
		t.Fatalf("expected roles []string, got %T", profile["roles"])
	}
	if len(roles) != 2 || roles[1] != "admin" {
		t.Fatalf("unexpected roles payload: %#v", roles)
	}

	isAdmin, ok := profile["isAdmin"].(bool)
	if !ok {
		t.Fatalf("expected isAdmin bool, got %T", profile["isAdmin"])
	}
	if !isAdmin {
		t.Fatal("expected isAdmin to be true when admin role is present")
	}
}

func TestBuildUserProfileResponseDefaultsToMemberRole(t *testing.T) {
	user := models.User{ID: primitive.NewObjectID(), Email: "user@example.com", Provider: "google"}
	profile := buildUserProfileResponse(user, nil)

	roles, ok := profile["roles"].([]string)
	if !ok {
		t.Fatalf("expected roles []string, got %T", profile["roles"])
	}
	if len(roles) != 1 || roles[0] != "member" {
		t.Fatalf("expected default member role, got %#v", roles)
	}

	isAdmin, ok := profile["isAdmin"].(bool)
	if !ok {
		t.Fatalf("expected isAdmin bool, got %T", profile["isAdmin"])
	}
	if isAdmin {
		t.Fatal("expected isAdmin to be false without admin role")
	}
}

func TestBuildUserProfileResponseUsesGinMap(t *testing.T) {
	user := models.User{ID: primitive.NewObjectID(), Email: "user@example.com", Provider: "google"}
	profile := buildUserProfileResponse(user, []string{"admin"})
	if _, ok := profile["id"]; !ok {
		t.Fatal("expected id key to exist")
	}
	if _, ok := profile["email"]; !ok {
		t.Fatal("expected email key to exist")
	}
	if _, ok := profile["roles"]; !ok {
		t.Fatal("expected roles key to exist")
	}
	if _, ok := profile["isAdmin"]; !ok {
		t.Fatal("expected isAdmin key to exist")
	}
	_ = gin.H{}
}
