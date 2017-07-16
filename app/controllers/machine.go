package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"server-manager-revel/app/controllers/database"
	"server-manager-revel/app/controllers/funcs"
	"server-manager-revel/app/controllers/server-install"

	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	gossh "golang.org/x/crypto/ssh"
)

type Machine struct {
	*revel.Controller
}

type machine struct {
	ID       int
	Title    string
	Addr     string
	Port     int
	Username string
	Password string
}

func (m *machine) GetFullAddr() string {
	return fmt.Sprintf("%s:%d", m.Addr, m.Port)
}

// /machine/add
func (c Machine) Add() revel.Result {
	var resp simpleJsonResp

	User, err := getUserFromSession(c.Session)
	if err != nil {
		resp.Err = "not-logged-in"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	title := c.Params.Get("title")
	titlePtr := &title
	if strings.Trim(title, " ") == "" {
		titlePtr = nil
	}
	address := c.Params.Get("address")
	if !govalidator.IsIP(address) {
		resp.Err = "invalid-address"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	portStr := c.Params.Get("port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 9999 {
		resp.Err = "invalid-port"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	if !c.Validation.Required("username").Ok {
		resp.Err = "empty-username"
		resp.Success = false
		return c.RenderJSON(&resp)
	}
	username := c.Params.Get("username")
	password := c.Params.Get("password")

	_, err = insertMachine(&User.ID, nil, titlePtr, &address, &port, &username, &password)
	if err != nil {
		resp.Err = "database-error"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	resp.Err = ""
	resp.Success = true
	return c.RenderJSON(&resp)
}

// /machine/list
func (c Machine) List() revel.Result {
	var resp struct {
		simpleJsonResp
		Machines []*machine `json:"machines"`
	}

	User, err := getUserFromSession(c.Session)
	if err != nil {
		resp.Err = "not-logged-in"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	resp.Machines, err = ListMachines(User.ID)
	if err != nil && err != sql.ErrNoRows {
		revel.ERROR.Println(err)
		resp.Err = "database-error"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	resp.Err = ""
	resp.Success = true
	return c.RenderJSON(&resp)
}

// /machine/install-dependencies
func (c Machine) InstallDependencies(machineID int, rootPassword, game string) revel.Result {
	var resp struct {
		simpleJsonResp
	}

	User, err := getUserFromSession(c.Session)
	if err != nil {
		resp.Err = "not-logged-in"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	if machineID == 0 {
		resp.Err = "invalid-param-machineID"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	if rootPassword == "" {
		resp.Err = "invalid-param-rootPassword"
		resp.Success = false
		return c.RenderJSON(&resp)
	}

	m, err := selectMachine("id=? AND id IN (SELECT machine_id FROM machine_owners WHERE user_id=?)", machineID, User.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Success = false
			resp.Err = "machine-does-not-exist"
			return c.RenderJSON(&resp)
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	deps, err := serverInstall.CheckDeps(m.GetFullAddr(), "root", rootPassword, game)
	if err != nil {
		switch err.(type) {
		case *gossh.ExitError:
			revel.ERROR.Println(err)
			resp.Success = false
			resp.Err = "command-failed"
			return c.RenderJSON(&resp)

		default:
			revel.ERROR.Println(err)
			resp.Success = false
			resp.Err = "machine-connection-failed"
			return c.RenderJSON(&resp)
		}
	}
	if len(deps.Missing) > 0 {
		resp.Success = false
		resp.Err = "missing-deps"
		return c.RenderJSON(&resp)
	}

	resp.Success = true
	resp.Err = ""
	return c.RenderJSON(&resp)
}

// Use -1 as userID to retrieve all machines.
func ListMachines(userID int) ([]*machine, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if userID == -1 {
		rows, err = database.DB.Query("SELECT id, title, address, port, username FROM machines")
	} else {
		rows, err = database.DB.Query("SELECT id, title, address, port, username FROM machines WHERE id IN (SELECT machine_id FROM machine_owners WHERE user_id=?)", userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := make([]*machine, 0)

	for rows.Next() {
		m := new(machine)
		var title sql.NullString
		err := rows.Scan(&m.ID, &title, &m.Addr, &m.Port, &m.Username)
		if err != nil {
			return nil, err
		}

		m.Title = title.String

		machines = append(machines, m)
	}

	return machines, nil
}

func getMachineFromID(userID, id int) (*machine, error) {
	var m *machine
	err := cache.Get("machine_"+strconv.Itoa(id), &m)
	if err != nil {
		return selectMachine("id=? AND id IN (SELECT machine_id FROM machine_owners WHERE user_id=?)", id, userID)
	}

	return m, nil
}

func selectMachine(query string, args ...interface{}) (*machine, error) {
	var (
		id       int
		title    sql.NullString
		address  string
		port     int
		username string
		password string
	)

	err := database.DB.QueryRow("SELECT id, title, address, port, username, password FROM machines WHERE "+query, args...).
		Scan(&id, &title, &address, &port, &username, &password)
	if err != nil {
		return nil, err
	}

	m := &machine{
		ID:       id,
		Title:    title.String,
		Addr:     address,
		Port:     port,
		Username: username,
		Password: password,
	}
	go cache.Set("machine_"+strconv.Itoa(id), m, cache.DefaultExpiryTime)

	return m, nil
}

func insertMachine(userID, id *int, title, address *string, port *int, username, password *string) (sqlResult, error) {
	argStr := make([]string, 0, 6)
	argVals := make([]interface{}, 0, 6)
	if id != nil {
		argStr = append(argStr, "id")
		argVals = append(argVals, id)
	}
	if title != nil {
		argStr = append(argStr, "title")
		argVals = append(argVals, title)
	}
	if address != nil {
		argStr = append(argStr, "address")
		argVals = append(argVals, address)
	}
	if port != nil {
		argStr = append(argStr, "port")
		argVals = append(argVals, port)
	}
	if username != nil {
		argStr = append(argStr, "username")
		argVals = append(argVals, username)
	}
	if password != nil {
		argStr = append(argStr, "password")
		argVals = append(argVals, password)
	}
	if len(argStr) == 0 {
		return sqlResult{}, errors.New("all fields are nil")
	}
	argValStr := strings.Join(argStr, ", ")
	placeholders := make([]string, len(argStr))
	for i := 0; i < len(argStr); i++ {
		placeholders[i] = "?"
	}
	placeholderStr := strings.Join(placeholders, ", ")

	result, err := database.DB.Exec("INSERT INTO machines ("+argValStr+") VALUES ("+placeholderStr+")", argVals...)
	if err != nil {
		return sqlResult{}, err
	}

	machineID, err := result.LastInsertId()
	if err != nil {
		return sqlResult{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return sqlResult{LastInsertID: int(machineID)}, err
	}

	res := sqlResult{LastInsertID: int(machineID), RowsAffected: int(rowsAffected)}

	if userID != nil {
		result, err = database.DB.Exec("INSERT INTO machine_owners (user_id, machine_id) VALUES (?, ?)", userID, machineID)
		if err != nil {
			return res, err
		}
	}

	m := &machine{
		ID:       int(machineID),
		Title:    funcs.GetPointerStr(title),
		Addr:     funcs.GetPointerStr(address),
		Port:     funcs.GetPointerInt(port),
		Username: funcs.GetPointerStr(username),
		Password: funcs.GetPointerStr(password),
	}
	go cache.Set("machine_"+strconv.Itoa(int(machineID)), m, cache.DefaultExpiryTime)

	return res, nil
}

func updateMachine(userID *int, id int, title, address *string, port *int, username, password *string) (sqlResult, error) {
	argStr := make([]string, 0, 5)
	argVals := make([]interface{}, 0, 6)
	if title != nil {
		argStr = append(argStr, "title=?")
		argVals = append(argVals, title)
	}
	if address != nil {
		argStr = append(argStr, "address=?")
		argVals = append(argVals, address)
	}
	if port != nil {
		argStr = append(argStr, "port=?")
		argVals = append(argVals, port)
	}
	if username != nil {
		argStr = append(argStr, "username=?")
		argVals = append(argVals, username)
	}
	if password != nil {
		argStr = append(argStr, "password=?")
		argVals = append(argVals, password)
	}
	if len(argStr) == 0 {
		return sqlResult{}, errors.New("all fields are nil")
	}
	argVals = append(argVals, &id)
	argValStr := strings.Join(argStr, ", ")

	var userCheckStr string
	if userID != nil {
		userCheckStr = " AND id IN (SELECT machine_id FROM machine_owners WHERE user_id=?)"
		argVals = append(argVals, userID)
	}

	result, err := database.DB.Exec("UPDATE machines SET "+argValStr+" WHERE id=?"+userCheckStr, argVals...)
	if err != nil {
		return sqlResult{}, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return sqlResult{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return sqlResult{LastInsertID: int(lastInsertID)}, err
	}

	res := sqlResult{LastInsertID: int(lastInsertID), RowsAffected: int(rowsAffected)}

	m := &machine{
		ID:       id,
		Title:    funcs.GetPointerStr(title),
		Addr:     funcs.GetPointerStr(address),
		Port:     funcs.GetPointerInt(port),
		Username: funcs.GetPointerStr(username),
		Password: funcs.GetPointerStr(password),
	}
	go cache.Set("machine_"+strconv.Itoa(id), m, cache.DefaultExpiryTime)

	return res, nil
}

func deleteMachine(id int) (sqlResult, error) {
	go cache.Delete("machine_" + strconv.Itoa(id))

	_, err := database.DB.Exec("DELETE FROM machine_owners WHERE machine_id=?", id)
	if err != nil {
		return sqlResult{}, err
	}

	result, err := database.DB.Exec("DELETE FROM machines WHERE id=?", id)
	if err != nil {
		return sqlResult{}, err
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return sqlResult{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return sqlResult{LastInsertID: int(lastInsertID)}, err
	}

	return sqlResult{LastInsertID: int(lastInsertID), RowsAffected: int(rowsAffected)}, nil
}
