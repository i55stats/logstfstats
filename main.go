package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/bitly/go-simplejson"
)

type Stat struct {
	steamid string
	stat    uint64
}

type StatArr []*Stat

func (s StatArr) Len() int {
	return len(s)
}

func (s StatArr) Less(i, j int) bool {
	return s[i].stat > s[j].stat
}

func (s StatArr) Swap(i, j int) {
	tmp := s[i]
	s[i] = s[j]
	s[j] = tmp
}

func getJson(url string) (*simplejson.Json, error) {
	resp, err := http.Get(url)

	bytes, _ := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()
	json, err := simplejson.NewFromReader(strings.NewReader(string(bytes)))

	return json, err
}

func checkEmpty(flag string, name string) {
	if flag == "" {
		log.Fatalf("flag %s empty", name)
	}
}

func main() {
	var logs []*simplejson.Json

	var urls, stat string
	flag.StringVar(&urls, "urls", "", "logs.tf log")
	flag.StringVar(&stat, "stat", "", "player stat")

	flag.Parse()
	checkEmpty(urls, "urls")
	checkEmpty(stat, "stat")

	urlArr := strings.Split(urls, ",")
	for _, url := range urlArr {
		json, err := getJson("http://logs.tf/json/" + url[14:])
		if err != nil {
			log.Fatal(err)
		}
		logs = append(logs, json)
	}

	steamidNameMap := make(map[string]string)
	steamidStatsMap := make(map[string][](map[string]interface{}))
	matchesPlayerMap := make(map[string]uint64)

	var maxstat uint64
	for _, json := range logs {
		jsonmap, _ := json.Get("names").Map()
		for steamid, playerName := range jsonmap {
			// fmt.Println(steamid)
			// fmt.Println(player)
			steamidNameMap[steamid] = playerName.(string)
			matchesPlayerMap[steamid]++
			classArray, _ := json.Get("players").Get(steamid).Get("class_stats").Array()

			for _, stats := range classArray {
				// fmt.Println(steamid)
				// fmt.Println(stats)
				steamidStatsMap[steamid] = append(steamidStatsMap[steamid], stats.(map[string]interface{}))
			}
		}
	}

	var classKillsMap = make(map[string](map[string]uint))
	classes := []string{"scout", "soldier", "demoman"}
	var statTitleMap = map[string]string{
		"dmg":   "Damage",
		"kills": "Kills",
	}

	for steamid, stats := range steamidStatsMap { //loop over stats array of every steamid
		for _, classname := range classes { //loop over every class's name
			if _, exists := classKillsMap[classname]; exists {
				for _, class_stats := range stats { // for every class played in player's stats
					if class_stats["type"] == classname { //if player has played classname
						kills, _ := (class_stats[stat].(json.Number)).Int64()
						classKillsMap[classname][steamid] += uint(kills)
						if uint64(kills) > maxstat {
							maxstat = uint64(kills)
						}
					}
				}
			} else {
				classKillsMap[classname] = make(map[string]uint)
			}
		}
	}

	var classStatArrMap = make(map[string]StatArr)

	for _, class := range classes {
		for steamid, kills := range classKillsMap[class] {
			classStatArrMap[class] = append(classStatArrMap[class], &Stat{
				class,
				uint64(kills) / matchesPlayerMap[steamid],
			})
		}
	}

	//Sort the stats, Decreasing order
	sort.Sort(classStatArrMap["scout"])
	sort.Sort(classStatArrMap["soldier"])
	sort.Sort(classStatArrMap["demoman"])

	fmt.Println(`<barchart title="Bullet Graph" left="300">`)
	format := `<bitem name="%s" value="%d" color="blue"/>`
	for class, statArr := range classStatArrMap {
		fmt.Printf(`<bdata title="Average %s %s" showdata="true" color="red" unit="">`+"\n",
			strings.Title(class), statTitleMap[stat])
		for _, stat := range statArr {
			fmt.Printf(format+"\n",
				steamidNameMap[stat.steamid],
				stat.stat)
		}
		fmt.Println(`</bdata>`)
	}
	fmt.Println(`</barchart>`)
}
