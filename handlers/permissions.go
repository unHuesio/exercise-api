package handlers

import (
	"context"
	"net/http"
	"time"

	"gym-api/m/models"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PermissionHandler struct {
	DB       *mongo.Client
	Enforcer *casbin.Enforcer
}

func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	ok, err := h.Enforcer.GetPolicy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ok)
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var permission models.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, err := h.Enforcer.AddPolicy(permission.Subject, permission.Object, permission.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission already exists"})
		return
	}
	c.JSON(http.StatusOK, permission)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	var permission models.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, err := h.Enforcer.RemovePolicy(permission.Subject, permission.Object, permission.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted"})
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var permission models.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.DB.Database("gym-app").Collection("permissions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": permission})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

func (h *PermissionHandler) GetPermissionsBySubject(c *gin.Context) {
	subject := c.Param("subject")

	permissions, err := h.Enforcer.GetFilteredPolicy(0, subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, permissions)
}

func (h *PermissionHandler) AssignUserToRole(c *gin.Context) {
	var group models.GroupingPolicy
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, err := h.Enforcer.AddGroupingPolicy(group.User, group.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already assigned to role"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *PermissionHandler) GetRoles(c *gin.Context) {
	roles, err := h.Enforcer.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make(map[string][]string)
	for _, role := range roles {
		users, err := h.Enforcer.GetUsersForRole(role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result[role] = users
	}
	c.JSON(http.StatusOK, result)
}

func (h *PermissionHandler) GetRolesByUser(c *gin.Context) {
	user := c.Param("user")

	roles, err := h.Enforcer.GetRolesForUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *PermissionHandler) RemoveUserFromRole(c *gin.Context) {
	var group models.GroupingPolicy
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, err := h.Enforcer.RemoveGroupingPolicy(group.User, group.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not assigned to role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User removed from role"})
}
