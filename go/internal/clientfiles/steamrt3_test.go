package clientfiles

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLinux64IntegrityAndModes(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "version.txt"), []byte("3.6.3\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, f := range Linux() {
		if err := os.WriteFile(filepath.Join(root, f.Name), []byte("library"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := WriteChecksums(root); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLinux64(root); err == nil {
		t.Fatal("legacy install compatible with SteamRT3")
	}
	for i, f := range Linux64() {
		if err := os.WriteFile(filepath.Join(root, f.Name), []byte("new library"), 0644); err != nil {
			t.Fatal(err)
		}
		if i < 2 {
			if err := ValidateFiles(root); err == nil {
				t.Fatal("partial x64 install accepted")
			}
		}
	}
	if err := ValidateLinux64(root); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if HelpersExecutable(root) {
			t.Fatal("missed non-executable x64 Lua helper")
		}
		if err := NormalizeHelpers(root); err != nil {
			t.Fatal(err)
		}
		if !HelpersExecutable(root) {
			t.Fatal("x64 helper mode not repaired")
		}
	}
	if err := WriteChecksums(root); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, Linux64()[1].Name), []byte("corrupt"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(root); err == nil {
		t.Fatal("x64 corruption missed")
	}
	for _, f := range Linux64() {
		if err := os.Remove(filepath.Join(root, f.Name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := Verify(root); err == nil {
		t.Fatal("removing entire x64 group bypassed integrity")
	}

}
