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
	cn := a.focused()
	if cn == nil || cn.tv == nil {
		t.Fatal("no focused terminal")
	}
	waitFor(t, tt, "connected", func() bool { return cn.s.Status().State == session.Connected })
	_ = sh.take()

	// Execute a command first so it enters the current host's command history.
	cn.tv.input(ui.InputEvent{Kind: ui.InputText, Text: "apt install -y htop"}, a.svc)
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyEnter}, a.svc)
	readExactly(t, tt, sh, "apt install -y htop\r")

	if got := a.hostHistory[h.ID]; !slices.Contains(got, "apt install -y htop") {
		t.Fatalf("host history = %v, want apt install -y htop", got)
	}

	// Typing "apt" opens the floating autocomplete list with snippet, history, and Linux subcommands.
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
	for _, wantCmd := range []string{
		"apt update && apt list --upgradable",
		"apt install -y htop",
		"apt update",
		"apt upgrade",
	} {
		if !slices.Contains(cmds, wantCmd) {
			t.Fatalf("suggestions %v missing %q", cmds, wantCmd)
		}
	}
	for _, wantKind := range []termSuggestKind{suggestSnippet, suggestHistory, suggestLinux} {
		if !slices.Contains(kinds, wantKind) {
			t.Fatalf("suggestion kinds %v missing %v", kinds, wantKind)
		}
	}

	// The floating suggestion list renders in the UI with "apt update".
	if !tt.HasText("apt update") {
		t.Fatalf("floating autocomplete UI missing 'apt update'; texts: %v", tt.Texts())
	}

	// Navigate to "apt update" and press Tab to complete it.
	idx := slices.Index(cmds, "apt update")
	for range idx {
		cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyDown}, a.svc)
	}
	if cn.tv.acIndex != idx {
		t.Fatalf("acIndex = %d, want %d", cn.tv.acIndex, idx)
	}
	cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyTab}, a.svc)
	readExactly(t, tt, sh, " update")
	if !strings.EqualFold(cn.tv.inputBuf, "apt update") {
		t.Fatalf("inputBuf after Tab = %q, want %q", cn.tv.inputBuf, "apt update")
	}

	// Pressing Escape dismisses the floating popup.
	if cn.tv.acOpen() {
		cn.tv.input(ui.InputEvent{Kind: ui.InputKeyDown, Key: ui.KeyEscape}, a.svc)
		if cn.tv.acOpen() {
			t.Fatal("autocomplete popup still open after Escape")
		}
	}
}
