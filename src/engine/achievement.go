// Achievement check - done by client
// Achievement test - done by engine
// Achievement award - done by engine

package engine

import (
	"github.com/vincebel7/ltdnet/src/model"
)

func BuildAchievementCatalog() {
	eng := Instance()
	add := func(a model.Achievement) {
		eng.AchievementsMap[a.ID] = a
	}
	add(model.Achievement{
		ID:          model.AchRoutineBusiness,
		Name:        "Route-ine Business",
		Description: "Add a router to your network",
		Hint:        "Have you tried the 'add router' command?",
	})
	add(model.Achievement{
		ID:          model.AchUnitedPingdom,
		Name:        "United Pingdom",
		Description: "Successfully ping from one device to another",
		Hint:        "Control a host and ping your default gateway, see if you get a response.",
	})
	add(model.Achievement{
		ID:          model.AchArpHot,
		Name:        "ARP It Like It's Hot",
		Description: "Manually send an ARP request, and receive a reply",
		Hint:        "ARP is how hosts find out other MAC addresses on their network. Try 'arp ?' from a host.",
	})
	add(model.Achievement{
		ID:          model.AchSniffFrames,
		Name:        "Sniffing Your Own Frames",
		Description: "Talk to yourself on localhost",
		Hint:        "Every host has a loopback interface with an address of 127.0.0.1. Try pinging it.",
	})
	add(model.Achievement{
		ID:          model.AchTenHosts,
		Name:        "Room For Ten",
		Description: "Have ten hosts on your network",
		Hint:        "The hosts don't need to be linked...",
	})
	add(model.Achievement{
		ID:          model.AchMyName,
		Name:        "My Name Is",
		Description: "Get a host record from a DNS server",
		Hint:        "nslookup is a utility to retrieve records from a DNS server.",
	})
}

// Kicks off tests for incomplete Achievements
func AchievementTester(id int) bool {
	if !UserSettings().AchievementsOn {
		return false
	}
	if _, unlocked := UserSettings().Achievements[id]; unlocked {
		return false
	}
	switch id {
	case model.AchRoutineBusiness, model.AchUnitedPingdom, model.AchArpHot, model.AchSniffFrames, model.AchMyName:
		return achievementAward(Instance().AchievementsMap[id])
	case model.AchTenHosts:
		return achievement5Test()
	}
	return false
}

// Achievement 5: Have ten hosts on your network (state-based)
func achievement5Test() bool {
	if len(Net().Hosts) >= 10 {
		achievement := Instance().AchievementsMap[model.AchTenHosts]
		return achievementAward(achievement)
	}
	return false
}

// Award the achievement to the user
func achievementAward(a model.Achievement) bool {
	UserSettings().Achievements[a.ID] = a
	Instance().SaveUserSettings()
	return true
}
