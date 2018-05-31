package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"encoding/json"

	"server-manager-revel/app/controllers/database"
	"server-manager-revel/app/controllers/funcs"
	"server-manager-revel/app/controllers/games"
	"server-manager-revel/app/controllers/server-install"
	"server-manager-revel/app/controllers/ssh"

	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	*revel.Controller
}

type server struct {
	ID         int
	MachineID  int
	Name       string
	Game       string
	Map        string
	Players    int
	MaxPlayers int
	Status     string
	Path       string
}

func (c Server) Start() revel.Result {
	var resp struct {
		simpleJsonResp
		LogFile string `json:"logFile,omitempty"`
	}

	User, err := getUserFromSession(c.Session)
	if err != nil {
		resp.Success = false
		resp.Err = "not-logged-in"
		return c.RenderJSON(&resp)
	}

	serverIDStr := c.Params.Get("serverID")
	serverID, err := strconv.Atoi(serverIDStr)
	if err != nil {
		resp.Success = false
		resp.Err = "invalid-param-serverID"
		return c.RenderJSON(&resp)
	}

	machineIDStr := c.Params.Get("machineID")
	machineID, err := strconv.Atoi(machineIDStr)
	if err != nil {
		resp.Success = false
		resp.Err = "invalid-param-machineID"
		return c.RenderJSON(&resp)
	}

	m, err := getMachineFromID(User.ID, machineID)
	if err != nil {
		if err == sql.ErrNoRows {
			revel.ERROR.Println(err)
			resp.Success = false
			resp.Err = "machine-does-not-exist"
			return c.RenderJSON(&resp)
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	s, err := getServerFromID(User.ID, serverID, machineID)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Success = false
			resp.Err = "server-does-not-exist"
			return c.RenderJSON(&resp)
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	setServerStatus(serverID, machineID, proto.String("starting"))
	defer setServerStatus(serverID, machineID, nil)

	conn, err := ssh.Connect(m.GetFullAddr(), m.Username, m.Password)
	if err != nil {
		resp.Success = false
		resp.Err = "machine-connection-failed"
		return c.RenderJSON(&resp)
	}

	output, err := conn.SendCommands(
		fmt.Sprintf("./%s/%sserver start", s.Path, s.Game),
	)
	if err != nil {
		if len(output) > 0 {
			logFile, err := funcs.WriteLogFile(output, User.ID, serverID, machineID, "start")
			if err == nil {
				_, logFilePath := filepath.Split(logFile.Name())
				if logFilePath != "" {
					resp.LogFile = logFilePath
				}
			}
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "command-failed"
		return c.RenderJSON(&resp)
	}

	resp.Success = true
	resp.Err = ""
	return c.RenderJSON(&resp)
}

func (c Server) Stop() revel.Result {
	var resp struct {
		simpleJsonResp
		LogFile string `json:"logFile,omitempty"`
	}

	User, err := getUserFromSession(c.Session)
	if err != nil {
		resp.Success = false
		resp.Err = "not-logged-in"
		return c.RenderJSON(&resp)
	}

	serverIDStr := c.Params.Get("serverID")
	serverID, err := strconv.Atoi(serverIDStr)
	if err != nil {
		resp.Success = false
		resp.Err = "invalid-param-serverID"
		return c.RenderJSON(&resp)
	}

	machineIDStr := c.Params.Get("machineID")
	machineID, err := strconv.Atoi(machineIDStr)
	if err != nil {
		resp.Success = false
		resp.Err = "invalid-param-machineID"
		return c.RenderJSON(&resp)
	}

	m, err := getMachineFromID(User.ID, machineID)
	if err != nil {
		if err == sql.ErrNoRows {
			revel.ERROR.Println(err)
			resp.Success = false
			resp.Err = "machine-does-not-exist"
			return c.RenderJSON(&resp)
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	s, err := getServerFromID(User.ID, serverID, machineID)
	if err != nil {
		if err == sql.ErrNoRows {
			resp.Success = false
			resp.Err = "server-does-not-exist"
			return c.RenderJSON(&resp)
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	setServerStatus(serverID, machineID, proto.String("stopping"))
	defer setServerStatus(serverID, machineID, nil)

	conn, err := ssh.Connect(m.GetFullAddr(), m.Username, m.Password)
	if err != nil {
		resp.Success = false
		resp.Err = "machine-connection-failed"
		return c.RenderJSON(&resp)
	}

	output, err := conn.SendCommands(fmt.Sprintf("./%s/%sserver stop", s.Path, s.Game))
	if err != nil {
		if len(output) > 0 {
			logFile, err := funcs.WriteLogFile(output, User.ID, serverID, machineID, "stop")
			if err == nil {
				_, logFilePath := filepath.Split(logFile.Name())
				if logFilePath != "" {
					resp.LogFile = logFilePath
				}
			} else {
				revel.ERROR.Println(err)
			}
		}
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "command-failed"
		return c.RenderJSON(&resp)
	}

	resp.Success = true
	resp.Err = ""
	return c.RenderJSON(&resp)
}

func (c Server) Install() revel.Result {
	var resp struct {
		simpleJsonResp
		Deps *serverInstall.Dependencies `json:"dependencies"`
	}

	User, err := getUserFromSession(c.Session)
	if err != nil {
		resp.Success = false
		resp.Err = "not-logged-in"
		return c.RenderJSON(&resp)
	}

	if !c.Validation.Required("game").Ok {
		resp.Success = false
		resp.Err = "missing-param-game"
		return c.RenderJSON(&resp)
	}
	gameAbbr := c.Params.Get("game")

	if !c.Validation.Required("machine").Ok {
		resp.Success = false
		resp.Err = "missing-param-machine"
		return c.RenderJSON(&resp)
	}
	machineIDStr := c.Params.Get("machine")

	machineID, err := strconv.Atoi(machineIDStr)
	if err != nil {
		resp.Success = false
		resp.Err = "invalid-param-machine"
		return c.RenderJSON(&resp)
	}

	if !c.Validation.Required("launchOpt").Ok {
		resp.Success = false
		resp.Err = "missing-param-launchOpt"
		return c.RenderJSON(&resp)
	}
	launchOpt := c.Params.Get("launchOpt")

	if !c.Validation.Required("params").Ok {
		resp.Success = false
		resp.Err = "missing-param-params"
		return c.RenderJSON(&resp)
	}
	var params [][]string
	err = json.Unmarshal([]byte(c.Params.Get("params")), &params)
	if err != nil {
		resp.Success = false
		resp.Err = "invalid-param-params"
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

	g, err := games.GetGame(gameAbbr)
	if err != nil {
		revel.WARN.Println(err)
		resp.Success = false
		resp.Err = "game-not-found"
		return c.RenderJSON(&resp)
	}

	resp.Deps, err = serverInstall.CheckDeps(m.GetFullAddr(), m.Username, m.Password, g.File)
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
	if len(resp.Deps.Missing) > 0 {
		resp.Success = false
		resp.Err = "missing-deps"
		return c.RenderJSON(&resp)
	}

	serverPath := gameAbbr + "server"
	result, err := insertServer(&User.ID, nil, machineID, &gameAbbr, proto.String("installing"), &serverPath)
	if err != nil {
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	serverID := result.LastInsertID

	conn, err := ssh.Connect(m.GetFullAddr(), m.Username, m.Password)
	if err != nil {
		revel.ERROR.Println(err)
		go deleteServer(serverID, machineID)
		resp.Success = false
		resp.Err = "machine-connection-failed"
		return c.RenderJSON(&resp)
	}

	for i := 0; i < 9999; i++ {
		if i == 0 {
			serverPath = gameAbbr + "server"
		} else {
			serverPath = gameAbbr + "server" + strconv.Itoa(i)
		}
		_, err = conn.SendCommands(fmt.Sprintf("mkdir %s", serverPath))
		if err == nil {
			break
		}
	}

	_, err = updateServer(&User.ID, serverID, machineID, nil, nil, &serverPath)
	if err != nil {
		revel.ERROR.Println(err)
		go deleteServer(serverID, machineID)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	go func(conn *ssh.Connection, serverPath string, g *games.Game, serverID, machineID, userID int, params [][]string, launchOpt string) {
		output, err := conn.SendCommands(
			fmt.Sprintf("mkdir %s", serverPath),
			fmt.Sprintf("cd %s", serverPath),
			fmt.Sprintf("wget -O \""+g.File+"\" --no-check-certificate https://gameservermanagers.com/dl/linuxgsm.sh"),
			fmt.Sprintf(`sed -i -e "s/shortname=\"core\"/shortname=\"`+g.Abbr+`\"/g" "`+g.File+`"`),
			fmt.Sprintf(`sed -i -e "s/gameservername=\"core\"/gameservername=\"`+g.File+`\"/g" "`+g.File+`"`),

			fmt.Sprintf("chmod +x %s", g.File),
			fmt.Sprintf("./%s auto-install", g.File),
		)
		if err != nil {
			revel.ERROR.Println(err)
		}

		fmt.Println(string(output))

		var sedCommands []string

		for _, p := range params {
			var (
				key   string
				value string
			)
			if len(p) >= 1 {
				key = p[0]
			}
			if len(p) >= 2 {
				value = p[1]
			}
			if key != "" {
				sed := fmt.Sprintf(`sed -i -e '/^%s=/{h;s/=.*/="%s"/};${x;/^$/{s//%s="%s"/;H};x}' %s/lgsm/config-lgsm/%s/%s.cfg`, key, value, key, value, serverPath, g.File, g.File)
				sedCommands = append(sedCommands, sed)
			}
		}

		// TODO: Fix launchOpt sed
		// if launchOpt != "" {
		// 	// sed -i -e '/^fn_parms(){\nparms=/{h;s/=.*/="fff"/};${x;/^$/{s//fn_parms(){\nparms="ffff"/;H};x}'
		// 	sed := fmt.Sprintf(`sed -i -e '/^fn_parms(){\nparms=/{h;s/=.*/="%s"/};${x;/^$/{s//%s="%s"/;H};x}' %s/lgsm/config-lgsm/%s/%s.cfg`, key, value, key, value, serverPath, g.File, g.File)
		// 	sedCommands = append(sedCommands, sed)
		// }

		output, err = conn.SendCommands(sedCommands...)
		fmt.Println(string(output), err)

		err = setServerStatus(serverID, machineID, nil)
		if err != nil {
			log.Println(err)
		}
	}(conn, serverPath, g, int(serverID), machineID, User.ID, params, launchOpt)

	resp.Success = true
	resp.Err = ""
	return c.RenderJSON(&resp)
}

func (c Server) Status(ws *websocket.Conn) revel.Result {
	if ws == nil {
		return nil
	}

	User, err := getUserFromSession(c.Session)
	if err != nil {
		return nil
	}

	stillConnected := true
	go func() {
		for {
			if ws != nil {
				err := websocket.Message.Receive(ws, nil)
				if err != nil {
					stillConnected = false
					return
				}
			} else {
				return
			}
		}
	}()

	servers := make([]*server, 0)

	rows, err := database.DB.Query("SELECT id, machine_id, game, status, path FROM servers WHERE id IN (SELECT server_id FROM server_owners WHERE user_id=?) AND machine_id IN (SELECT machine_id FROM machine_owners WHERE user_id=?)", User.ID, User.ID)
	if err != nil {
		revel.ERROR.Println(err)
		return nil
	}

	for rows.Next() {
		var (
			id        int
			machineID int
			game      string
			status    sql.NullString
			path      string
		)

		err := rows.Scan(&id, &machineID, &game, &status, &path)
		if err != nil {
			revel.ERROR.Println(err)
			continue
		}

		servers = append(servers, &server{
			ID:        id,
			MachineID: machineID,
			Game:      game,
			Status:    status.String,
			Path:      path,
		})
	}

	rows.Close()

	for _, s := range servers {
		if s.Status == "" {
			s.Status = "..."
		}
		websocket.JSON.Send(ws, &s)
		if s.Status == "..." {
			s.Status = ""
		}
	}

	for stillConnected {
		rows, err := database.DB.Query("SELECT id, machine_id, game, status, path FROM servers WHERE id IN (SELECT server_id FROM server_owners WHERE user_id=?) AND machine_id IN (SELECT machine_id FROM machine_owners WHERE user_id=?)", User.ID, User.ID)
		if err != nil {
			revel.ERROR.Println(err)
			continue
		}

		for rows.Next() {
			var (
				id        int
				machineID int
				game      string
				status    sql.NullString
				path      string
			)

			err := rows.Scan(&id, &machineID, &game, &status, &path)
			if err != nil {
				revel.ERROR.Println(err)
				continue
			}

			var (
				exists    bool
				newServer = server{
					ID:        id,
					MachineID: machineID,
					Game:      game,
					Status:    status.String,
					Path:      path,
				}
			)

			for _, s := range servers {
				if s.ID == id && s.MachineID == machineID {
					exists = true
					oldStatus := s.Status
					*s = newServer
					if status.String != "" && oldStatus != status.String {
						websocket.JSON.Send(ws, &newServer)
					}
					break
				}
			}

			if !exists {
				websocket.JSON.Send(ws, &newServer)
				servers = append(servers, &newServer)
				continue
			}
		}

		rows.Close()

		machines := make([]*machine, 0)
		for _, s := range servers {
			var found bool
			for _, m := range machines {
				if s.MachineID == m.ID {
					found = true
					break
				}
			}

			if !found {
				machines = append(machines, &machine{ID: s.MachineID})
			}
		}

		for _, m := range machines {
			if m.Username == "" {
				m, err = selectMachine("id=?", m.ID)
				if err != nil {
					for _, s := range servers {
						if s.MachineID == m.ID && s.Status == "" {
							s.Status = "offline"
							websocket.JSON.Send(ws, s)
						}
					}
					continue
				}
			}

			go func(m machine) {
				conn, err := ssh.Connect(m.GetFullAddr(), m.Username, m.Password, time.Second*6)
				if err != nil {
					for _, s := range servers {
						if s.MachineID == m.ID && s.Status == "" {
							s.Status = "offline"
							websocket.JSON.Send(ws, s)
						}
					}
					return
				}

				var serverWg sync.WaitGroup
				for _, s := range servers {
					if s.MachineID == m.ID && s.Status == "" {
						serverWg.Add(1)
						go func(s server) {
							defer serverWg.Done()

							output, err := conn.SendCommands(fmt.Sprintf("./%s/%sserver details", s.Path, s.Game))
							if err != nil {
								s.Status = "offline"
								websocket.JSON.Send(ws, s)
								return
							}

							lines := strings.Split(string(output), "\n")
							for i := len(lines) - 1; i >= 0; i-- {
								if strings.Contains(lines[i], "ONLINE") {
									s.Status = "running"
									websocket.JSON.Send(ws, s)
									return
								} else if strings.Contains(lines[i], "OFFLINE") {
									s.Status = "stopped"
									websocket.JSON.Send(ws, s)
									return
								}
							}

							s.Status = "unknown"
							websocket.JSON.Send(ws, s)
						}(*s)
					}
				}

				serverWg.Wait()
			}(*m)
		}

		time.Sleep(time.Second * 5)
	}

	return nil
}

func setServerStatus(id, machineID int, status *string) error {
	_, err := database.DB.Exec("UPDATE servers SET status=? WHERE id=? AND machine_id=?", status, id, machineID)
	if err != nil {
		return err
	}

	_, err = selectServer("id=? AND machine_id=?", id, machineID)
	return err
}

func getServerFromID(userID, id, machineID int) (*server, error) {
	var s *server
	err := cache.Get("server", &s)
	if err != nil {
		s, err := selectServer("id=? AND machine_id=? AND id IN (SELECT server_id FROM server_owners WHERE user_id=?)", id, machineID, userID)
		if err != nil {
			return nil, err
		}

		go cache.Set("server_"+strconv.Itoa(id)+"_"+strconv.Itoa(machineID), s, cache.DefaultExpiryTime)
		return s, nil
	}

	return s, nil
}

func selectServer(query string, args ...interface{}) (*server, error) {
	var (
		id        int
		machineID int
		game      string
		status    sql.NullString
		path      string
	)
	err := database.DB.QueryRow("SELECT id, machine_id, game, status, path FROM servers WHERE "+query, args...).Scan(&id, &machineID, &game, &status, &path)
	if err != nil {
		return nil, err
	}

	s := &server{
		ID:        id,
		MachineID: machineID,
		Game:      game,
		Status:    status.String,
		Path:      path,
	}
	go cache.Set("server_"+strconv.Itoa(id)+"_"+strconv.Itoa(machineID), s, cache.DefaultExpiryTime)

	return s, nil
}

func insertServer(userID, id *int, machineID int, game, status, path *string) (sqlResult, error) {
	argStr := make([]string, 0, 5)
	argVals := make([]interface{}, 0, 5)
	if id != nil {
		argStr = append(argStr, "id")
		argVals = append(argVals, id)
	}
	argStr = append(argStr, "machine_id")
	argVals = append(argVals, &machineID)
	if game != nil {
		argStr = append(argStr, "game")
		argVals = append(argVals, game)
	}
	if status != nil {
		argStr = append(argStr, "status")
		argVals = append(argVals, status)
	}
	if path != nil {
		argStr = append(argStr, "path")
		argVals = append(argVals, path)
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

	result, err := database.DB.Exec("INSERT INTO servers ("+argValStr+") VALUES ("+placeholderStr+")", argVals...)
	if err != nil {
		return sqlResult{}, err
	}

	serverID, err := result.LastInsertId()
	if err != nil {
		return sqlResult{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return sqlResult{LastInsertID: int(serverID)}, err
	}

	res := sqlResult{LastInsertID: int(serverID), RowsAffected: int(rowsAffected)}

	if userID != nil {
		result, err = database.DB.Exec("INSERT INTO server_owners (user_id, server_id, machine_id) VALUES (?, ?, ?)", userID, serverID, machineID)
		if err != nil {
			return res, err
		}
	}

	s := &server{
		ID:        int(serverID),
		MachineID: machineID,
		Game:      funcs.GetPointerStr(game),
		Status:    funcs.GetPointerStr(status),
		Path:      funcs.GetPointerStr(path),
	}
	go cache.Set("server_"+strconv.Itoa(int(serverID))+"_"+strconv.Itoa(machineID), s, cache.DefaultExpiryTime)

	return res, nil
}

func updateServer(userID *int, id, machineID int, game, status, path *string) (sqlResult, error) {
	argStr := make([]string, 0, 3)
	argVals := make([]interface{}, 0, 5)
	if game != nil {
		argStr = append(argStr, "game=?")
		argVals = append(argVals, game)
	}
	if status != nil {
		argStr = append(argStr, "status=?")
		argVals = append(argVals, status)
	}
	if path != nil {
		argStr = append(argStr, "path=?")
		argVals = append(argVals, path)
	}
	if len(argStr) == 0 {
		return sqlResult{}, errors.New("all fields are nil")
	}
	argVals = append(argVals, &id)
	argVals = append(argVals, &machineID)
	argValStr := strings.Join(argStr, ", ")

	var userCheckStr string
	if userID != nil {
		userCheckStr = " AND id IN (SELECT server_id FROM server_owners WHERE user_id=? AND machine_id=?)"
		argVals = append(argVals, userID, machineID)
	}

	result, err := database.DB.Exec("UPDATE servers SET "+argValStr+" WHERE id=? AND machine_id=?"+userCheckStr, argVals...)
	if err != nil {
		return sqlResult{}, err
	}

	serverID, err := result.LastInsertId()
	if err != nil {
		return sqlResult{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return sqlResult{LastInsertID: int(serverID)}, err
	}

	res := sqlResult{LastInsertID: int(serverID), RowsAffected: int(rowsAffected)}

	s := &server{
		ID:        int(serverID),
		MachineID: machineID,
		Game:      funcs.GetPointerStr(game),
		Status:    funcs.GetPointerStr(status),
		Path:      funcs.GetPointerStr(path),
	}
	go cache.Set("server_"+strconv.Itoa(int(serverID))+"_"+strconv.Itoa(machineID), s, cache.DefaultExpiryTime)

	return res, nil
}

func deleteServer(id, machineID int) (sqlResult, error) {
	go cache.Delete("server_" + strconv.Itoa(id) + "_" + strconv.Itoa(machineID))

	_, err := database.DB.Exec("DELETE FROM server_owners WHERE server_id=? AND machine_id=?", id, machineID)
	if err != nil {
		return sqlResult{}, err
	}

	result, err := database.DB.Exec("DELETE FROM servers WHERE id=? AND machine_id=?", id, machineID)
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
