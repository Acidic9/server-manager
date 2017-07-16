package games

import (
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"io/ioutil"
	"server-manager-revel/app/controllers/funcs"

	"sort"

	"github.com/revel/revel"
)

type Game struct {
	Abbr   string // cs
	File   string // csserver
	Title  string // Counter-Strike 1.6
	Engine string // goldsource
}

func LoadAllGames() {
	resp, err := http.Get("https://raw.githubusercontent.com/GameServerManagers/LinuxGSM/master/lgsm/data/serverlist.csv")
	if err != nil {
		revel.ERROR.Fatal(err)
	}
	defer resp.Body.Close()

	var (
		csvReader = csv.NewReader(resp.Body)
		games     []*Game
		wg        sync.WaitGroup
	)
	for {
		row, err := csvReader.Read()
		if err != nil {
			break
		}

		wg.Add(1)
		go func(row []string) {
			defer wg.Done()

			g := new(Game)
			if len(row) >= 1 {
				g.Abbr = row[0]
			}
			if len(row) >= 2 {
				g.File = row[1]
			}
			if len(row) >= 3 {
				g.Title = row[2]
			}

			resp, err := http.Get("https://raw.githubusercontent.com/GameServerManagers/LinuxGSM/master/lgsm/config-default/config-lgsm/" + g.File + "/_default.cfg")
			if err != nil {
				revel.ERROR.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				revel.ERROR.Fatal(err)
			}

			funcs.LoopBashParams(string(body), func(k, v string) bool {
				if k == "engine" {
					g.Engine = v
					return false
				}
				return true
			})

			games = append(games, g)
		}(row)
	}

	wg.Wait()

	sort.Slice(games, func(i, j int) bool {
		return games[i].Abbr < games[j].Abbr
	})

	serverlistFile, err := os.OpenFile(
		filepath.Join(revel.BasePath, "misc", "serverlist.csv"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		os.ModePerm)
	if err != nil {
		revel.ERROR.Fatal(err)
	}

	csvWriter := csv.NewWriter(serverlistFile)
	for _, g := range games {
		err := csvWriter.Write([]string{g.Abbr, g.File, g.Title, g.Engine})
		if err != nil {
			revel.ERROR.Fatal(err)
		}
	}

	csvWriter.Flush()

	err = csvWriter.Error()
	if err != nil {
		revel.ERROR.Fatal(err)
	}
}

func GetGame(query string) (*Game, error) {
	serverlist, err := os.Open(filepath.Join(revel.BasePath, "misc", "serverlist.csv"))
	if err != nil {
		return nil, err
	}
	defer serverlist.Close()

	r := csv.NewReader(serverlist)
	for {
		row, err := r.Read()
		if err != nil {
			break
		}

		for _, c := range row {
			if c == query {
				g := new(Game)
				if len(row) >= 1 {
					g.Abbr = row[0]
				}
				if len(row) >= 2 {
					g.File = row[1]
				}
				if len(row) >= 3 {
					g.Title = row[2]
				}
				if len(row) >= 4 {
					g.Engine = row[3]
				}
				return g, nil
			}
		}
	}

	return nil, errors.New("game not found")
}
