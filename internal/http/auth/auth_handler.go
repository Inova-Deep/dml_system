package auth

import (
	"encoding/json"
	"net/http"

	logic "github.com/INOVA/DML/internal/logic/auth"
	"github.com/INOVA/DML/internal/response"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	service *logic.AuthService
}

func NewAuthHandler(service *logic.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/login", h.HandleLogin)
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the token payload
type LoginResponse struct {
	Token string `json:"token"`
}

// HandleLogin godoc
// @Summary      Login and get JWT token
// @Description  Authenticates a user via email and password and returns a JWT token for Authorization.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login credentials"
// @Success      200      {object}  LoginResponse "Successfully authenticated"
// @Failure      400      {object}  map[string]interface{} "Bad request payload"
// @Failure      401      {object}  map[string]interface{} "Invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := response.Validate.Struct(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	token, err := h.service.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
