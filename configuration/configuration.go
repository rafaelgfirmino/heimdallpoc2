package configuration

import (
	"fmt"
	viper2 "github.com/spf13/viper"
	"log"
	"os"
)

var Env = viper2.New()
var (
	ErrorFindWorkDirectory = "Error loading working directory"
	ErrorFileNotFaund      = "heimdall.toml, file not find"
)

func Load() {

	dir, err := os.Getwd()

	if err != nil {
		log.Fatal(ErrorFindWorkDirectory)
	}

	Env.SetConfigFile(fmt.Sprintf(`%s/heimdall.toml`, dir))
	err = Env.ReadInConfig()

	if err != nil {
		panic(ErrorFileNotFaund)
	}
}
