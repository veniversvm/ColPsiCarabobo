// api/internal/request_structs/requests_settings.structs.go
package request_structs

// UpdateReceptionRequest actualiza un interruptor de recepción global.
// Solo el Sudo puede invocarlo.
type UpdateReceptionRequest struct {
	Key     string `json:"key" validate:"required"`
	Enabled *bool  `json:"enabled" validate:"required"`
	Message string `json:"message"`
}