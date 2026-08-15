package surrealdb

import "testing"

func TestHTTPEndpointForGraph(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"ws://surrealdb:8000", "http://surrealdb:8000"},
		{"wss://example.com:8000/rpc", "https://example.com:8000/rpc"},
		{"http://surrealdb:8000", "http://surrealdb:8000"},
		{"https://example.com", "https://example.com"},
	}
	for _, tc := range cases {
		got, err := httpEndpointForGraph(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
	if _, err := httpEndpointForGraph("ftp://nope"); err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}
