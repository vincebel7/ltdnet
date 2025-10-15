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

func printVersion() {
	fmt.Println("ltdnet " + Net().ProgramVer)
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

		inScanner := EngineInstance().Scanner
		inScanner.Scan()
		option := inScanner.Text()

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

		inScanner := EngineInstance().Scanner
		inScanner.Scan()
		option := inScanner.Text()

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
	inScanner := EngineInstance().Scanner

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

		// Split the input into action words, and safely parse arguments
		commandSplit := strings.Fields(line)
		cmd := safeArg(commandSplit, 0)
		arg1 := safeArg(commandSplit, 1)
		arg2 := safeArg(commandSplit, 2)
		arg3 := safeArg(commandSplit, 3)

		switch cmd {
		case "":

		case "add":
			if arg2 != "" {
				switch arg1 {
				case "router":
					model := "Bobcat"
					if arg3 != "" {
						model = arg3
					}
					addRouter(arg2, model)
					save()
				case "switch":
					addSwitch(arg2)
					save()
				case "host":
					addHost(arg2)
					save()
				default:
					fmt.Println(" Usage: add <host|switch|router> <hostname>")
				}
			} else {
				fmt.Println(" Usage: add <host|switch|router> <hostname>")
			}

		case "del", "delete":
			switch arg1 {
			case "router":
				fmt.Printf("\nAre you sure you want do delete router %s? [y/N]: ", arg2)
				inScanner.Scan()
				confirmation := inScanner.Text()
				confirmation = strings.ToUpper(confirmation)
				if confirmation == "Y" {
					delRouter() // Only one router per network currently
					save()
				}

			case "switch":
				if arg2 != "" {
					fmt.Printf("\nAre you sure you want do delete switch %s? [y/N]: ", arg2)
					inScanner.Scan()
					confirmation := inScanner.Text()
					confirmation = strings.ToUpper(confirmation)
					if confirmation == "Y" {
						delSwitch(arg2)
						save()
					}
				}

			case "host":
				if arg2 != "" {
					fmt.Printf("\nAre you sure you want do delete host %s? [y/N]: ", arg2)
					inScanner.Scan()
					confirmation := inScanner.Text()
					confirmation = strings.ToUpper(confirmation)
					if confirmation == "Y" {
						delHost(arg2)
						save()
					}
				}

			default:
				fmt.Println(" Usage: del <host|switch|router> <hostname>")
			}

		case "link":
			if (arg1 == "host") && (arg2 != "") && (arg3 != "") {
				linkHostTo(arg2, arg3)
				save()
			} else if (arg1 == "switch") && (arg2 != "") && (arg3 != "") {
				linkSwitchTo(arg2, arg3)
				save()
			} else {
				fmt.Println(" Usage: link <host|switch> <hostname> <router_hostname>")
			}

		case "unlink":
			if (arg1 == "host") && (arg2 != "") {
				unlinkHost(arg2)
				save()
			} else {
				fmt.Println(" Usage: unlink host <hostname>")
			}

		case "control", "c":
			if arg1 != "" {
				switch arg1 {
				case "host":
					controlHost(arg2)
					save()

				case "switch":
					controlSwitch(arg2)
					save()

				case "router":
					controlRouter(arg2)
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
				if arg1 == "info" {
					if arg2 != "" {
						printAchievementInfo(arg2)
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
			loadNetwork(Net().Name, "user")

		case "show", "sh":
			switch commandString {
			case "show overview", "sh overview":
				overview()

			case "show diagram", "sh diagram":
				drawDiagram(Net().Router.ID)

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
			fmt.Println(Net(), "")

		case "debug":
			if arg1 != "" {
				setDebug(arg1)
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

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		fmt.Println(fileScanner.Text())
	}
	if err := fileScanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
}

// returns the argument at idx or an empty string if out of range
func safeArg(args []string, idx int) string {
	if len(args) > idx {
		return args[idx]
	}
	return ""
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

	for h := range Net().Hosts {
		for range Net().Hosts[h].Interfaces {
			<-EngineInstance().ListenSync
		}
	}
	fmt.Printf("\n[Notice] Debug level is set to %d\n", getDebug())
	fmt.Printf("[Notice] Please note that switches can't yet link to routers or other switches.\n")

	fmt.Println("\nltdnetOS:")

	actionsMenu()
}
