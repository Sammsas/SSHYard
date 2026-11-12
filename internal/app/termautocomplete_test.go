package app

import (
	"slices"
	"strings"
	"testing"

	"github.com/Licoy/sshyard/internal/assets"
	"github.com/Licoy/sshyard/internal/session"
	"github.com/egoist/mygo/ui"
)

func TestTerminalAutocompleteSuggestionsAndCompletion(t *testing.T) {
	a, tt := newTestApp(t)
	srv, sh := startScript(t, "ready")
	h := addHost(t, a, srv, "srv-ac", "secret")
	a.known.Add(srv.Addr, srv.HostKey.PublicKey())

	if _, err := a.saveSnippet(assets.Snippet{Name: "更新软件源"}, "apt update && apt list --upgradable"); err != nil {
		t.Fatal(err)
	}

	if _, err := a.Connect(h.ID, openAt{}); err != nil {
		t.Fatal(err)
	}
	a.setSidebarHidden(true)
	cn := a.focused()
	if cn == nil || cn.tv == nil {
		t.Fatal("no focused terminal")
	}
	waitFor(t, tt, "connected", func() bool { return cn.s.Status().State == session.Connected })
	_ = sh.take()

	// Execute a command first; host history should not be matched in autocomplete.
	cn.tv.input(ui.InputEvent{Kind: ui.InputText, Text: "apt install -y htop"}, a.svc)
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyEnter}, a.svc)
	readExactly(t, tt, sh, "apt install -y htop\r")

	// Typing "apt" opens the floating autocomplete list with snippets and Linux subcommands only.
	cn.tv.input(ui.InputEvent{Kind: ui.InputText, Text: "apt"}, a.svc)
	readExactly(t, tt, sh, "apt")
	tt.Frame()
	if !cn.tv.acOpen() {
		t.Fatal("expected autocomplete popup to be open after typing 'apt'")
	}

	var cmds []string
	var kinds []termSuggestKind
	for _, item := range cn.tv.acItems {
		cmds = append(cmds, item.Cmd)
		kinds = append(kinds, item.Kind)
	}
	if slices.Contains(cmds, "apt install -y htop") {
		t.Fatalf("suggestions %v unexpectedly contained host history command", cmds)
	}
	for _, wantCmd := range []string{
		"apt update && apt list --upgradable",
		"apt update",
		"apt upgrade",
	} {
		if !slices.Contains(cmds, wantCmd) {
			t.Fatalf("suggestions %v missing %q", cmds, wantCmd)
		}
	}
	for _, wantKind := range []termSuggestKind{suggestSnippet, suggestLinux} {
		if !slices.Contains(kinds, wantKind) {
			t.Fatalf("suggestion kinds %v missing %v", kinds, wantKind)
		}
	}

	// Navigate to "apt update" with KeyDown and press Enter to complete it into the terminal.
	idx := slices.Index(cmds, "apt update")
	for range idx + 1 {
		cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyDown}, a.svc)
	}
	if cn.tv.acIndex != idx || !cn.tv.acTouched {
		t.Fatalf("acIndex = %d (want %d), acTouched = %v (want true)", cn.tv.acIndex, idx, cn.tv.acTouched)
	}
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyEnter}, a.svc)
	readExactly(t, tt, sh, " update")
	if !strings.EqualFold(cn.tv.inputBuf, "apt update") {
		t.Fatalf("inputBuf after Enter completion = %q, want %q", cn.tv.inputBuf, "apt update")
	}

	// Mouse click on a suggestion in the floating list also completes it into the terminal.
	tt.Frame()
	mustClick(t, tt, "apt update && apt list --upgradable")
	readExactly(t, tt, sh, " && apt list --upgradable")

	// Pressing Tab sends a literal tab to the shell instead of completing from the floating popup.
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyEnter}, a.svc)
	readExactly(t, tt, sh, "\r")
	cn.tv.input(ui.InputEvent{Kind: ui.InputText, Text: "apt"}, a.svc)
	readExactly(t, tt, sh, "apt")
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyTab}, a.svc)
	readExactly(t, tt, sh, "\t")
	if cn.tv.acOpen() {
		t.Fatal("autocomplete popup should reset after Tab passes through to terminal")
	}

	// Pressing Escape dismisses the floating popup.
	cn.tv.input(ui.InputEvent{Kind: ui.InputText, Text: "docker"}, a.svc)
	readExactly(t, tt, sh, "docker")
	if !cn.tv.acOpen() {
		t.Fatal("expected autocomplete popup to open after typing 'docker'")
	}
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyEscape}, a.svc)
	if cn.tv.acOpen() {
		t.Fatal("autocomplete popup still open after Escape")
	}
}
