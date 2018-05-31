package app

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"server-manager-revel/app/controllers/database"

	"server-manager-revel/app/controllers/games"

	"github.com/revel/revel"
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.ActionInvoker,           // Invoke the action.
	}

	// register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)

	revel.OnAppStart(database.InitDB)
	revel.OnAppStart(games.LoadAllGames)
	revel.OnAppStart(InitTmplFuncs)
	revel.OnAppStart(CleanLogs)
}

// HeaderFilter adds common security headers
// TODO turn this into revel.HeaderFilter
// should probably also have a filter for CSRF
// not sure if it can go in the same filter or not
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

func InitTmplFuncs() {
	revel.TemplateFuncs["isset"] = func(v interface{}) bool {
		if v != nil {
			return true
		}
		return false
	}

	revel.TemplateFuncs["strBefore"] = func(s, sep string) string {
		return strings.Split(s, sep)[0]
	}

	revel.TemplateFuncs["capitalize"] = func(s string) string {
		return strings.Title(s)
	}
}

func CleanLogs() {
	revel.INFO.Println("Cleaning all log files")

	files, err := ioutil.ReadDir(filepath.Join(revel.BasePath, "private", "logs"))
	if err != nil {
		revel.ERROR.Println(err)
	} else {
		for _, logFile := range files {
			os.Remove(filepath.Join(
				revel.BasePath,
				"private",
				"logs",
				logFile.Name(),
			))
		}
	}

	go func() {
		for {
			time.Sleep(time.Minute * 5)

			revel.INFO.Println("Cleaning log files older than 24 hrs")
			files, err := ioutil.ReadDir(filepath.Join(revel.BasePath, "private", "logs"))
			if err != nil {
				revel.ERROR.Println(err)
				continue
			}

			for _, logFile := range files {
				if logFile.IsDir() {
					continue
				}

				{
					logFileNameParts := strings.Split(strings.TrimSuffix(logFile.Name(), ".log"), "_")
					if len(logFileNameParts) < 4 {
						goto deleteFile
					}

					timestamp, err := strconv.ParseInt(logFileNameParts[3], 10, 64)
					if err != nil {
						goto deleteFile
					}

					t := time.Unix(timestamp, 0)

					if time.Now().Sub(t).Hours() <= 24 {
						continue
					}
				}

			deleteFile:
				err = os.Remove(filepath.Join(
					revel.BasePath,
					"private",
					"logs",
					logFile.Name(),
				))
				if err != nil {
					revel.ERROR.Println(err)
				}

			}
		}
	}()
}
