/*
File:		client.go
Author: 	https://github.com/vincebel7
Purpose:	User menus and main program loop
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var currentVersion = "v0.5.1"

func printVersion() {
	fmt.Println("ltdnet " + currentVersion)
}

func intro() {
	printVersion()

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
			fmt.Println("Not a valid option. Options: 1, 2, 3, 4")
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
			resetAllPrompt()
		default:
			fmt.Println("Not a valid option. Options: 1, 2, 3, 4, 5")
		}
	}

}

func actionsMenu() {
	// Set up readline for actionsMenu
	rl, err := readline.New("> ")
	if err != nil {
		fmt.Printf("Error setting up readline: %v\n", err)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // Exit on Ctrl+D or any read error
			fmt.Println("\nExiting...")
			break
		}

		commandString := strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split the input into action words
		commandSplit := strings.Fields(line)

		switch commandSplit[0] {
		case "":

		case "add":
			if commandSplit[2] != "" {
				switch commandSplit[1] {
				case "router":
					if commandSplit[3] == "" {
						commandSplit[3] = "Bobcat"
					}
					addRouter(commandSplit[2], commandSplit[3])
					save()

				case "switch":
					addSwitch(commandSplit[2])
					save()

				case "host":
					addHost(commandSplit[2])
					save()

				default:
					fmt.Println(" Usage: add <host|switch|router> <hostname>")
				}
			} else {
				fmt.Println(" Usage: add <host|switch|router> <hostname>")
			}

		case "del", "delete":
			switch commandSplit[1] {
			case "router":
				fmt.Printf("\nAre you sure you want do delete router %s? [y/N]: ", commandSplit[2])
				scanner.Scan()
				confirmation := scanner.Text()
				confirmation = strings.ToUpper(confirmation)
				if confirmation == "Y" {
					delRouter() // Only one router per network currently
					save()
				}

			case "switch":
				if commandSplit[2] != "" {
					fmt.Printf("\nAre you sure you want do delete switch %s? [y/N]: ", commandSplit[2])
					scanner.Scan()
					confirmation := scanner.Text()
					confirmation = strings.ToUpper(confirmation)
					if confirmation == "Y" {
						delSwitch(commandSplit[2])
						save()
					}
				}

			case "host":
				if commandSplit[2] != "" {
					fmt.Printf("\nAre you sure you want do delete host %s? [y/N]: ", commandSplit[2])
					scanner.Scan()
					confirmation := scanner.Text()
					confirmation = strings.ToUpper(confirmation)
					if confirmation == "Y" {
						delHost(commandSplit[2])
						save()
					}
				}

			default:
				fmt.Println(" Usage: del <host|switch|router> <hostname>")
			}

		case "link":
			if (commandSplit[1] == "host") && (commandSplit[2] != "") && (commandSplit[3] != "") {
				linkHostTo(commandSplit[2], commandSplit[3])
				save()
			} else if (commandSplit[1] == "switch") && (commandSplit[2] != "") && (commandSplit[3] != "") {
				linkSwitchTo(commandSplit[2], commandSplit[3])
				save()
			} else {
				fmt.Println(" Usage: link <host|switch> <hostname> <router_hostname>")
			}

		case "unlink":
			if (commandSplit[1] == "host") && (commandSplit[2] != "") {
				unlinkHost(commandSplit[2])
				save()
			} else {
				fmt.Println(" Usage: unlink host <hostname>")
			}

		case "control", "c":
			if commandSplit[1] != "" {
				switch commandSplit[1] {
				case "host":
					controlHost(commandSplit[2])
					save()

				case "switch":
					controlSwitch(commandSplit[2])
					save()

				case "router":
					controlRouter(commandSplit[2])
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
			switch commandString {
			case "achievements", "achievements help", "achievements ?":
				printAchievementsHelp()
			case "achievements show":
				displayAchievements()
			case "achievements info":

				fmt.Println("Not implemented yet")
			case "achievements explain":
				printAchievementsExplanation()
			default:
				if commandSplit[1] == "info" {
					if commandSplit[2] != "" {
						//achieveNum, _ := strconv.Atoi(actionword3)
						printAchievementInfo(commandSplit[2])
					} else {
						fmt.Println("usage: achievements info <#>")
					}
				} else {
					fmt.Println(" Invalid command. Type 'achievements ?' for a list of commands.")
				}
			}

		case "save":
			save()

		case "reload":
			loadNetwork(snet.Name, "user")

		case "show", "sh":
			switch commandString {
			case "show overview", "sh overview":
				overview()

			case "show diagram", "sh diagram":
				drawDiagram(snet.Router.ID)

			default:
				if len(commandString) > 12 { // show device
					show(commandString[12:])
				} else {
					fmt.Println("",
						"show overview\n",
						"show device <hostname>\n",
						"show diagram",
					)
				}
			}

		case "netdump":
			fmt.Println(snet, "")

		case "debug":
			if commandSplit[1] != "" {
				setDebug(commandSplit[1])
				save()
			} else {
				fmt.Printf("Current debug level: %d\n", getDebug())
				fmt.Println("\nAll levels (least to most verbose):\n",
					"0 - No debugging\n",
					"1 - Errors\n",
					"2 - Network traffic (receive)\n",
					"3 - Network traffic (send+receive) + Warnings\n",
					"4 - Step-by-step device actions")
			}

		case "manual", "man":
			launchManual()

		case "version", "ver":
			printVersion()

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
				"version\t\tltdnet version info\n",
				"exit\t\t\tExits the program",
				//"netdump\t\tPrints loaded Network object (developer use)\n", HIDDEN
			)

		default:
			fmt.Println(" Invalid command. Type 'help' for a list of commands.")
		}

		achievementCheck()
	}
}

func launchManual() {
	file, err := os.Open("USER-MANUAL")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
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

	for h := range snet.Hosts {
		for range snet.Hosts[h].Interfaces {
			<-listenSync
		}
	}
	fmt.Printf("\n[Notice] Debug level is set to %d\n", getDebug())
	fmt.Printf("[Notice] Please note that switches can't yet link to routers or other switches.\n")

	fmt.Println("\nltdnetOS:")

	actionsMenu()
}
