/*
File:		user_settings.go
Author: 	https://github.com/vincebel7
Purpose:	User-wide settings
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/logging"
	"github.com/vincebel7/ltdnet/src/model"
	"github.com/vincebel7/ltdnet/src/version"
)

func UserSettings() *model.Settings { return engine.Instance().Settings }

func loadUserSettings() {
	// Check if file / directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		systemLog(1, "loadUserSettings", fmt.Sprintf("Error finding home directory: %v", err))
		return
	}
	savesDir := filepath.Join(homeDir, "ltdnet_saves")
	userSavesDir := filepath.Join(savesDir, "user")
	settingsFile := filepath.Join(savesDir, "user_settings.json")

	// Check if the saves directory exists, and create it if not
	if err := os.MkdirAll(userSavesDir, 0755); err != nil {
		systemLog(1, "loadUserSettings", fmt.Sprintf("Error creating directory: %v", err))
		return
	}

	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		// Create default settings
		def := model.Settings{
			ID:             idgen(8),
			Author:         "",
			Achievements:   make(map[int]model.Achievement),
			AchievementsOn: true,
			ProgramVer:     version.ProgramVersion,
		}
		engine.Instance().Settings = &def
		saveUserSettings()
		buildAchievementCatalog()
		return
	}

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		systemLog(1, "loadUserSettings", fmt.Sprintf("Could not read settings file: %v", err))
		return
	}

	// Unmarshal
	var loaded model.Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
		systemLog(1, "loadUserSettings", fmt.Sprintf("Could not parse settings file: %v", err))
		// Fallback to defaults
		loaded = model.Settings{
			ID:             idgen(8),
			Achievements:   make(map[int]model.Achievement),
			AchievementsOn: true,
			ProgramVer:     version.ProgramVersion,
		}
	}
	// Ensure maps not nil
	if loaded.Achievements == nil {
		loaded.Achievements = make(map[int]model.Achievement)
	}
	engine.Instance().Settings = &loaded
	buildAchievementCatalog()

	// Apply log level
	val := strconv.Itoa(UserSettings().LogLevel)
	intval, _ := strconv.Atoi(val)
	engine.Instance().Logger.SetLevel(logging.Level(intval))
}

func saveUserSettings() {
	// Check if file / directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		systemLog(1, "saveUserSettings", fmt.Sprintf("Error finding home directory: %v", err))
		return
	}
	savesDir := filepath.Join(homeDir, "ltdnet_saves")
	if err := os.MkdirAll(savesDir, 0755); err != nil {
		systemLog(1, "saveUserSettings", fmt.Sprintf("Error creating saves directory: %v", err))
		return
	}
	settingsFile := filepath.Join(savesDir, "user_settings.json")

	marshString, err := json.MarshalIndent(UserSettings(), "", " ")
	if err != nil {
		systemLog(1, "saveUserSettings", fmt.Sprintf("Error marshaling user settings: %v", err))
		return
	}

	// Write to file
	if err := os.WriteFile(settingsFile, marshString, 0660); err != nil {
		systemLog(1, "saveUserSettings", fmt.Sprintf("Error writing user settings to file: %v", err))
	}
}

func changeSettingsName() {
	// Ensure save directories exist
	homeDir, err := os.UserHomeDir()
	if err == nil {
		savesDir := filepath.Join(homeDir, "ltdnet_saves")
		userSavesDir := filepath.Join(savesDir, "user")
		testSavesDir := filepath.Join(savesDir, "test")
		os.MkdirAll(userSavesDir, 0755)
		os.MkdirAll(testSavesDir, 0755)
	}

	fmt.Print("\nPlease enter your name: ")
	inScanner := engine.Instance().Scanner
	inScanner.Scan()
	username := inScanner.Text()
	UserSettings().Author = username
	saveUserSettings()
}

func toggleAchievements() {}

func resetAchievements() {
	UserSettings().Achievements = make(map[int]model.Achievement)
	saveUserSettings()
	systemLog(2, "resetAchievements", "User achievements have been reset")
}

func resetProgramSettings() {
	// Check if file / directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		systemLog(1, "resetProgramSettings", fmt.Sprintf("Error finding home directory: %v", err))
		return
	}

	settingsFile := filepath.Join(homeDir, "ltdnet_saves", "user_settings.json")

	_ = os.Remove(settingsFile)
	systemLog(2, "resetProgramSettings", "User preferences have been reset")
	loadUserSettings()
	intro()
}

func wipeSaves() {
	// Check if file / directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		systemLog(1, "wipeSaves", fmt.Sprintf("Error finding home directory: %v", err))
		return
	}

	savesDir := filepath.Join(homeDir, "ltdnet_saves")

	err = filepath.Walk(savesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file ends with .json
		if !info.IsDir() && filepath.Ext(info.Name()) == ".json" {
			err := os.Remove(path) // Delete the file
			if err != nil {
				systemLog(1, "wipeSaves", fmt.Sprintf("Failed to remove file: %s, error: %v", path, err))
			}
		}
		return nil
	})

	if err != nil {
		systemLog(1, "wipeSaves", fmt.Sprintf("Error wiping saves: %v", err))
	} else {
		systemLog(2, "wipeSaves", "All save files have been wiped")
	}
}

func resetAllPrompt() {
	fmt.Printf("\nAre you sure you want do delete all settings, Achievements, and saved networks? [y/n]: ")
	inScanner := engine.Instance().Scanner
	inScanner.Scan()
	confirmation := inScanner.Text()
	confirmation = strings.ToUpper(confirmation)

	fmt.Printf("\n")

	if confirmation == "Y" {
		wipeSaves()
		resetAchievements()
		resetProgramSettings()
	}
}
