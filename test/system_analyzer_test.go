// File: test/system_analyzer_test.go
package test

import (
	"testing"

	"github.com/minand-mohan/execute-my-will/internal/system"
)

func TestAnalyzer_AnalyzeSystem(t *testing.T) {
	analyzer := system.NewAnalyzer()

	info, err := analyzer.AnalyzeSystem()

	if err != nil {
		t.Errorf("AnalyzeSystem() should not error, got: %v", err)
	}

	if info == nil {
		t.Fatal("AnalyzeSystem() should return system info")
	}

	// Test that basic fields are populated
	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	if info.Shell == "" {
		t.Error("Shell should not be empty")
	}

	if info.CurrentDir == "" {
		t.Error("CurrentDir should not be empty")
	}

	if info.HomeDir == "" {
		t.Error("HomeDir should not be empty")
	}

	// Test that slices are initialized (not nil)
	if info.PackageManagers == nil {
		t.Error("PackageManagers should be initialized")
	}

	if info.InstalledPackages == nil {
		t.Error("InstalledPackages should be initialized")
	}

	if info.AvailableCommands == nil {
		t.Error("AvailableCommands should be initialized")
	}

	if info.PathDirectories == nil {
		t.Error("PathDirectories should be initialized")
	}

	// Test that at least some package managers are detected
	if len(info.PackageManagers) == 0 {
		t.Error("At least one package manager should be detected (even 'unknown')")
	}

	// Test that PATH directories are populated
	if len(info.PathDirectories) == 0 {
		t.Error("PATH directories should be populated")
	}
}

func TestAnalyzer_SystemInfoContent(t *testing.T) {
	analyzer := system.NewAnalyzer()

	info1, err1 := analyzer.AnalyzeSystem()
	if err1 != nil {
		t.Fatalf("First analysis failed: %v", err1)
	}

	info2, err2 := analyzer.AnalyzeSystem()
	if err2 != nil {
		t.Fatalf("Second analysis failed: %v", err2)
	}

	// Test consistency - multiple calls should return same basic info
	if info1.OS != info2.OS {
		t.Errorf("OS should be consistent across calls: %s vs %s", info1.OS, info2.OS)
	}

	if info1.Shell != info2.Shell {
		t.Errorf("Shell should be consistent across calls: %s vs %s", info1.Shell, info2.Shell)
	}

	if info1.HomeDir != info2.HomeDir {
		t.Errorf("HomeDir should be consistent across calls: %s vs %s", info1.HomeDir, info2.HomeDir)
	}
}

func TestAnalyzer_Interface(t *testing.T) {
	// Test that NewAnalyzer returns the SystemAnalyzer interface
	var analyzer system.SystemAnalyzer = system.NewAnalyzer()

	info, err := analyzer.AnalyzeSystem()
	if err != nil {
		t.Errorf("Interface method should work: %v", err)
	}

	if info == nil {
		t.Error("Interface method should return info")
	}
}

func TestSystemInfo_Structure(t *testing.T) {
	// Test the Info struct directly
	info := &system.Info{
		OS:                "test-os",
		Shell:             "test-shell",
		PackageManagers:   []string{"test-pm"},
		CurrentDir:        "/test/current",
		HomeDir:           "/test/home",
		PathDirectories:   []string{"/test/bin"},
		InstalledPackages: []string{"test-package"},
		AvailableCommands: []string{"test-command"},
	}

	// Verify all fields are accessible and correct
	if info.OS != "test-os" {
		t.Errorf("Expected OS 'test-os', got '%s'", info.OS)
	}

	if info.Shell != "test-shell" {
		t.Errorf("Expected Shell 'test-shell', got '%s'", info.Shell)
	}

	if len(info.PackageManagers) != 1 || info.PackageManagers[0] != "test-pm" {
		t.Errorf("Expected PackageManagers ['test-pm'], got %v", info.PackageManagers)
	}

	if info.CurrentDir != "/test/current" {
		t.Errorf("Expected CurrentDir '/test/current', got '%s'", info.CurrentDir)
	}

	if info.HomeDir != "/test/home" {
		t.Errorf("Expected HomeDir '/test/home', got '%s'", info.HomeDir)
	}

	if len(info.PathDirectories) != 1 || info.PathDirectories[0] != "/test/bin" {
		t.Errorf("Expected PathDirectories ['/test/bin'], got %v", info.PathDirectories)
	}

	if len(info.InstalledPackages) != 1 || info.InstalledPackages[0] != "test-package" {
		t.Errorf("Expected InstalledPackages ['test-package'], got %v", info.InstalledPackages)
	}

	if len(info.AvailableCommands) != 1 || info.AvailableCommands[0] != "test-command" {
		t.Errorf("Expected AvailableCommands ['test-command'], got %v", info.AvailableCommands)
	}
}
