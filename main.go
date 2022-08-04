package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	dataframe "github.com/rocketlaunchr/dataframe-go"
	export "github.com/rocketlaunchr/dataframe-go/exports"
)

func createGameStateEntry(gs dem.GameState, roundStartTick int, bombPlantTick int, tickRate int) map[string]int {
	m := make(map[string]int)
	m["RoundTimeLeft"] = 0
	if bombPlantTick > 0 {
		m["IsBombPlanted"] = 1
		//fmt.Println("BombPlanted", gs.IngameTick(), bombPlantTick, tickRate)
		m["RoundTimeLeft"] = 40 - (gs.IngameTick()-bombPlantTick)/tickRate
	} else {
		m["IsBombPlanted"] = 0
		//fmt.Println("BombNotPlanted:", gs.IngameTick(), roundStartTick, tickRate)
		m["RoundTimeLeft"] = 115 - (gs.IngameTick()-roundStartTick)/tickRate
	}
	m["CTEquipmentValue"] = gs.TeamCounterTerrorists().CurrentEquipmentValue()
	m["TEquipmentValue"] = gs.TeamTerrorists().CurrentEquipmentValue()
	m["CTPlayersAlive"] = 0
	m["TPlayersAlive"] = 0
	m["CTNumHelmets"] = 0
	m["TNumHelmets"] = 0
	m["CTNumArmors"] = 0
	m["TNumArmors"] = 0
	m["CTNumDefuseKits"] = 0
	m["CTNumSmokes"] = 0
	m["TNumSmokes"] = 0
	m["CTNumMollies"] = 0
	m["TNumMollies"] = 0
	m["CTNumHEGrenades"] = 0
	m["TNumHEGrenades"] = 0
	m["CTNumFlashes"] = 0
	m["TNumFlashes"] = 0
	for _, participant := range gs.TeamTerrorists().Members() {
		if participant.IsAlive() {
			m["TPlayersAlive"] += 1
			if participant.HasHelmet() {
				m["TNumHelmets"] += 1
				m["TNumArmors"] += 1
			} else if participant.Armor() > 0 {
				m["TNumArmors"] += 1
			}
			for _, val := range participant.Inventory {
				switch val.Type {
				case common.EqSmoke:
					m["TNumSmokes"] += 1
				case common.EqIncendiary:
					m["CTNumMollies"] += 1
				case common.EqMolotov:
					m["TNumMollies"] += 1
				case common.EqHE:
					m["TNumHEGrenades"] += 1
				case common.EqFlash:
					m["TNumFlashes"] += 1
				default:
					continue
				}
			}
		}
	}
	for _, participant := range gs.TeamCounterTerrorists().Members() {
		if participant.IsAlive() {
			m["CTPlayersAlive"] += 1
			if participant.HasDefuseKit() {
				m["CTNumDefuseKits"] += 1
			}
			if participant.HasHelmet() {
				m["CTNumHelmets"] += 1
				m["CTNumArmors"] += 1
			} else if participant.Armor() > 0 {
				m["CTNumArmors"] += 1
			}
			for _, val := range participant.Inventory {
				switch val.Type {
				case common.EqSmoke:
					m["CTNumSmokes"] += 1
				case common.EqMolotov:
					m["CTNumMollies"] += 1
				case common.EqIncendiary:
					m["CTNumMollies"] += 1
				case common.EqHE:
					m["CTNumHEGrenades"] += 1
				case common.EqFlash:
					m["CTNumFlashes"] += 1
				default:
					continue
				}
			}
		}
	}
	return m
}

func createDataFrame(rounds []map[int]map[string]int) *dataframe.DataFrame {
	var columns []string
	columns = append(columns, "RoundTime", "RoundTimeLeft", "IsBombPlanted", "CTEquipmentValue", "TEquipmentValue", "CTPlayersAlive", "TPlayersAlive", "CTNumHelmets", "TNumHelmets", "CTNumArmors", "TNumArmors", "CTNumDefuseKits",
		"CTNumSmokes", "TNumSmokes", "CTNumMollies", "TNumMollies", "CTNumHEGrenades", "TNumHEGrenades", "CTNumFlashes", "TNumFlashes", "winner", "CTScore", "TScore")
	var columnsAsSeries []dataframe.Series

	for _, s := range columns {
		columnsAsSeries = append(columnsAsSeries, dataframe.NewSeriesInt64(s, nil))
	}
	df := dataframe.NewDataFrame(columnsAsSeries...)
	//df.Remove(0)
	for _, round := range rounds {
		for roundTime, roundInfo := range round {
			roundInfo["RoundTime"] = roundTime
			df.Append(nil, map[string]interface{}{
				"RoundTime":        roundInfo["RoundTime"],
				"RoundTimeLeft":    roundInfo["RoundTimeLeft"],
				"IsBombPlanted":    roundInfo["IsBombPlanted"],
				"CTEquipmentValue": roundInfo["CTEquipmentValue"],
				"TEquipmentValue":  roundInfo["TEquipmentValue"],
				"CTPlayersAlive":   roundInfo["CTPlayersAlive"],
				"TPlayersAlive":    roundInfo["TPlayersAlive"],
				"CTNumHelmets":     roundInfo["CTNumHelmets"],
				"TNumHelmets":      roundInfo["TNumHelmets"],
				"CTNumArmors":      roundInfo["CTNumArmors"],
				"TNumArmors":       roundInfo["TNumArmors"],
				"CTNumDefuseKits":  roundInfo["CTNumDefuseKits"],
				"CTNumSmokes":      roundInfo["CTNumSmokes"],
				"TNumSmokes":       roundInfo["TNumSmokes"],
				"CTNumMollies":     roundInfo["CTNumMollies"],
				"TNumMollies":      roundInfo["TNumMollies"],
				"CTNumHEGrenades":  roundInfo["CTNumHEGrenades"],
				"TNumHEGrenades":   roundInfo["TNumHEGrenades"],
				"CTNumFlashes":     roundInfo["CTNumFlashes"],
				"TNumFlashes":      roundInfo["TNumFlashes"],
				"winner":           roundInfo["winner"],
				"CTScore":          roundInfo["CTScore"],
				"TScore":           roundInfo["TScore"],
			})
		}
	}
	return df
}

func parseDemo(f *os.File, err error, matchName string, mapName string) {
	p := dem.NewParser(f)
	defer p.Close()

	var rounds []map[int]map[string]int
	roundEntry := make(map[int]map[string]int)
	var roundStartTick int
	var bombPlantedTick int
	roundStarts := 0

	// Register handler on kill events
	p.RegisterEventHandler(func(e events.Kill) {
		gs := p.GameState()
		if gs.IsMatchStarted() {
			roundEntry[(gs.IngameTick()-roundStartTick)/int(p.TickRate())] = createGameStateEntry(gs, roundStartTick, bombPlantedTick, int(p.TickRate()))
		}
	})

	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		roundStartTick = p.GameState().IngameTick()
		bombPlantedTick = -1
		roundEntry[0] = createGameStateEntry(p.GameState(), roundStartTick, bombPlantedTick, int(p.TickRate()))
		roundStarts += 1
	})

	p.RegisterEventHandler(func(e events.MatchStartedChanged) {
		fmt.Println("Match Start")
		roundStartTick = p.GameState().IngameTick()
		bombPlantedTick = -1
		roundEntry[0] = createGameStateEntry(p.GameState(), roundStartTick, bombPlantedTick, int(p.TickRate()))
		roundStarts += 1
	})

	p.RegisterEventHandler(func(e events.BombPlanted) {
		gs := p.GameState()
		bombPlantedTick = p.GameState().IngameTick()
		roundEntry[(gs.IngameTick()-roundStartTick)/int(p.TickRate())] = createGameStateEntry(gs, roundStartTick, bombPlantedTick, int(p.TickRate()))
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		gs := p.GameState()
		switch e.Winner {
		case common.TeamTerrorists:
			// Winner's score + 1 because it hasn't actually been updated yet
			fmt.Printf("Round finished: winnerSide=T  ; score=%d:%d\n", gs.TeamTerrorists().Score()+1, gs.TeamCounterTerrorists().Score())
			//Set winner to Terrorist (0), update all save states with winner and subsequet score
			for k := range roundEntry {
				roundEntry[k]["winner"] = 0
				roundEntry[k]["TScore"] = gs.TeamTerrorists().Score() + 1
				roundEntry[k]["CTScore"] = gs.TeamCounterTerrorists().Score()
			}
			rounds = append(rounds, roundEntry)
			roundEntry = make(map[int]map[string]int)
		case common.TeamCounterTerrorists:
			fmt.Printf("Round finished: winnerSide=CT ; score=%d:%d\n", gs.TeamCounterTerrorists().Score()+1, gs.TeamTerrorists().Score())
			//Set winner to Counter-Terrorist, update all save states with winner and subsequet score
			for k := range roundEntry {
				roundEntry[k]["winner"] = 1
				roundEntry[k]["TScore"] = gs.TeamTerrorists().Score()
				roundEntry[k]["CTScore"] = gs.TeamCounterTerrorists().Score() + 1
			}
			rounds = append(rounds, roundEntry)
			roundEntry = make(map[int]map[string]int)
		default:
			// Probably match medic or something similar
			fmt.Println("Round finished: No winner (tie)")
		}
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		log.Panic("failed to parse demo: ", err)
	}

	df := createDataFrame(rounds)
	ctx := context.TODO()
	sks := []dataframe.SortKey{
		{Key: "RoundTime", Desc: false},
		{Key: "CTScore", Desc: false},
		{Key: "TScore", Desc: false},
	}
	df.Sort(ctx, sks)
	fmt.Print(df.Table())
	outputF, nerr := os.OpenFile("C:\\Users\\John\\developer\\CSGORoundPredictor\\data\\"+mapName+"\\"+matchName+".csv", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if nerr != nil {
		panic(nerr)
	}
	defer outputF.Close()

	export.ExportToCSV(ctx, outputF, df)
}

func main() {
	fileName := os.Args[1]
	initDemoParse(fileName)
}

func initDemoParse(demoLoc string) {
	f, err := os.Open(demoLoc)
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}
	defer f.Close()
	s1 := strings.Split(demoLoc, "\\")
	s2 := strings.Split(s1[len(s1)-1], "-")
	matchName := strings.Split(s1[len(s1)-1], ".")[0]
	mapName := strings.Split(s2[len(s2)-1], ".")[0]
	parseDemo(f, err, matchName, mapName)
}
