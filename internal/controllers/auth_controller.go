package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/meeting-minutes-ai/internal/dto"
	"github.com/yourusername/meeting-minutes-ai/internal/services"
)

// AuthController handles authentication endpoints
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register handles user registration
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.APIResponse
// @Router /api/auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	resp, err := c.authService.Register(&req)
	if err != nil {
		ctx.JSON(http.StatusConflict, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: "Registrasi berhasil",
		Data:    resp,
	})
}

// Login handles user login
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.APIResponse
// @Router /api/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	resp, err := c.authService.Login(&req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.APIResponse{
			Success: false,
			Error:   "Email atau password salah",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: "Login berhasil",
		Data:    resp,
	})
}