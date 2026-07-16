package sqlserver

import "testing"

func TestGetNullableStringWithStringPointer(t *testing.T) {
	value := "card-abc-123"
	payload := map[string]any{"card_id": &value}

	got := getNullableString(payload, "card_id")
	if got != value {
		t.Fatalf("expected %q, got %#v", value, got)
	}
}

func TestGetNullableStringWithNilAndEmptyPointer(t *testing.T) {
	var nilValue *string
	empty := ""

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{
			name:    "nil pointer",
			payload: map[string]any{"card_id": nilValue},
		},
		{
			name:    "empty pointer",
			payload: map[string]any{"card_id": &empty},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getNullableString(tc.payload, "card_id")
			if got != nil {
				t.Fatalf("expected nil, got %#v", got)
			}
		})
	}
}
