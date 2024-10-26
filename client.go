/*
File:		client.go
Author: 	https://github.com/vincebel7
Purpose:	User menus and main program loop
*/

package main

import (
	"fmt"
	"os"
	"strings"
)

var currentVersion = "v0.3.0"

func intro() {
	fmt.Println("ltdnet " + currentVersion)

	if user_settings.Author == "" {
		changeSettingsName()
	}
}

func startMenu() bool {
	advanceMenus := false
	selection := false
	fmt.Println("\nPlease select an option:")
	fmt.Println(" 1) Create new network")
	fmt.Println(" 2) Select saved network")
	fmt.Println(" 3) Show Achievements")
	fmt.Println(" 4) Preferences")
	for !selection {
		fmt.Print("\nAction: ")

		scanner.Scan()
		option := scanner.Text()

		switch strings.ToUpper(option) {
		case "1", "C", "NEW", "CREATE":
			selection = true
			advanceMenus = true
			newNetworkPrompt()
		case "2", "S", "SELECT":
			selection = true
			advanceMenus = true
			selectNetwork()
		case "3", "A", "ACHIEVEMENTS":
			selection = true
			displayAchievements()
		case "4", "P", "PREFERENCES", "PREF":
			selection = true
			preferencesMenu()
		default:
			fmt.Println("Not a valid option. Options: 1, 2, 3")
		}
	}

	return advanceMenus
}

func preferencesMenu() {
	selection := false
	fmt.Println("\nUSER PREFERENCES")
	fmt.Println("\nPlease select an option:")
	fmt.Println(" 1) Change name")
	fmt.Println(" 2) Disable/Enable Achievements")
	fmt.Println(" 3) Reset Achievements")
	fmt.Println(" 4) Reset user preferences")
	fmt.Println(" 5) Reset all program data")

	for !selection {
		fmt.Print("\nAction: ")

		scanner.Scan()
		option := scanner.Text()

		switch strings.ToUpper(option) {
		case "1":
			selection = true
			changeSettingsName()
		case "2":
			toggleAchievements()
		case "3":
			selection = true
			resetAchievements()
		case "4":
			selection = true
			resetProgramSettings()
		case "5":
			selection = true
			resetProgramPrompt()
		default:
			fmt.Println("Not a valid option. Options: 1, 2, 3, 4, 5")
		}
	}

}

func actionsMenu() {
	fmt.Print("> ")
	scanner.Scan()
	action_selection := scanner.Text()
	actionword1 := ""
	actionword2 := ""
	actionword3 := ""
	actionword4 := ""
	if action_selection != "" {
		actionword0 := strings.Fields(action_selection)
		if len(actionword0) > 0 {
			actionword1 = actionword0[0]
		}

		if len(actionword0) > 1 {
			actionword2 = actionword0[1]

			if len(actionword0) > 2 {
				actionword3 = actionword0[2]

				if len(actionword0) > 3 {
					actionword4 = actionword0[3]
				}
			}
		}
	}
	switch actionword1 {
	case "":

	case "add":
		if actionword3 != "" {
			switch actionword2 {
			case "router":
				if actionword4 == "" {
					actionword4 = "Bobcat"
				}
				addRouter(actionword3, actionword4)
				save()

			case "switch":
				addSwitch(actionword3)
				save()

			case "host":
				addHost(actionword3)
				save()

			default:
				fmt.Println(" Usage: add <host|switch|router> <hostname>")
			}
		} else {
			fmt.Println(" Usage: add <host|switch|router> <hostname>")
		}

	case "del", "delete":
		switch actionword2 {
		case "router":
			fmt.Printf("\nAre you sure you want do delete router %s? [y/N]: ", actionword3)
			scanner.Scan()
			confirmation := scanner.Text()
			confirmation = strings.ToUpper(confirmation)
			if confirmation == "Y" {
				delRouter() // Only one router per network currently
				save()
			}

		case "switch":
			if actionword3 != "" {
				fmt.Printf("\nAre you sure you want do delete switch %s? [y/N]: ", actionword3)
				scanner.Scan()
				confirmation := scanner.Text()
				confirmation = strings.ToUpper(confirmation)
				if confirmation == "Y" {
					delSwitch()
					save()
				}
			}

		case "host":
			if actionword3 != "" {
				fmt.Printf("\nAre you sure you want do delete host %s? [y/N]: ", actionword3)
				scanner.Scan()
				confirmation := scanner.Text()
				confirmation = strings.ToUpper(confirmation)
				if confirmation == "Y" {
					delHost(actionword3)
					save()
				}
			}

		default:
			fmt.Println(" Usage: del <host|switch|router> <hostname>")
		}

	case "link":
		if (actionword2 == "host") && (actionword3 != "") && (actionword4 != "") {
			linkHost(actionword3, actionword4)
			save()
		} else {
			fmt.Println(" Usage: link host <hostname> <router_hostname>")
		}

	case "unlink":
		if (actionword2 == "host") && (actionword3 != "") {
			unlinkHost(actionword3)
			save()
		} else {
			fmt.Println(" Usage: unlink host <hostname>")
		}

	case "control":
		if actionword2 != "" {
			switch actionword2 {
			case "host":
				controlHost(actionword3)
				save()

			case "switch":
				controlSwitch(actionword3)
				save()

			case "router":
				controlRouter(actionword3)
				save()

			default:
				fmt.Println(" Usage: control <host|switch|router> <hostname>")
			}
		} else {
			fmt.Println(" Usage: control <host|switch|router> <hostname>")
		}

	case "achievements":
		printAchievementsHelp := func() {
			fmt.Println("",
				"achievements show\tShow your Achievements\n",
				"achievements info <#>\tGet information about an Achievement\n",
				"achievements explain\tLearn about ltdnet's Achievements system",
			)
		}
		switch action_selection {
		case "achievements", "achievements help", "achievements ?":
			printAchievementsHelp()
		case "achievements show":
			displayAchievements()
		case "achievements info":
			fmt.Println("Not implemented yet")
		default:
			fmt.Println(" Invalid command. Type 'achievements ?' for a list of commands.")
		}

	case "save":
		save()

	case "reload":
		loadNetwork(snet.Name, "user")

	case "show", "sh":
		switch action_selection {
		case "show network overview", "sh network overview":
			overview()

		case "show diagram", "sh diagram":
			drawDiagram(snet.Router.ID)

		default:
			if len(action_selection) > 12 { // show device
				show(action_selection[12:])
			} else {
				fmt.Println(
					"show network overview\n",
					"show device <hostname>\n",
					"show diagram",
				)
			}
		}

	case "netdump":
		fmt.Println(snet, "")

	case "debug":
		if actionword2 != "" {
			setDebug(actionword2)
			save()
		} else {
			fmt.Printf("Current debug level: %d\n", getDebug())
			fmt.Println("\nAll levels:\n",
				"0 - No debugging\n",
				"1 - Errors\n",
				"2 - General network traffic\n",
				"3 - All network traffic\n",
				"4 - All sorts of garbage (development+learning)")
		}

	case "manual", "man":
		launchManual()

	case "exit", "quit", "q":
		os.Exit(0)

	case "help", "?":
		fmt.Println(
			"NETWORK COMMANDS:\n",
			"show <args>\t\tDisplays information\n",
			"add <args>\t\tAdds device to network\n",
			"del <args>\t\tRemoves device from network\n",
			"link <args>\t\tLinks two devices\n",
			"unlink <args>\t\tUnlinks two devices\n",
			"control <args>\t\tLogs in as device\n",

			"\nSYSTEM COMMANDS:\n",
			"achievements <action>\tView user achievements\n",
			"save\t\t\tManually saves network changes\n",
			"reload\t\t\tReloads the network file. May fix runtime bugs\n",
			"debug <0-4>\t\tSets debug level. Default is 1\n",
			"manual\t\t\tLaunches the user manual. Great for beginners!\n",
			"exit\t\t\tExits the program",
			//"netdump\t\tPrints loaded Network object (developer use)\n", HIDDEN
		)

	default:
		fmt.Println(" Invalid command. Type 'help' for a list of commands.")
	}

	achievementCheck()
}

func main() {
	loadUserSettings()
	intro()

	for {
		if startMenu() {
			break
		}
	}

	go Listener()

	for range snet.Hosts {
		<-listenSync
	}
	fmt.Printf("\n[Notice] Debug level is set to %d\n", getDebug())
	fmt.Printf("[Notice] Please note that switch functionality is limited and in development\n")

	fmt.Println("\nltdnetOS:")

	for {
		actionsMenu()
	}
}
