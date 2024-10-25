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

func intro() {
	fmt.Println("ltdnet v0.3.0")
	fmt.Println("by vincebel")
	fmt.Println("\nPlease note that switch functionality is limited and in development")
}

func startMenu() {
	selection := false
	fmt.Println("\nPlease create or select a network:")
	fmt.Println(" 1) Create new network")
	fmt.Println(" 2) Select saved network")
	for !selection {
		fmt.Print("\nAction: ")

		scanner.Scan()
		option := scanner.Text()

		switch strings.ToUpper(option) {
		case "1", "C", "NEW", "CREATE":
			selection = true
			newNetworkPrompt()
		case "2", "S", "SELECT":
			selection = true
			selectNetwork()
		default:
			fmt.Println("Not a valid option. Options: 1, 2")
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
			fmt.Printf("\nAre you sure you want do delete router %s? [y/n]: ", actionword3)
			scanner.Scan()
			confirmation := scanner.Text()
			confirmation = strings.ToUpper(confirmation)
			if confirmation == "Y" {
				delRouter() // Only one router per network currently
				save()
			}

		case "switch":
			if actionword3 != "" {
				fmt.Printf("\nAre you sure you want do delete switch %s? [y/n]: ", actionword3)
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
				fmt.Printf("\nAre you sure you want do delete host %s? [y/n]: ", actionword3)
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

	case "save":
		save()

	case "reload":
		loadNetwork(snet.Name, "user")

	case "show", "sh":
		switch action_selection {
		case "show network overview", "sh network overview":
			overview()

		case "show diagram":
			drawDiagram(snet.Router.ID)

		default:
			if len(action_selection) > 12 { // show device
				show(action_selection[12:])
			} else {
				fmt.Println(" Usage: show network overview\n\tshow device <hostname>\n\tshow diagram")
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
				"4 - All sorts of garbage (dev use)")
		}

	case "manual", "man":
		launchManual()

	case "exit", "quit", "q":
		os.Exit(0)

	case "help", "?":
		fmt.Println("",
			"show <args>\t\tDisplays information\n",
			"add <args>\t\tAdds device to network\n",
			"del <args>\t\tRemoves device from network\n",
			"link <args>\t\tLinks two devices\n",
			"unlink <args>\t\tUnlinks two devices\n",
			"control <args>\t\tLogs in as device\n",
			"save\t\t\tManually saves network changes\n",
			"reload\t\t\tReloads the network file (May fix runtime bugs)\n",
			"debug <0-4>\t\tSets debug level. Default is 1\n",
			"manual\t\t\tLaunches the user manual. Great for beginners!\n",
			"exit\t\t\tExits the program",
			//"netdump\t\tPrints loaded Network object (developer use)\n", HIDDEN
		)

	default:
		fmt.Println(" Invalid command. Type 'help' for a list of commands.")
	}
}

func main() {
	intro()
	startMenu()
	go Listener()

	for range snet.Hosts {
		<-listenSync
	}
	fmt.Printf("\n[Notice] Debug level is set to %d\n", getDebug())
	fmt.Println("\nltdnetOS:")

	for {
		actionsMenu()
	}
}
