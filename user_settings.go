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
	"strings"
)

type Settings struct {
	ID           string `json:"id"`
	Author       string `json:"author"`
	Challenges   []Host `json:"challenges"`
	ChallengesOn bool   `json:"challenges_on"`
	ProgramVer   string `json:"program_ver"`
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
			ID:     idgen(8),
			Author: "",
		}
		user_settings = settings
		saveUserSettings()
	}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("1File not found: %s", filename)
	}

	b1 := make([]byte, 1000000) //TODO: secure this
	n1, err := f.Read(b1)

	if err != nil {
		fmt.Printf("1File not found: %s", filename)
	}

	//unmarshal
	var settings_obj Settings
	err = json.Unmarshal(b1[:n1], &settings_obj)
	if err != nil {
		fmt.Printf("err: %v", err)
	}

	user_settings = settings_obj
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

func toggleChallenges() {}

func resetChallenges() {}

func resetProgramSettings() {
	os.Remove("saves/user_settings.json")
}

func resetProgramPrompt() {
	fmt.Printf("\nAre you sure you want do delete all settings, challenges, and saved networks? [y/n]: ")
	scanner.Scan()
	confirmation := scanner.Text()
	confirmation = strings.ToUpper(confirmation)
	if confirmation == "Y" {
		resetProgramSettings()
		//TODO: Wipe saves
	}
}
