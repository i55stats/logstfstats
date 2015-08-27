package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bitly/go-simplejson"
)

type Class int

const (
	Scout   Class = 0
	Soldier Class = 1
	Pyro    Class = 2

	Engi  Class = 3
	Heavy Class = 4
	Demo  Class = 5

	Sniper Class = 6
	Medic  Class = 7
	Spy    Class = 8
)

var stringClassMap = map[string]Class{
	"scout":   Scout,
	"soldier": Soldier,
	"pyro":    Pyro,

	"demoman":      Demo,
	"heavyweapons": Heavy,
	"engineer":     Engi,

	"sniper": Sniper,
	"medic":  Medic,
	"spy":    Spy,
}

var classStringMap = map[Class]string{
	Scout:   "scout",
	Soldier: "soldier",
	Pyro:    "pyro",

	Demo:  "demoman",
	Heavy: "heavyweapons",
	Engi:  "engineer",

	Sniper: "sniper",
	Medic:  "medic",
	Spy:    "spy",
}

type steamidKills struct {
	teamid string
	Kills  uint
}

func getJson(url string) (*simplejson.Json, error) {
	resp, err := http.Get(url)

	bytes, _ := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()
	json, err := simplejson.NewFromReader(strings.NewReader(string(bytes)))

	return json, err
}

func getClassesPlayed(player *simplejson.Json) []Class {
	class_stats, _ := player.Get("class_stats").Array()
	var classes []Class

	for _, class := range class_stats {
		classString := class.(map[string]interface{})["type"].(string)
		classes = append(classes, stringClassMap[classString])
	}
	return classes
}

func main() {
	logs := make([]*simplejson.Json, len(os.Args)-1)

	fmt.Print("Getting logs...")
	for i := 1; i <= len(os.Args)-1; i++ {
		json, err := getJson("http://logs.tf/json/" + (os.Args[i])[15:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		logs[i-1] = json
	}

	fmt.Println("Done.")

	steamidNameMap := make(map[string]string)
	steamidStatsMap := make(map[string][](map[string]interface{}))

	for _, json := range logs {
		jsonmap, _ := json.Get("names").Map()
		for steamid, playerName := range jsonmap {
			// fmt.Println(steamid)
			// fmt.Println(player)
			steamidNameMap[steamid] = playerName.(string)
			// steamidClassMap[steamid] = getClassesPlayed(json.Get("players").Get(steamid))
			classArray, _ := json.Get("players").Get(steamid).Get("class_stats").Array()

			for _, stats := range classArray {
				// fmt.Println(steamid)
				// fmt.Println(stats)
				steamidStatsMap[steamid] = append(steamidStatsMap[steamid], stats.(map[string]interface{}))
			}
		}
	}

	var classKillsMap = make(map[string](map[string]uint))
	classes := []string{"scout", "soldier", "pyro", "engineer", "heavyweapons",
		"demoman", "sniper", "medic", "spy"}

	for steamid, stats := range steamidStatsMap { //loop over stats array of every steamid
		for _, classname := range classes { //loop over every class's name
			if _, exists := classKillsMap[classname]; exists {
				for _, class_stats := range stats { // for every class played in player's stats
					if class_stats["type"] == classname { //if player has played classname
						kills, _ := (class_stats["kills"].(json.Number)).Int64()
						classKillsMap[classname][steamid] += uint(kills)
					}
				}
			} else {
				classKillsMap[classname] = make(map[string]uint)
			}
		}
	}

	i := 0
	for class, stats := range classKillsMap {
		fmt.Println(class)
		for steamid, kills := range stats {
			if kills == 0 {
				continue
			}
			fmt.Printf("%d \"%s\" %d\n", i, shorten(steamidNameMap[steamid], 15), kills)
			i += 1
		}
		i = 0
	}
	// for steamid, classes := range steamidClassMap {
	// 	fmt.Print(steamidNameMap[steamid] + ": ")
	// 	for _, i := range classes {
	// 		fmt.Print(classStringMap[i] + " ")
	// 	}
	// 	fmt.Println("")
	// }

}

func shorten(str string, to int) string {
	if len(str) < to {
		return str
	}
	return str[:to]
}
