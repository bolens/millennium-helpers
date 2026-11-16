package jsonobject

import (
	"encoding/json"
	"testing"
)

func TestDecodePreservesIntegerIdentityAndRejectsNonDocuments(t *testing.T) {
	got, err := Decode([]byte(`{"id":9007199254740993}`))
	if err != nil {
		t.Fatal(err)
	}
	number, ok := got["id"].(json.Number)
	if !ok || number.String() != "9007199254740993" {
		t.Fatalf("id=%v", got["id"])
	}
	for _, input := range []string{"null", "[]", "true", "42", `"text"`, "{} {}", "{} junk", "{"} {
		if _, err := Decode([]byte(input)); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}
