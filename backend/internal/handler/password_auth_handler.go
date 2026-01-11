package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/your-org/go-nextjs-oauth-template/backend/internal/service"
)

// PasswordAuthHandler handles password authentication endpoints.
type PasswordAuthHandler struct {
	passwordService *service.PasswordAuthService
	logger          *slog.Logger
}

// NewPasswordAuthHandler creates a new password auth handler.
func NewPasswordAuthHandler(
	passwordService *service.PasswordAuthService,
	logger *slog.Logger,
) *PasswordAuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PasswordAuthHandler{
		passwordService: passwordService,
		logger:          logger,
	}
}

// RegisterRoutes registers the password auth routes.
func (h *PasswordAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/forgot-password", h.ForgotPassword)
	mux.HandleFunc("POST /api/auth/reset-password", h.ResetPassword)
	mux.HandleFunc("POST /api/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/auth/resend-verification", h.ResendVerification)
}

// RegisterRequestBody represents the register request body.
type RegisterRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /api/auth/register
func (h *PasswordAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body RegisterRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Email == "" || body.Password == "" {
		h.writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.passwordService.Register(r.Context(), service.RegisterRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		h.logger.Error("registration failed", "error", err)
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": result.Message,
	})
}

// LoginRequestBody represents the login request body.
type LoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponseBody represents the login response body.
type LoginResponseBody struct {
	AccountID         string `json:"account_id"`
	NeedsVerification bool   `json:"needs_verification"`
	Message           string `json:"message"`
}

// Login handles POST /api/auth/login
func (h *PasswordAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Email == "" || body.Password == "" {
		h.writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.passwordService.Login(r.Context(), service.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Note: In a real application, you would generate JWT tokens here
	// and return them in the response or set them as cookies.
	h.writeJSON(w, http.StatusOK, LoginResponseBody{
		AccountID:         result.AccountID.String(),
		NeedsVerification: result.NeedsVerification,
		Message:           "Login successful",
	})
}

// ForgotPasswordRequestBody represents the forgot password request body.
type ForgotPasswordRequestBody struct {
	Email string `json:"email"`
}

// ForgotPassword handles POST /api/auth/forgot-password
func (h *PasswordAuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body ForgotPasswordRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Email == "" {
		h.writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Always return success to prevent email enumeration
	_ = h.passwordService.ForgotPassword(r.Context(), service.ForgotPasswordRequest{
		Email: body.Email,
	})

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "If this email is registered, you will receive a password reset link.",
	})
}

// ResetPasswordRequestBody represents the reset password request body.
type ResetPasswordRequestBody struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles POST /api/auth/reset-password
func (h *PasswordAuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body ResetPasswordRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Token == "" || body.NewPassword == "" {
		h.writeError(w, http.StatusBadRequest, "token and new_password are required")
		return
	}

	if err := h.passwordService.ResetPassword(r.Context(), service.ResetPasswordRequest{
		Token:       body.Token,
		NewPassword: body.NewPassword,
	}); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Password has been reset successfully.",
	})
}

// VerifyEmailRequestBody represents the verify email request body.
type VerifyEmailRequestBody struct {
	Token string `json:"token"`
}

// VerifyEmail handles POST /api/auth/verify-email
func (h *PasswordAuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var body VerifyEmailRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Token == "" {
		h.writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	if err := h.passwordService.VerifyEmail(r.Context(), body.Token); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully.",
	})
}

// ResendVerificationRequestBody represents the resend verification request body.
type ResendVerificationRequestBody struct {
	Email string `json:"email"`
}

// ResendVerification handles POST /api/auth/resend-verification
func (h *PasswordAuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var body ResendVerificationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Email == "" {
		h.writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Always return success to prevent email enumeration
	_ = h.passwordService.ResendVerificationEmail(r.Context(), body.Email)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "If this email is registered and unverified, a new verification email has been sent.",
	})
}

// helper functions

func (h *PasswordAuthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *PasswordAuthHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": message,
	})
}
