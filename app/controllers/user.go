package controllers

import (
	"database/sql"
	"errors"
	"strings"

	"strconv"

	"server-manager-revel/app/controllers/database"
	"server-manager-revel/app/controllers/funcs"

	"github.com/revel/revel"
	"github.com/revel/revel/cache"
)

type User struct {
	*revel.Controller
}

type user struct {
	ID       int
	Email    string
	Username string
	Password string
}

func (c User) Login() revel.Result {
	var resp simpleJsonResp

	_, err := getUserFromSession(c.Session)
	if err == nil {
		resp.Success = false
		resp.Err = "already-logged-in"
		return c.RenderJSON(&resp)
	}

	username := c.Params.Get("username")
	password := c.Params.Get("password")

	if !c.Validation.Required(username).Ok {
		resp.Success = false
		resp.Err = "empty-username"
		return c.RenderJSON(&resp)
	}

	if password != "" {
		password = funcs.HashPassword(username, password)
	}

	u, err := selectUser("username=? AND password=?", username, password)
	if err == sql.ErrNoRows {
		resp.Success = false
		resp.Err = "incorrect-credentials"
		return c.RenderJSON(&resp)
	}
	if err != nil {
		revel.ERROR.Println(err)
		resp.Success = false
		resp.Err = "database-error"
		return c.RenderJSON(&resp)
	}

	c.Session["id"] = strconv.Itoa(u.ID)

	resp.Success = true
	resp.Err = ""
	return c.RenderJSON(&resp)
}

func getUserFromID(id int) (*user, error) {
	var u *user
	err := cache.Get("user_"+strconv.Itoa(id), &u)
	if err != nil {
		return selectUser("id=?", id)
	}

	return u, nil
}

func getUserFromSession(s revel.Session) (*user, error) {
	if s["id"] == "" {
		return nil, errors.New("user not logged in")
	}

	id, err := strconv.Atoi(s["id"])
	if err != nil {
		return nil, err
	}

	return getUserFromID(id)
}

func selectUser(query string, args ...interface{}) (*user, error) {
	var (
		id       int
		email    sql.NullString
		username string
	)
	err := database.DB.QueryRow("SELECT id, email, username FROM users WHERE "+query, args...).
		Scan(&id, &email, &username)
	if err != nil {
		return nil, err
	}

	u := &user{
		ID:       id,
		Email:    email.String,
		Username: username,
	}
	go cache.Set("user_"+strconv.Itoa(id), u, cache.DefaultExpiryTime)

	return u, nil
}

func insertUser(id *int, email, username, password *string) (sqlResult, error) {
	argStr := make([]string, 0, 4)
	argVals := make([]interface{}, 0, 4)
	if id != nil {
		argStr = append(argStr, "id")
		argVals = append(argVals, id)
	}
	if email != nil {
		argStr = append(argStr, "email")
		argVals = append(argVals, email)
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

	result, err := database.DB.Exec("INSERT INTO users ("+argValStr+") VALUES ("+placeholderStr+")", argVals...)
	if err != nil {
		return sqlResult{}, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return sqlResult{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return sqlResult{LastInsertID: int(userID)}, err
	}

	res := sqlResult{LastInsertID: int(userID), RowsAffected: int(rowsAffected)}

	u := &user{
		ID:       int(userID),
		Email:    funcs.GetPointerStr(email),
		Username: funcs.GetPointerStr(username),
		Password: funcs.GetPointerStr(password),
	}
	go cache.Set("user_"+strconv.Itoa(int(userID)), u, cache.DefaultExpiryTime)

	return res, nil
}

func updateUser(id int, email, username, password *string) (sqlResult, error) {
	argStr := make([]string, 0, 3)
	argVals := make([]interface{}, 0, 4)
	if email != nil {
		argStr = append(argStr, "email=?")
		argVals = append(argVals, email)
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

	result, err := database.DB.Exec("UPDATE users SET "+argValStr+" WHERE id=?", argVals...)
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

	u := &user{
		ID:       id,
		Email:    funcs.GetPointerStr(email),
		Username: funcs.GetPointerStr(username),
		Password: funcs.GetPointerStr(password),
	}
	go cache.Set("user_"+strconv.Itoa(id), u, cache.DefaultExpiryTime)

	return res, nil
}

func deleteUser(id int) (sqlResult, error) {
	go cache.Delete("user_" + strconv.Itoa(id))

	result, err := database.DB.Exec("DELETE FROM users WHERE id=?", id)
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
