package main

import (
	"encoding/csv"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"server-manager-revel/app/controllers/funcs"
	"sync"

	"github.com/revel/revel"
)

type game struct {
	Abbr  string // cs
	File  string // csserver
	Title string // Counter-Strike 1.6
}

var (
	startSettingsRegex = regexp.MustCompile(`\#+\#+\n\#+\s?\bSettings\b\s?\#+\n\#+\n`)
	endSettingsRegex   = regexp.MustCompile(`\#+[\s\w]+\#+`)
	games              = make([]game, 0)
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	serverlist, err := os.Open(filepath.Join(revel.BasePath, "misc", "serverlist.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer serverlist.Close()

	var (
		r     = csv.NewReader(serverlist)
		games []*game
	)
	for {
		row, err := r.Read()
		if err != nil {
			break
		}

		g := new(game)
		if len(row) >= 1 {
			g.Abbr = row[0]
		}
		if len(row) >= 2 {
			g.File = row[1]
		}
		if len(row) >= 3 {
			g.Title = row[2]
		}
		games = append(games, g)
	}

	var wg sync.WaitGroup

	for _, g := range games {
		wg.Add(1)
		go func(g *game) {
			defer wg.Done()
			resp, err := http.Get("https://raw.githubusercontent.com/GameServerManagers/LinuxGSM/master/lgsm/config-default/config-lgsm/" + g.File + "/_default.cfg")
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}

			var (
				paramsTableRows string
				launchOpt       string
			)
			funcs.LoopBashParams(string(body), func(k, v string) bool {
				if k == "parms" {
					launchOpt = v
					return false
				}

				paramsTableRows += "		<tr>\n"
				paramsTableRows += "			<td><input class=\"input\" value=\"" + k + "\"></td>\n"
				paramsTableRows += "			<td><input class=\"input\" value=\"" + v + "\"></td>\n"
				paramsTableRows += "		</tr>\n\n"

				return true
			})

			var htmlFile string
			if paramsTableRows != "" {
				htmlFile += "<table class=\"table\">\n"
				htmlFile += "	<thead>\n"
				htmlFile += "		<tr>\n"
				htmlFile += "			<th>Parameter</th>\n"
				htmlFile += "			<th>Value</th>\n"
				htmlFile += "		</tr>\n"
				htmlFile += "	</thead>\n"
				htmlFile += "	<tbody>\n"
				htmlFile += paramsTableRows
				htmlFile += "	</tbody>\n"
				htmlFile += "</table>\n"
			}

			if launchOpt != "" {
				htmlFile += "<div class=\"field\">\n"
				htmlFile += "	<label class=\"label\">Launch Options</label>\n"
				htmlFile += "	<p class=\"control\">\n"
				htmlFile += "		<textarea class=\"textarea\">" + html.EscapeString(launchOpt) + "</textarea>\n"
				htmlFile += "	</p>\n"
				htmlFile += "</div>\n"
			}

			file := filepath.Join(revel.BasePath, "public", "inc", "game-setup", g.Abbr+".html")
			err = ioutil.WriteFile(file, []byte(htmlFile), os.FileMode(os.O_RDWR))
			if err != nil {
				log.Println(err)
			}

			fmt.Println(file + ": successfully written")
		}(g)
	}

	wg.Wait()
}
