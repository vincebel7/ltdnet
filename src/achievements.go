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
	"sort"
	"strconv"

	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/model"
)

func displayAchievements() {
	eng := engine.Instance()
	fmt.Printf("Achievements:\n")
	fmt.Printf("#\tName\t\t\t\tDescription\t\t\t\t\t\tUnlocked\n")

	keys := make([]int, 0, len(eng.AchievementsMap))
	for k := range eng.AchievementsMap {
		if k <= 0 { // Skip nonexistent achievement ID 0
			continue
		}
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, id := range keys {
		a := eng.AchievementsMap[id]
		unlockedChar := "."

		if _, exists := engine.UserSettings().Achievements[id]; exists {
			unlockedChar = "Yes"
		}

		fmt.Printf("%d\t%s\t%s\t%s\n",
			a.ID,
			PadRight(a.Name, 25),
			PadRight(a.Description, 50),
			unlockedChar,
		)
	}
}

func printAchievementsExplanation() {
	fmt.Println("Achievements are a fun way to learn how to use ltdnet, and learn networking concepts. Use the achievement commands to guide what you try next. 'achievements show' will list all the achievements and whether you've earned it or not. 'achievements info <#>' will give you a hint about how to unlock an achievement.")
}

func printAchievementInfo(achieveStr string) {
	id, _ := strconv.Atoi(achieveStr)
	a, ok := engine.Instance().AchievementsMap[id]
	if !ok {
		fmt.Printf("No achievement \"%s\" found. Usage: achievements info <#>\n", achieveStr)
		return
	}
	fmt.Printf("Achievement #%d: %s\nDescription: %s\nHint: %s\n",
		a.ID, a.Name, a.Description, a.Hint)
}

// Does a check of state-based Achievements
func achievementStateCheck() {
	achievementCheck(model.AchTenHosts)
}

func achievementCheck(achieveID int) {
	if achieveID <= 0 { // skip invalid ID 0
		return
	}
	a := engine.Instance().AchievementsMap[achieveID]
	if engine.AchievementTester(achieveID) {
		fmt.Printf("\n[ACHIEVEMENT COMPLETE] \"%s\" (#%d)\n\n", a.Name, a.ID)
	}
}
