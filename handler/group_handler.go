package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chat/models"
	"chat/service"
)

type GroupHandler struct {
	groupService service.GroupService
}

func NewGroupHandler(groupService service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// GET /groups/:id
func (h *GroupHandler) GetGroupByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	group, err := h.groupService.GetGroupByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

// GET /groups
func (h *GroupHandler) GetAllGroups(c *gin.Context) {
	groups, err := h.groupService.GetAllGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// PUT /groups/:id
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required"`
		RequesterID uint   `json:"requester_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupService.UpdateGroup(uint(id), req.RequesterID, req.Name); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNotMember || err == service.ErrNotOwner {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group updated"})
}

// PATCH /groups/:id/members/:user_id
func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		Role        string `json:"role" binding:"required"`
		RequesterID uint   `json:"requester_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: must be owner, admin, or member"})
		return
	}
	if err := h.groupService.UpdateMemberRole(uint(groupID), uint(userID), req.RequesterID, req.Role); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNotMember || err == service.ErrInsufficientRole || err == service.ErrNotOwner {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member role updated"})
}

// POST /groups/:id/leave
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupService.LeaveGroup(uint(groupID), req.UserID); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNotMember || err == service.ErrNotOwner {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "left group"})
}

// POST /groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.groupService.CreateGroup(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
}

// DELETE /groups/:id
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		RequesterID uint `json:"requester_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupService.DeleteGroup(uint(id), req.RequesterID); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNotMember || err == service.ErrNotOwner {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}

// validRoles contains the allowed group member role values.
var validRoles = map[string]bool{
	models.RoleOwner:  true,
	models.RoleAdmin:  true,
	models.RoleMember: true,
}

// POST /groups/:id/members
func (h *GroupHandler) AddMember(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		RequesterID uint   `json:"requester_id" binding:"required"`
		UserID      uint   `json:"user_id" binding:"required"`
		Role        string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role if provided
	if req.Role != "" && !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role: must be owner, admin, or member"})
		return
	}

	if err := h.groupService.AddMember(uint(groupID), req.UserID, req.Role, req.RequesterID); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNotMember || err == service.ErrInsufficientRole || err == service.ErrNotOwner {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member added"})
}

// DELETE /groups/:id/members/:user_id
func (h *GroupHandler) RemoveMember(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req struct {
		RequesterID uint `json:"requester_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupService.RemoveMember(uint(groupID), uint(userID), req.RequesterID); err != nil {
		code := http.StatusInternalServerError
		if err == service.ErrNotMember || err == service.ErrInsufficientRole || err == service.ErrNotOwner {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

// GET /groups/:id/members
func (h *GroupHandler) GetMembers(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	members, err := h.groupService.GetMembers(uint(groupID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

