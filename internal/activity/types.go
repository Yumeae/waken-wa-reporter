package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ReportRequest matches the activity API JSON body.
type ReportRequest struct {
	GeneratedHashKey string         `json:"generatedHashKey"`
	Device           string         `json:"device,omitempty"`
	DeviceName       string         `json:"device_name,omitempty"`
	DeviceType       string         `json:"device_type,omitempty"`
	ProcessName      string         `json:"process_name"`
	ProcessTitle     string         `json:"process_title,omitempty"`
	BatteryLevel     *int           `json:"battery_level,omitempty"`
	PushMode         string         `json:"push_mode,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

// pendingResponse matches the server JSON when the device is awaiting admin approval (HTTP 202).
type pendingResponse struct {
	Success     bool   `json:"success"`
	Pending     bool   `json:"pending"`
	ApprovalURL string `json:"approvalUrl"`
	Error       string `json:"error"`
}

// PendingApprovalError is returned when POST /api/activity responds 202 (device not yet approved).
type PendingApprovalError struct {
	ApprovalURL string
	Message     string
}

func (e *PendingApprovalError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return fmt.Sprintf("activity: pending approval: %s", e.Message)
	}
	return "activity: pending approval"
}

// Client posts activity events to the configured server.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}
