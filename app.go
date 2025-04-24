package main

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"os/exec"
	"log"
	"regexp"
	"strings"
)

// ParadigmVersion represents a found Paradigm software version
type ParadigmVersion struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Path         string `json:"path"`
	ExecutablePath string `json:"executablePath"`
}

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// AddCustomParadigm allows the user to manually add a Paradigm executable
func (a *App) AddCustomParadigm(executablePath string) ParadigmVersion {
	log.Printf("Adding custom Paradigm at: %s", executablePath)
	
	// Validate path exists
	if _, err := os.Stat(executablePath); os.IsNotExist(err) {
		log.Printf("Custom executable does not exist: %s", executablePath)
		return ParadigmVersion{}
	}
	
	// Get directory containing executable
	dirPath := filepath.Dir(executablePath)
	
	// Try to extract version from directory path
	paradigmRegex := regexp.MustCompile(`Paradigm[_\s]*v?(\d+\.\d+\.\d+)`)
	matches := paradigmRegex.FindStringSubmatch(dirPath)
	
	version := "Custom"
	if len(matches) > 1 {
		version = matches[1]
	}
	
	// Get the parent folder name as a fallback
	folderName := filepath.Base(dirPath)
	if folderName == "." || folderName == "/" {
		folderName = "Custom Paradigm"
	}
	
	// Use filename if we're still uncertain
	name := "Paradigm"
	if version == "Custom" {
		exeName := filepath.Base(executablePath)
		name = strings.TrimSuffix(exeName, filepath.Ext(exeName))
		// Make it look nicer
		name = strings.ReplaceAll(name, "_", " ")
		name = strings.Title(strings.ToLower(name))
	}
	
	return ParadigmVersion{
		Name:          name,
		Version:       version,
		Path:          dirPath,
		ExecutablePath: executablePath,
	}
}

// ScanForParadigmVersions scans for Paradigm installations
func (a *App) ScanForParadigmVersions(baseDirPath string) []ParadigmVersion {
	// Initialize with empty slice, not nil
	versions := []ParadigmVersion{}
	
	log.Printf("Scanning for Paradigm versions in: %s", baseDirPath)
	
	// Ensure directory exists
	if _, err := os.Stat(baseDirPath); os.IsNotExist(err) {
		log.Printf("Base directory does not exist: %s", baseDirPath)
		return versions
	}
	
	// Regular expression to match Paradigm X.X.X format
	// More flexible pattern that catches variations like:
	// - "Paradigm 3.4.2"
	// - "Paradigm3.4.2"
	// - "Paradigm_3.4.2"
	// - "Paradigm v3.4.2"
	paradigmRegex := regexp.MustCompile(`Paradigm[_\s]*v?(\d+\.\d+\.\d+)`)
	
	// Scan the main directory first
	findParadigmVersions(baseDirPath, paradigmRegex, &versions)
	
	// Also scan common subdirectories where Paradigm might be installed
	// e.g., ETC/Software/Paradigm X.X.X
	commonSubdirs := []string{"Software", "Programs", "Applications"}
	for _, subdir := range commonSubdirs {
		subdirPath := filepath.Join(baseDirPath, subdir)
		if _, err := os.Stat(subdirPath); err == nil {
			findParadigmVersions(subdirPath, paradigmRegex, &versions)
		}
	}
	
	// If no versions found, add at least a placeholder
	if len(versions) == 0 {
		// Add common folders for the user to browse
		*&versions = append(*&versions, ParadigmVersion{
			Name:         "No versions found",
			Version:      "N/A",
			Path:         baseDirPath,
			ExecutablePath: "",
		})
	}
	
	// Sort versions by version number (newer versions first)
	if len(versions) > 0 {
		sort.Slice(versions, func(i, j int) bool {
			// This is a simple string comparison - may not work correctly for all version formats
			// For more complex version comparison, a dedicated version comparison library should be used
			return versions[i].Version > versions[j].Version
		})
	}
	
	log.Printf("Returning %d Paradigm versions", len(versions))
	return versions
}

// findParadigmVersions is a helper function that scans a directory for Paradigm versions
func findParadigmVersions(dirPath string, paradigmRegex *regexp.Regexp, versions *[]ParadigmVersion) {
	// Get all subdirectories in the target directory
	dirs, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return
	}
	
	log.Printf("Found %d items in directory %s", len(dirs), dirPath)
	
	for _, dir := range dirs {
		if dir.IsDir() {
			dirName := dir.Name()
			log.Printf("Checking directory: %s", dirName)
			
			matches := paradigmRegex.FindStringSubmatch(dirName)
			
			if len(matches) > 1 {
				versionNumber := matches[1]
				log.Printf("Match found! Version: %s", versionNumber)
				versionPath := filepath.Join(dirPath, dirName)
				
				// Create a list of potential executable paths to try
				execPaths := []string{}
				
				// Common executable names
				execNames := []string{
					"light_designer.exe",
					"lightdesigner.exe",
					"LightDesigner.exe", 
					"ld.exe",
					"paradigm.exe",
				}
				
				// Common directory structures
				dirStructures := []string{
					filepath.Join(versionPath, "LightDesigner"),
					filepath.Join(versionPath, "Light Designer"), // try with space too
					filepath.Join(versionPath, "bin"),
					filepath.Join(versionPath, "app"),
					versionPath, // root folder
				}
				
				// Generate all possible combinations
				for _, dir := range dirStructures {
					for _, name := range execNames {
						execPaths = append(execPaths, filepath.Join(dir, name))
					}
				}
				
				// Try all paths until we find an executable
				foundExe := false
				exePath := ""
				
				for _, path := range execPaths {
					log.Printf("Checking for executable at: %s", path)
					if _, err := os.Stat(path); err == nil {
						exePath = path
						foundExe = true
						log.Printf("Found executable at: %s", exePath)
						break
					}
				}
				
				// If we found an executable, add it to our versions list
				if foundExe {
					log.Printf("Found Paradigm version %s with executable: %s", versionNumber, exePath)
					*versions = append(*versions, ParadigmVersion{
						Name:          "Paradigm",
						Version:       versionNumber,
						Path:          versionPath,
						ExecutablePath: exePath,
					})
				} else {
					log.Printf("Found Paradigm version %s but no executable found", versionNumber)
				}
			}
		}
	}
}

// GetDefaultParadigmPath returns the default path where Paradigm is typically installed
func (a *App) GetDefaultParadigmPath() string {
	// Typically in Program Files (x86)\ETC
	path := filepath.Join(os.Getenv("ProgramFiles(x86)"), "ETC")
	log.Printf("Default Paradigm path: %s", path)
	return path
}

// LaunchParadigm launches the specified Paradigm version
func (a *App) LaunchParadigm(executablePath string) bool {
	log.Printf("Launching Paradigm: %s", executablePath)
	cmd := exec.Command(executablePath)
	err := cmd.Start()
	if err != nil {
		log.Printf("Error launching Paradigm: %v", err)
		return false
	}
	return true
}

// BrowseExecutable allows user to get files in a directory to help manually select
func (a *App) BrowseDirectory(dirPath string) []string {
	log.Printf("Browsing directory: %s", dirPath)
	
	fileList := []string{}
	
	// Ensure directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("Directory does not exist: %s", dirPath)
		return fileList
	}
	
	// Get all files in the directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return fileList
	}
	
	// Add each file to the list
	for _, file := range files {
		itemPath := filepath.Join(dirPath, file.Name())
		if file.IsDir() {
			itemPath += "/"
		}
		fileList = append(fileList, itemPath)
	}
	
	return fileList
}
