/*
File:		user_settings.go
Author: 	https://github.com/vincebel7
Purpose:	User-wide settings
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	ID             string              `json:"id"`
	Author         string              `json:"author"`
	Achievements   map[int]Achievement `json:"achievements"`
	AchievementsOn bool                `json:"achievements_on"`
	ProgramVer     string              `json:"program_ver"`
}

var user_settings Settings

func loadUserSettings() {
	filename := "saves/user_settings.json"

	// Check if file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// File doesn't exist, create it
		os.Create(filename)

		settings := Settings{
			ID:             idgen(8),
			Author:         "",
			Achievements:   make(map[int]Achievement),
			AchievementsOn: true,
		}
		user_settings = settings
		saveUserSettings()
	}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("[Error] File not found: %s", filename)
	}

	b1 := make([]byte, 1000000) //TODO: secure this
	n1, err := f.Read(b1)

	if err != nil {
		fmt.Printf("[Error] File not found: %s", filename)
	}

	//unmarshal
	var settings_obj Settings
	err = json.Unmarshal(b1[:n1], &settings_obj)
	if err != nil {
		fmt.Printf("err: %v", err)
	}

	user_settings = settings_obj

	buildAchievementCatalog()
}

func saveUserSettings() {
	marshString, err := json.Marshal(user_settings)
	if err != nil {
		log.Println(err)
	}
	//Write to file
	filename := "saves/user_settings.json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	os.Truncate(filename, int64(len(marshString)))
}

func changeSettingsName() {
	fmt.Print("\nPlease enter your name: ")
	scanner.Scan()
	username := scanner.Text()
	user_settings.Author = username
	saveUserSettings()
}

func toggleAchievements() {}

func resetAchievements() {
	user_settings.Achievements = make(map[int]Achievement)
	saveUserSettings()
	fmt.Println("[Notice] Achievements have been reset")
}

func resetProgramSettings() {
	os.Remove("saves/user_settings.json")
	fmt.Println("[Notice] User preferences have been reset\n")
	loadUserSettings()
	intro()
}

func wipeSaves() {
	dir := "saves/user_saves"
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file ends with .json
		if !info.IsDir() && filepath.Ext(info.Name()) == ".json" {
			err := os.Remove(path) // Delete the file
			if err != nil {
				log.Printf("[Error] Failed to remove file: %s, error: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("[Error] Error wiping saves: %v", err)
	} else {
		fmt.Println("[Notice] Save files have been wiped")
	}
}

func resetAllPrompt() {
	fmt.Printf("\nAre you sure you want do delete all settings, Achievements, and saved networks? [y/n]: ")
	scanner.Scan()
	confirmation := scanner.Text()
	confirmation = strings.ToUpper(confirmation)

	fmt.Printf("\n")

	if confirmation == "Y" {
		wipeSaves()
		resetAchievements()
		resetProgramSettings()
	}
}
