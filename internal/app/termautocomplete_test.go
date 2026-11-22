package app

import (
	"slices"
	"strings"
	"testing"

	"github.com/Licoy/sshyard/internal/assets"
	"github.com/Licoy/sshyard/internal/cmdline"
	"github.com/Licoy/sshyard/internal/session"
	"github.com/Licoy/sshyard/internal/settings"
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

func TestTerminalAutocompleteCoexistsWithSuggest(t *testing.T) {
	// When upstream's path list is open alongside acItems, Down/Enter drive the path list.
	r := newSuggestRig(settings.CompletionAuto)
	r.src.es = []cmdline.Entry{{Name: "Makefile"}, {Name: "main.go"}}
	r.text("cat ")
	r.frame(true)
	if !r.tv.suggest.list.shows() {
		t.Fatal("expected upstream path list to be open")
	}
	r.tv.acItems = []termSuggestion{{Kind: suggestLinux, Cmd: "cat /etc/hosts"}}
	r.tv.acDismissed = false
	r.key(ui.KeyDown, 0)
	n := len(r.out)
	r.key(ui.KeyEnter, 0)
	if got := r.sentSince(n); got != "Makefile" {
		t.Fatalf("path list Enter sent %q, want %q", got, "Makefile")
	}
	if r.tv.acOpen() {
		t.Fatal("expected custom autocomplete to reset after path list completion")
	}

	// When only inline ghost text and custom autocomplete are active, Right accepts ghost text.
	r = newSuggestRig(settings.CompletionInline)
	r.src.cmds = []string{"git status"}
	r.text("git s")
	r.frame(false)
	if !r.tv.acOpen() || r.tv.suggest.ghost != "tatus" {
		t.Fatalf("acOpen=%v ghost=%q, want both active", r.tv.acOpen(), r.tv.suggest.ghost)
	}
	n = len(r.out)
	r.key(ui.KeyRight, 0)
	if got := r.sentSince(n); got != "tatus" {
		t.Fatalf("Right sent %q, want %q", got, "tatus")
	}
	if r.tv.acOpen() {
		t.Fatal("expected custom autocomplete to reset after accepting ghost suggestion")
	}
}
