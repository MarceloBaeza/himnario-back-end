package validations

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/validation"
)

type Header struct {
	Country     string
	EventId     string
	ContentType string
	Accept      string
	XClient     string
}
type HeadersValidator struct{}

func NewHeadersValidator() *HeadersValidator {
	return &HeadersValidator{}
}

func (v *HeadersValidator) Validate(h http.Header) *validation.Result {
	hdr := Header{
		Country:     strings.TrimSpace(h.Get("country")),
		EventId:     strings.TrimSpace(h.Get("event-id")),
		ContentType: strings.TrimSpace(h.Get("Content-Type")),
		Accept:      strings.TrimSpace(h.Get("Accept")),
		XClient:     strings.TrimSpace(h.Get("x-client")),
	}

	// required
	if hdr.Country == "" {
		return validation.CreateInvalid("headers/country", "country header is required")
	}
	if hdr.EventId == "" {
		return validation.CreateInvalid("headers/x-uuid", "x-uuid header is required")
	}
	if hdr.ContentType == "" {
		return validation.CreateInvalid("headers/content-type", "Content-Type header is required")
	}
	if hdr.Accept == "" {
		return validation.CreateInvalid("headers/accept", "Accept header is required")
	}
	if hdr.XClient == "" {
		return validation.CreateInvalid("headers/x-client", "x-client header is required")
	}
	if !slices.Contains(property.GetValidationsProperty().Validations.Headers.XClient, hdr.XClient) {
		return validation.CreateInvalid("headers/x-client", fmt.Sprintf("header not valid %s", hdr.XClient))
	}

	// constraints
	if hdr.Country != "CL" {
		return validation.CreateInvalid("headers/country", "country must be CL")
	}

	ct := strings.ToLower(hdr.ContentType)
	if !strings.HasPrefix(ct, "application/json") {
		return validation.CreateInvalid("headers/content-type", "Content-Type must be application/json")
	}

	ac := strings.ToLower(hdr.Accept)
	if ac != "*/*" && !strings.Contains(ac, "application/json") {
		return validation.CreateInvalid("headers/accept", "Accept must allow application/json")
	}

	return validation.CreateValid()
}
