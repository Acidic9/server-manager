package controllers

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"strings"

	"server-manager-revel/app/controllers/database"

	"io/ioutil"

	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

type simpleJsonResp struct {
	Success bool   `json:"success"`
	Err     string `json:"error,omitempty"`
}

type sqlResult struct {
	LastInsertID int
	RowsAffected int
}

// /
func (c App) Index() revel.Result {
	User, err := getUserFromSession(c.Session)
	if err != nil {
		return c.Redirect(App.Logout)
	}

	return c.Render(User)
}

// /servers
func (c App) Servers() revel.Result {
	User, err := getUserFromSession(c.Session)
	if err != nil {
		revel.WARN.Println(err)
		return c.Redirect(App.Logout)
	}

	//rows, err := database.DB.Query("SELECT id, machine_id, (SELECT IFNULL(title, CONCAT(username, ' @ ', address, ':', port)) FROM machines WHERE id=servers.machine_id) AS 'machine_title', game, status FROM servers WHERE servers.id IN (SELECT server_id FROM server_owners WHERE user_id=?) AND machine_id IN (SELECT machine_id FROM machine_owners WHERE user_id=?) ORDER BY machine_id ASC", User.ID, User.ID)
	rows, err := database.DB.Query("SELECT id, IFNULL(title, CONCAT(username, ' @ ', address, ':', port)) AS 'machine_title' FROM machines WHERE id IN (SELECT machine_id FROM machine_owners WHERE user_id=?) ORDER BY id ASC", User.ID)
	if err != nil && err != sql.ErrNoRows {
		revel.ERROR.Println(err)
		c.Flash.Error("Something went wrong when retrieving the servers")
		return c.Render()
	}
	if err == nil {
		defer rows.Close()
	}

	Machines := make([]*machine, 0)

	for rows.Next() {
		var (
			id    int
			title string
		)

		rows.Scan(&id, &title)

		Machines = append(Machines, &machine{
			ID:    id,
			Title: title,
		})
	}

	return c.Render(User, Machines)
}

// /machines
func (c App) Machines() revel.Result {
	User, err := getUserFromSession(c.Session)
	if err != nil {
		revel.WARN.Println(err)
		return c.Redirect(App.Logout)
	}

	rows, err := database.DB.Query("SELECT id, IFNULL(title, CONCAT(username, '@', address, ':', port)) AS 'title', address, port, username FROM machines WHERE id IN (SELECT machine_id FROM machine_owners WHERE user_id=?) ORDER BY id ASC", User.ID)
	if err != nil && err != sql.ErrNoRows {
		revel.ERROR.Println(err)
		c.Flash.Error("Something went wrong when retrieving the machines")
		return c.Render()
	}
	if err == nil {
		defer rows.Close()
	}

	Machines := make([]*machine, 0)

	for rows.Next() {
		var (
			id       int
			title    string
			address  string
			port     int
			username string
		)

		rows.Scan(&id, &title, &address, &port, &username)

		Machines = append(Machines, &machine{
			ID:       id,
			Title:    title,
			Addr:     address,
			Port:     port,
			Username: username,
		})
	}

	return c.Render(User, Machines)
}

// /new-server
func (c App) NewServer() revel.Result {
	User, err := getUserFromSession(c.Session)
	if err != nil {
		revel.WARN.Println(err)
		return c.Redirect(App.Logout)
	}

	return c.Render(User)
}

// /new-machine
func (c App) NewMachine() revel.Result {
	User, err := getUserFromSession(c.Session)
	if err != nil {
		revel.WARN.Println(err)
		return c.Redirect(App.Logout)
	}

	return c.Render(User)
}

// /login
func (c App) Login() revel.Result {
	_, err := getUserFromSession(c.Session)
	if err == nil {
		c.Redirect(App.Index)
	}

	return c.Render()
}

// /logout
func (c App) Logout() revel.Result {
	c.Session = revel.Session{}

	return c.Redirect(App.Login)
}

// /logs/:logFile
func (c App) Logs() revel.Result {
	User, err := getUserFromSession(c.Session)
	if err != nil {
		return c.NotFound("Log file does not exist or you have insufficient permissions")
	}

	logFile := c.Params.Get("logFile")
	logFileParts := strings.Split(logFile, "_")
	if len(logFileParts) == 0 || strconv.Itoa(User.ID) != logFileParts[0] {
		return c.NotFound("Log file does not exist or you have insufficient permissions")
	}

	logFilePath := filepath.Join(revel.BasePath, "private", "logs", logFile)
	content, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		return c.Redirect(App.NotFound)
	}

	return c.RenderText(string(content))
}
