package dto

import (
	"testing"
)

func TestAuthorizePurchaseRequestResolveEstablishmentLocationPrefersISO8583(t *testing.T) {
	t.Parallel()

	isoLocation := "  MERCEARIA CENTRAL  "
	establishmentName := "PADARIA BAIRRO"
	request := AuthorizePurchaseRequest{
		OriginalIso8583: OriginalIso8583{RequestCardAcceptorNameLocation: &isoLocation},
		Establishment:   EstablishmentInput{Name: &establishmentName},
	}

	resolved := request.ResolveEstablishmentLocation()
	if resolved != "MERCEARIA CENTRAL" {
		t.Fatalf("expected ISO8583 location, got %q", resolved)
	}
}

func TestAuthorizePurchaseRequestResolveEstablishmentLocationFallbackEstablishment(t *testing.T) {
	t.Parallel()

	establishmentName := "  PADARIA BAIRRO  "
	request := AuthorizePurchaseRequest{
		OriginalIso8583: OriginalIso8583{},
		Establishment:   EstablishmentInput{Name: &establishmentName},
	}

	resolved := request.ResolveEstablishmentLocation()
	if resolved != "PADARIA BAIRRO" {
		t.Fatalf("expected establishment name fallback, got %q", resolved)
	}
}
