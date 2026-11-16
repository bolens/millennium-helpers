//go:build unix

package steam

import "testing"

func TestRelaunchArgumentsDropsOnlyLoggerTransportMarker(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"empty", "", ""},
		{"captured", "'-srt-logger-opened' '-silent' '' ", "'-silent' ''"},
		{"legacy", "-silent -srt-logger-opened --example", "-silent --example"},
		{"quoted", `"-srt-logger-opened" '-silent'`, "'-silent'"},
		{"carriage-return", "'-srt-logger-opened' name\rvalue", "name\rvalue"},
		{"duplicates", "'-srt-logger-opened' -srt-logger-opened", ""},
		{"prefix", "'-srt-logger-opened-extra' '-silent'", "'-srt-logger-opened-extra' '-silent'"},
		{"embedded", `"message -srt-logger-opened suffix" 'x'`, `"message -srt-logger-opened suffix" 'x'`},
		{"escaped", `'-srt-logger-opened' 'quote'\''arg' '$(literal)' ';literal'`, `'quote'\''arg' '$(literal)' ';literal'`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := relaunchArguments(tc.input)
			if err != nil || got != tc.want {
				t.Fatalf("got %q %v, want %q", got, err, tc.want)
			}
			again, err := relaunchArguments(got)
			if err != nil || again != got {
				t.Fatalf("not idempotent: %q %v", again, err)
			}
		})
	}
	for _, input := range []string{"'unterminated", `"unterminated`, "trailing\\"} {
		if _, err := relaunchArguments(input); err == nil {
			t.Fatalf("accepted invalid saved args %q", input)
		}
	}
}
