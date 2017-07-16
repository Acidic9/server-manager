package serverInstall

import (
	"fmt"
	"strings"

	"server-manager-revel/app/controllers/funcs"
	"server-manager-revel/app/controllers/ssh"

	"errors"

	"github.com/revel/revel"
)

type Dependencies struct {
	Missing        []string `json:"missing"`
	InstallCommand string   `json:"installCommand"`
}

// CheckDeps returns the missing dependencies on a remote host for installing a specific server.
func CheckDeps(addr, user, password, game string) (*Dependencies, error) {
	conn, err := ssh.Connect(addr, user, password)
	if err != nil {
		return nil, err
	}

	output, err := conn.SendCommands(
		"rm -rf check-deps.sh",
		fmt.Sprintf("wget -q %s/public/sh/check-deps.sh", revel.Config.StringDefault("http.public_addr", "")),
		"chmod --quiet +x check-deps.sh",
		fmt.Sprintf("./check-deps.sh \"%s\"", game),
		"rm -rf check-deps.sh",
	)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(output))

	params := make(map[string]string)
	funcs.LoopBashParams(string(output), func(key, value string) bool {
		params[key] = value
		return true
	})

	deps := new(Dependencies)

	var found bool
	for k, v := range params {
		switch k {
		case "missing_deps":
			found = true
			missing := strings.Split(v, " ")
			for _, dep := range missing {
				if strings.Trim(dep, " ") != "" {
					deps.Missing = append(deps.Missing, dep)
				}
			}
		case "manual_install_deps_command":
			deps.InstallCommand = v
		}
	}

	if !found {
		return deps, errors.New("failed to retreive missing deps")
	}

	return deps, nil
}
