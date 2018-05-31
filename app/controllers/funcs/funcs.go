package funcs

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"path/filepath"

	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/revel/revel"
)

// HttpGetBashParams takes a url and attempts to toml decode each line of the response.
// For each line that successfully gets decoded with toml, it is added to the map returned.
func HttpGetBashParams(u string) (map[string]string, error) {
	resp, err := http.Get(u)
	if err != nil {
		return map[string]string{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return map[string]string{}, err
	}

	resp.Body.Close()

	params := make(map[string]string)
	lines := bytes.Split(body, []byte{'\n'})
	for _, line := range lines {
		param := make(map[string]string)
		toml.Unmarshal(line, &param)

		for k, v := range param {
			if k != "" {
				params[k] = v
			}
			break
		}
	}

	return params, nil
}

// LoopBashParams loops through each line and attempts to decode each line of a string.
// For each line that successfully gets decoded, it will call the function (2nd arg) parsing
// the key and value as the parameters. If false is returned, the looping stops.
func LoopBashParams(body string, each func(string, string) bool) {
	for _, line := range strings.Split(body, "\n") {
		param := make(map[string]string)
		toml.Unmarshal([]byte(line), &param)

		for k, v := range param {
			if strings.Trim(k, " ") != "" {
				if !each(k, v) {
					return
				}
				break
			}
		}
	}
}

// ReplaceBashParams loops through each line in a string and allows a user to return the new values.
// The original string is returned but with each line raplced accordingly.
func ReplaceBashParams(body string, each func(string, string) (string, string)) (newBody string) {
	for _, line := range strings.Split(body, "\n") {
		newLine := line

		param := make(map[string]string)
		toml.Unmarshal([]byte(line), &param)

		for k, v := range param {
			if strings.Trim(k, " ") != "" {
				newKey, newValue := each(k, v)
				newLine = fmt.Sprintf("%s=\"%s\"", newKey, newValue)
				break
			}
		}

		newBody += newLine + "\n"
	}

	return
}

// MakeTimestamp returns an int64 unix timestamp.
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// HashPassword hashes a password mixed with a username.
func HashPassword(username, password string) string {
	h := md5.New()
	io.WriteString(h, password)

	pwMD5 := fmt.Sprintf("%x", h.Sum(nil))

	salt := "!$11kv$nGw@!hhfE3XMJbuSSyN4APushP$ZNt2glbEC!J&%uD7FTGLykiiN4W%TA"
	io.WriteString(h, salt)
	io.WriteString(h, username)
	io.WriteString(h, pwMD5)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetPointerStr is a helper function which returns the value of a *string if the value of str is not nil.
func GetPointerStr(str *string) string {
	if str != nil {
		return *str
	}
	return ""
}

// GetPointerInt is a helper function which returns the value of a *int if the value of i is not nil.
func GetPointerInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

// CreateLogFIle creates a path for a log file
func CreateLogFile(userID, serverID, machineID int, task string) (*os.File, error) {
	logPath := filepath.Join(
		revel.BasePath,
		"private",
		"logs",
		fmt.Sprintf("%d_%d_%d_%d_%s.log",
			userID,
			serverID,
			machineID,
			MakeTimestamp()/1000,
			task,
		),
	)
	return os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
}

// var colorRegex = regexp.MustCompile(`[\x1B\x10](?:\\\d{1,3})?\[\d{0,3}\w`)
var colorRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func WriteLogFile(content []byte, userID, serverID, machineID int, task string) (*os.File, error) {
	logFile, err := CreateLogFile(userID, serverID, machineID, task)
	if err != nil {
		return logFile, err
	}
	defer logFile.Close()

	content = colorRegex.ReplaceAll(content, []byte(""))
	content = bytes.Replace(content, []byte("\r\033[K"), []byte{'\r'}, -1)

	_, err = logFile.Write(content)

	return logFile, err
}
