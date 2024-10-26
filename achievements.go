/*
File:		achievements.go
Author: 	https://github.com/vincebel7
Purpose:	Achievements

Note: There are two types of achievements: State-based, and action-based.
State-based: Achievements based on network state which can be checked regularly (network has a router)
Action-based: Achievements given when a particular action occurs (successful ping)
*/

package main

import "fmt"

type Achievement struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ProgramVer string `json:"program_ver"`
}

var AchievementCatalog = make(map[int]Achievement)

// achievement IDs
var ROUTINE_BUSINESS = 1
var UNITED_PINGDOM = 2

func buildAchievementCatalog() {
	achievement := Achievement{
		ID:   ROUTINE_BUSINESS,
		Name: "Route-ine Business",
	}
	AchievementCatalog[ROUTINE_BUSINESS] = achievement

	achievement = Achievement{
		ID:   UNITED_PINGDOM,
		Name: "United Pingdom",
	}
	AchievementCatalog[UNITED_PINGDOM] = achievement
}

func displayAchievements() {
	fmt.Printf("ACHIEVEMENTS:\n")
	fmt.Printf("#\tAchievement Name\t\tUnlocked\n")

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

		fmt.Printf("%d\t%s\t%s\n", AchievementCatalog[i].ID, PadRight(AchievementCatalog[i].Name, 25), unlockedChar)
	}
}

func achievementAward(achievement Achievement) {
	fmt.Printf("\n[ACHIEVEMENT COMPLETE] \"%s\" (#%d)\n\n", achievement.Name, achievement.ID)
	user_settings.Achievements[achievement.ID] = achievement
	saveUserSettings()
}

// Does a check of state-based achievements
func achievementCheck() {
	if !user_settings.AchievementsOn {
		return
	}

	achievementTester(ROUTINE_BUSINESS)
}

// Kicks off tests for incomplete achievements
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
		}
	}
}

// achievement 1: Create a router
func achievement1Test() {
	achievementComplete := false

	if snet.Router.ID != "" {
		achievementComplete = true
	}

	if achievementComplete {
		achievement := AchievementCatalog[ROUTINE_BUSINESS]
		achievementAward(achievement)
	}
}

// achievement 2: Successful ping
func achievement2Test() {
	// If this function is called, the achievement is already complete (action-based)
	achievement := AchievementCatalog[UNITED_PINGDOM]
	achievementAward(achievement)
}
