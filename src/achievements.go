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

// Achievement IDs
const (
	ROUTINE_BUSINESS = 1
	UNITED_PINGDOM   = 2
	ARP_HOT          = 3
	SNIFF_FRAMES     = 4
	TEN_HOSTS        = 5
	MY_NAME          = 6
)

func buildAchievementCatalog() {
	eng := engine.Instance()
	add := func(a model.Achievement) {
		eng.AchievementsMap[a.ID] = a
	}
	add(model.Achievement{
		ID:          ROUTINE_BUSINESS,
		Name:        "Route-ine Business",
		Description: "Add a router to your network",
		Hint:        "Have you tried the 'add router' command?",
	})

	add(model.Achievement{
		ID:          UNITED_PINGDOM,
		Name:        "United Pingdom",
		Description: "Successfully ping from one device to another",
		Hint:        "Control a host and ping your default gateway, see if you get a response.",
	})
	add(model.Achievement{
		ID:          ARP_HOT,
		Name:        "ARP It Like It's Hot",
		Description: "Manually send an ARP request, and receive a reply",
		Hint:        "ARP is how hosts find out other MAC addresses on their network. Try 'arp ?' from a host.",
	})
	add(model.Achievement{
		ID:          SNIFF_FRAMES,
		Name:        "Sniffing Your Own Frames",
		Description: "Talk to yourself on localhost",
		Hint:        "Every host has a loopback interface with an address of 127.0.0.1. Try pinging it.",
	})
	add(model.Achievement{
		ID:          TEN_HOSTS,
		Name:        "Room For Ten",
		Description: "Have ten hosts on your network",
		Hint:        "The hosts don't need to be linked...",
	})
	add(model.Achievement{
		ID:          MY_NAME,
		Name:        "My Name Is",
		Description: "Get a host record from a DNS server",
		Hint:        "nslookup is a utility to retrieve records from a DNS server.",
	})
}

func displayAchievements() {
	eng := engine.Instance()
	fmt.Printf("Achievements:\n")
	fmt.Printf("#\tName\t\t\t\tDescription\t\t\t\t\t\tUnlocked\n")

	keys := make([]int, 1, len(eng.AchievementsMap))
	for k := range eng.AchievementsMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, id := range keys {
		a := eng.AchievementsMap[id]
		unlockedChar := "."

		if _, exists := UserSettings().Achievements[id]; exists {
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

func achievementAward(a model.Achievement) {
	if _, exists := UserSettings().Achievements[a.ID]; exists {
		return
	}
	fmt.Printf("\n[ACHIEVEMENT COMPLETE] \"%s\" (#%d)\n\n", a.Name, a.ID)
	UserSettings().Achievements[a.ID] = a
	saveUserSettings()
}

// Does a check of state-based Achievements
func achievementStateCheck() {
	if !UserSettings().AchievementsOn {
		return
	}

	// Test state-based Achievements
	achievementTester(5)
}

// Kicks off tests for incomplete Achievements
func achievementTester(id int) {
	if !UserSettings().AchievementsOn {
		return
	}
	if _, unlocked := UserSettings().Achievements[id]; unlocked {
		return
	}
	switch id {
	case ROUTINE_BUSINESS, UNITED_PINGDOM, ARP_HOT, SNIFF_FRAMES, MY_NAME:
		achievementAward(engine.Instance().AchievementsMap[id])
	case TEN_HOSTS:
		achievement5Test()
	}
}

// Achievement 5: Have ten hosts on your network (state-based)
func achievement5Test() {
	if len(Net().Hosts) >= 10 {
		achievement := engine.Instance().AchievementsMap[TEN_HOSTS]
		achievementAward(achievement)
	}
}
