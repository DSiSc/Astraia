package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/DSiSc/craft/log"
)

const (
	// config file prefix
	ConfigPrefix = "light_client"
	// api gateway
	ApiHostName = "apigateway.hostname"
	ApiPort = "apigateway.port"
)


func LoadConfig() (config *viper.Viper) {
	config = viper.New()
	// for environment variables
	config.SetEnvPrefix(ConfigPrefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)

	config.SetConfigName("light_client")
	homePath, _ := Home()
	config.AddConfigPath(fmt.Sprintf("%s/.lightClient", homePath))
	// Path to look for the config file in based on GOPATH
	goPath := os.Getenv("GOPATH")
	for _, p := range filepath.SplitList(goPath) {
		config.AddConfigPath(filepath.Join(p, "src/github.com/DSiSc/lightClient/config"))
	}

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("error reading plugin config: %s", err))
	}
	return
}


func GetApiGatewayHostName() string {
	conf := LoadConfig()
	apiGatewayHostName := conf.GetString(ApiHostName)
	return apiGatewayHostName
}

func GetApiGatewayPort() string {
	conf := LoadConfig()
	apiGatewayPort := conf.GetString(ApiPort)
	return apiGatewayPort
}

func Home() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Error("sh -c eval echo ~$USER error.")
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		log.Error("blank output when reading home directory")
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		log.Error("Get home path error.")
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}
