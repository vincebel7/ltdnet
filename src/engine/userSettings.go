package engine

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/vincebel7/ltdnet/src/model"
)

func UserSettings() *model.Settings { return Instance().Settings }

func (e *Engine) SaveUserSettings() error {
	// Check if file / directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	savesDir := filepath.Join(homeDir, "ltdnet_saves")
	if err := os.MkdirAll(savesDir, 0755); err != nil {
		return err
	}
	settingsFile := filepath.Join(savesDir, "user_settings.json")

	marshString, err := json.MarshalIndent(UserSettings(), "", " ")
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(settingsFile, marshString, 0660); err != nil {
		return err
	}
	return nil
}
