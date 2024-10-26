/*
File:		achievements.go
Author: 	https://github.com/vincebel7
Purpose:	Achievements

Note: There are two types of Achievements: State-based, and action-based.
State-based: Achievements based on network state which can be checked regularly (network has a router)
Action-based: Achievements given when a particular action occurs (successful ping)
*/

package main

import "fmt"

type Achievement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ProgramVer  string `json:"program_ver"`
}

var AchievementCatalog = make(map[int]Achievement)

// Achievement IDs
var ROUTINE_BUSINESS = 1
var UNITED_PINGDOM = 2
var ARP_HOT = 3

func buildAchievementCatalog() {
	achievement := Achievement{
		ID:          ROUTINE_BUSINESS,
		Name:        "Route-ine Business",
		Description: "Add a router to your network.",
	}
	AchievementCatalog[ROUTINE_BUSINESS] = achievement

	achievement = Achievement{
		ID:          UNITED_PINGDOM,
		Name:        "United Pingdom",
		Description: "Successfully ping from one device to another.",
	}
	AchievementCatalog[UNITED_PINGDOM] = achievement

	achievement = Achievement{
		ID:          ARP_HOT,
		Name:        "ARP It Like It's Hot",
		Description: "Manually send an ARP request, and receive a reply.",
	}
	AchievementCatalog[ARP_HOT] = achievement
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
			unlockedChar = "Y"
		}

		fmt.Printf("%d\t%s\t%s\t%s\n", AchievementCatalog[i].ID, PadRight(AchievementCatalog[i].Name, 25), PadRight(AchievementCatalog[i].Description, 50), unlockedChar)
	}
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

// Achievement 2: Successful manual ARP request
func achievement3Test() {
	// If this function is called, the achievement is already complete (action-based)
	achievement := AchievementCatalog[ARP_HOT]
	achievementAward(achievement)
}
