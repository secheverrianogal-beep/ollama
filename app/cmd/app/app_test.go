//go:build windows || darwin

package main

import (
	"errors"
	"testing"

	"github.com/ollama/ollama/app/store"
)


func TestSetContextMenuItemsReleasesReplacedLabels(t *testing.T) {
	clearContextMenuItems()
	t.Cleanup(clearContextMenuItems)

	items := []map[string]interface{}{
		{"label": "Rename", "enabled": true},
		{"label": "Delete", "enabled": true},
	}
	if err := setContextMenuItems(items); err != nil {
		t.Fatal(err)
	}
	if len(menuItems) != 2 || menuItems[0].label == nil || menuItems[1].label == nil {
		t.Fatalf("menu items after first set = %+v", menuItems)
	}

	// Native menu consumers request the Go-backed array before copying labels.
	if menu_get_items() == nil || !menuItemsPinned {
		t.Fatal("menu item array was not pinned for the native consumer")
	}
	// A repeated read must not pin the same backing array repeatedly.
	if menu_get_items() == nil || !menuItemsPinned {
		t.Fatal("menu item array was not available on repeated read")
	}

	if err := setContextMenuItems([]map[string]interface{}{{"label": "Open", "enabled": true}}); err != nil {
		t.Fatal(err)
	}
	if menuItemsPinned {
		t.Fatal("previous menu slice remained pinned after replacement")
	}
	if len(menuItems) != 1 || menuItems[0].label == nil {
		t.Fatalf("replacement menu items = %+v", menuItems)
	}

	// Replacing the menu with an empty list must release the final label too.
	if err := setContextMenuItems(nil); err != nil {
		t.Fatal(err)
	}
	if len(menuItems) != 0 || menuItemsPinned {
		t.Fatalf("menu state after empty replacement = len %d, pinned %v", len(menuItems), menuItemsPinned)
	}
}

func TestSetContextMenuItemsCleansPartialAllocationOnInvalidItem(t *testing.T) {
	clearContextMenuItems()
	t.Cleanup(clearContextMenuItems)

	err := setContextMenuItems([]map[string]interface{}{
		{"label": "Rename", "enabled": true},
		{"label": 42, "enabled": true},
	})
	if err == nil {
		t.Fatal("expected invalid label type to return an error")
	}
	if len(menuItems) != 0 || menuItemsPinned {
		t.Fatalf("partial menu state was retained after error: len %d, pinned %v", len(menuItems), menuItemsPinned)
	}
}

func TestClearContextMenuItemsIsIdempotent(t *testing.T) {
	if err := setContextMenuItems([]map[string]interface{}{{"label": "Rename"}}); err != nil {
		t.Fatal(err)
	}
	if menu_get_items() == nil {
		t.Fatal("menu item array is nil")
	}

	clearContextMenuItems()
	clearContextMenuItems()
	if len(menuItems) != 0 || menuItemsPinned {
		t.Fatalf("menu state after repeated clear = len %d, pinned %v", len(menuItems), menuItemsPinned)
	}
}

func TestSetContextMenuItemsRepeatedReplacement(t *testing.T) {
	clearContextMenuItems()
	t.Cleanup(clearContextMenuItems)

	for range 10000 {
		if err := setContextMenuItems([]map[string]interface{}{
			{"label": "Rename", "enabled": true},
			{"label": "Delete", "enabled": true},
		}); err != nil {
			t.Fatal(err)
		}
		if err := setContextMenuItems(nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(menuItems) != 0 || menuItemsPinned {
		t.Fatalf("menu state after repeated replacement = len %d, pinned %v", len(menuItems), menuItemsPinned)
	}
}

func TestShouldShowOnboarding(t *testing.T) {
	tests := []struct {
		name     string
		settings store.Settings
		err      error
		want     bool
	}{
		{
			name:     "fresh install",
			settings: store.Settings{OnboardingVersion: 0},
			want:     true,
		},
		{
			name:     "completed onboarding",
			settings: store.Settings{OnboardingVersion: store.CurrentOnboardingVersion},
			want:     false,
		},
		{
			name: "settings failure",
			err:  errors.New("settings unavailable"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldShowOnboarding(tt.settings, tt.err); got != tt.want {
				t.Fatalf("shouldShowOnboarding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDispatchURLSchemeRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     string
		wantConnect bool
		wantOpen    bool
		wantApps    bool
		wantErr     bool
	}{
		{name: "bare URL opens app", request: "ollama://", wantOpen: true},
		{name: "root URL opens app", request: "ollama:///", wantOpen: true},
		{name: "apps URL opens Apps", request: "ollama://apps", wantApps: true},
		{name: "apps path opens Apps", request: "ollama:///apps", wantApps: true},
		{name: "apps trailing slash opens Apps", request: "ollama://apps/", wantApps: true},
		{name: "connect URL starts connection", request: "ollama://connect", wantConnect: true},
		{name: "connect path starts connection", request: "ollama:///connect", wantConnect: true},
		{name: "unsupported URL", request: "ollama://unsupported", wantErr: true},
		{name: "invalid URL", request: "ollama://%", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connected := false
			opened := false
			openedApps := false
			err := dispatchURLSchemeRequest(
				tt.request,
				func() { connected = true },
				func() { opened = true },
				func() { openedApps = true },
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf("dispatchURLSchemeRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if connected != tt.wantConnect {
				t.Errorf("connect called = %v, want %v", connected, tt.wantConnect)
			}
			if opened != tt.wantOpen {
				t.Errorf("open called = %v, want %v", opened, tt.wantOpen)
			}
			if openedApps != tt.wantApps {
				t.Errorf("open Apps called = %v, want %v", openedApps, tt.wantApps)
			}
		})
	}
}

func TestRunInitialWindowsUIWithBareURL(t *testing.T) {
	hiddenCalls := 0
	urlCalls := 0
	onboardingCalls := 0
	openCalls := 0

	runInitialWindowsUI(
		false,
		true,
		"ollama://",
		func() { hiddenCalls++ },
		func(request string) {
			urlCalls++
			err := dispatchURLSchemeRequest(request,
				func() { t.Fatal("unexpected sign-in") },
				func() { openCalls++ },
				func() { t.Fatal("unexpected Apps navigation") },
			)
			if err != nil {
				t.Fatalf("dispatchURLSchemeRequest() error = %v", err)
			}
		},
		func(path string) {
			onboardingCalls++
		},
	)

	if urlCalls != 1 {
		t.Fatalf("URL handled %d times, want 1", urlCalls)
	}
	if openCalls != 1 {
		t.Errorf("app opened %d times, want 1", openCalls)
	}
	if hiddenCalls != 0 {
		t.Errorf("hidden startup called %d times, want 0", hiddenCalls)
	}
	if onboardingCalls != 0 {
		t.Errorf("onboarding opened %d times, want 0", onboardingCalls)
	}
}

func TestRunInitialWindowsUIWithAppsURL(t *testing.T) {
	appsCalls := 0
	runInitialWindowsUI(
		false,
		true,
		"ollama://apps",
		func() { t.Fatal("unexpected hidden startup") },
		func(request string) {
			err := dispatchURLSchemeRequest(request,
				func() { t.Fatal("unexpected sign-in") },
				func() { t.Fatal("unexpected home navigation") },
				func() { appsCalls++ },
			)
			if err != nil {
				t.Fatalf("dispatchURLSchemeRequest() error = %v", err)
			}
		},
		func(string) { t.Fatal("unexpected onboarding") },
	)
	if appsCalls != 1 {
		t.Fatalf("Apps opened %d times, want 1", appsCalls)
	}
}

func TestRunInitialWindowsUIRoutesInteractiveLaunch(t *testing.T) {
	for _, tt := range []struct {
		name           string
		showOnboarding bool
		wantPath       string
	}{
		{name: "fresh install preserves onboarding", showOnboarding: true, wantPath: "/"},
		{name: "returning launch opens apps", wantPath: "/connect"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			runInitialWindowsUI(
				false,
				tt.showOnboarding,
				"",
				func() { t.Fatal("unexpected hidden startup") },
				func(string) { t.Fatal("unexpected URL handling") },
				func(path string) { gotPath = path },
			)
			if gotPath != tt.wantPath {
				t.Fatalf("initial UI path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}
