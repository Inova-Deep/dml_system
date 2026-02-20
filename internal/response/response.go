package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
)

// Validate is the global singleton for struct validation
var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// ErrorResponse represents a standard API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// ValidationErrorResponse represents a detailed validation error
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details"`
}

// JSON marshals data to JSON and writes it to the response writer
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// Error sends a standardized error response
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}

// ValidationError formats go-playground validator errors into a readable map
func ValidationError(w http.ResponseWriter, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errMap := make(map[string]string)
		for _, e := range validationErrors {
			// e.g. "EmployeeNo": "required" or "Email": "email"
			errMap[e.Field()] = fmt.Sprintf("Failed validation on '%s' tag", e.Tag())
		}
		JSON(w, http.StatusBadRequest, ValidationErrorResponse{
			Error:   "Validation failed",
			Details: errMap,
		})
		return
	}

	// Fallback
	Error(w, http.StatusBadRequest, "Invalid request payload")
}

// DBError maps PostgreSQL errors to standard HTTP responses
func DBError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			Error(w, http.StatusConflict, "A record with this value already exists")
			return
		case "23503": // foreign_key_violation
			Error(w, http.StatusBadRequest, "Invalid reference to a related record")
			return
		}
	}
	// Log the actual error for debugging, but hide it from the client
	log.Printf("Database error: %v", err)
	Error(w, http.StatusInternalServerError, "Internal server error")
}

// PaginatedJSON marshals data and pagination metadata to JSON and writes it
func PaginatedJSON(w http.ResponseWriter, status int, data interface{}, page int, size int, total int) {
	payload := map[string]interface{}{
		"data": data,
		"metadata": map[string]interface{}{
			"currentPage": page,
			"pageSize":    size,
			"totalCount":  total,
			"totalPages":  (total + size - 1) / size,
		},
	}
	JSON(w, status, payload)
}
