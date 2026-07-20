package dto

import (
	"strings"
)

// ResolveEstablishmentLocation centralizes merchant location resolution
// from ISO8583 first, then falls back to establishment name.
func (r AuthorizeRequest) ResolveEstablishmentLocation() string {
	if location := trimPointerString(r.OriginalIso8583.RequestCardAcceptorNameLocation); location != "" {
		return location
	}

	return trimPointerString(r.Establishment.Name)
}

func trimPointerString(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}
