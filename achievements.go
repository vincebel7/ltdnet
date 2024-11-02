/*
File:		achievements.go
Author: 	https://github.com/vincebel7
Purpose:	Achievements

Note: There are two types of Achievements: State-based, and action-based.
State-based: Achievements based on network state which can be checked regularly (network has a router)
Action-based: Achievements given when a particular action occurs (successful ping)
*/

package main

import (
	"fmt"
	"strconv"
)

type Achievement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Hint        string `json:"hint"`
	ProgramVer  string `json:"program_ver"`
}

var AchievementCatalog = make(map[int]Achievement)

// Achievement IDs
var ROUTINE_BUSINESS = 1
var UNITED_PINGDOM = 2
var ARP_HOT = 3
var SNIFF_FRAMES = 4

func buildAchievementCatalog() {
	achievement := Achievement{
		ID:          ROUTINE_BUSINESS,
		Name:        "Route-ine Business",
		Description: "Add a router to your network.",
		Hint:        "Have you tried the 'add router' command?",
	}
	AchievementCatalog[ROUTINE_BUSINESS] = achievement

	achievement = Achievement{
		ID:          UNITED_PINGDOM,
		Name:        "United Pingdom",
		Description: "Successfully ping from one device to another.",
		Hint:        "Control a host and ping your default gateway, see if you get a response.",
	}
	AchievementCatalog[UNITED_PINGDOM] = achievement

	achievement = Achievement{
		ID:          ARP_HOT,
		Name:        "ARP It Like It's Hot",
		Description: "Manually send an ARP request, and receive a reply.",
		Hint:        "ARP is how hosts find out other MAC addresses on their network. Try 'arp ?' from a host.",
	}
	AchievementCatalog[ARP_HOT] = achievement

	achievement = Achievement{
		ID:          SNIFF_FRAMES,
		Name:        "Sniffing Your Own Frames",
		Description: "Talk to yourself on localhost",
		Hint:        "Every host has a loopback interface with an address of 127.0.0.1. Try pinging it.",
	}
	AchievementCatalog[SNIFF_FRAMES] = achievement
}

func displayAchievements() {
	fmt.Printf("Achievements:\n")
	fmt.Printf("#\tName\t\t\t\tDescription\t\t\t\t\t\tUnlocked\n")

	keys := make([]int, 1, len(AchievementCatalog))
	for k := range AchievementCatalog {
		keys = append(keys, k)
	}

	for i := range keys {
		if i == 0 {
			continue
		}

		unlockedChar := "."

		if _, exists := user_settings.Achievements[i]; exists {
			unlockedChar = "Yes"
		}

		fmt.Printf("%d\t%s\t%s\t%s\n", AchievementCatalog[i].ID, PadRight(AchievementCatalog[i].Name, 25), PadRight(AchievementCatalog[i].Description, 50), unlockedChar)
	}
}

func printAchievementsExplanation() {
	fmt.Println("Achievements are a fun way to learn how to use ltdnet, and learn networking concepts. Use the achievement commands to guide what you try next. 'achievements show' will list all the achievements and whether you've earned it or not. 'achievements info <#>' will give you a hint about how to unlock an achievement.")
}

func printAchievementInfo(achieveStr string) {
	achievement := Achievement{}
	achievementFound := false
	achieveNum, _ := strconv.Atoi(achieveStr)
	for a := range AchievementCatalog {
		if AchievementCatalog[a].ID == achieveNum {
			achievement = AchievementCatalog[a]
			achievementFound = true
		}
	}

	if !achievementFound {
		fmt.Printf("No achievement \"%s\" found. Usage: achievements info <#>\n", achieveStr)
		return
	}

	fmt.Printf("Achievement #%d: %s\n", achievement.ID, achievement.Name)
	fmt.Printf("Description: %s\n", achievement.Description)
	fmt.Printf("Hint: %s\n", achievement.Hint)
}

func achievementAward(achievement Achievement) {
	fmt.Printf("\n[ACHIEVEMENT COMPLETE] \"%s\" (#%d)\n\n", achievement.Name, achievement.ID)
	user_settings.Achievements[achievement.ID] = achievement
	saveUserSettings()
}

// Does a check of state-based Achievements
func achievementCheck() {
	if !user_settings.AchievementsOn {
		return
	}

	// Test state-based Achievements
	//achievementTester()
}

// Kicks off tests for incomplete Achievements
func achievementTester(achievementID int) {
	if !user_settings.AchievementsOn {
		return
	}

	if _, exists := user_settings.Achievements[achievementID]; !exists {
		switch achievementID {
		case 1:
			achievement1Test()
		case 2:
			achievement2Test()
		case 3:
			achievement3Test()
		case 4:
			achievement4Test()
		}
	}
}

// Achievement 1: Create a router
func achievement1Test() {
	// If this function is called, the achievement is already complete (action-based)
	achievement := AchievementCatalog[ROUTINE_BUSINESS]
	achievementAward(achievement)
}

// Achievement 2: Successful ping
func achievement2Test() {
	// If this function is called, the achievement is already complete (action-based)
	achievement := AchievementCatalog[UNITED_PINGDOM]
	achievementAward(achievement)
}

// Achievement 3: Successful manual ARP request
func achievement3Test() {
	// If this function is called, the achievement is already complete (action-based)
	achievement := AchievementCatalog[ARP_HOT]
	achievementAward(achievement)
}

// Achievement 4: Successful manual ARP request
func achievement4Test() {
	// If this function is called, the achievement is already complete (action-based)
	achievement := AchievementCatalog[SNIFF_FRAMES]
	achievementAward(achievement)
}
