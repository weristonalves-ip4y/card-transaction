package sqlserver

import "testing"

func TestConvertDBBalanceToCents(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    int64
		wantErr bool
	}{
		{name: "nil becomes zero", input: nil, want: 0},
		{name: "int64 cents", input: int64(1234), want: 1234},
		{name: "float64 units", input: 9655.8, want: 965580},
		{name: "float32 units", input: float32(10.5), want: 1050},
		{name: "byte decimal units", input: []byte("9655.8"), want: 965580},
		{name: "string decimal with comma", input: "9655,8", want: 965580},
		{name: "string integer cents", input: "1234", want: 1234},
		{name: "invalid string", input: "abc", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertDBBalanceToCents(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}
