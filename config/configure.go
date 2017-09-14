package config

import (
	"os/user"

	"github.com/spf13/viper"
)

//Configure loads the configuration from enviroment, 'CWD/.tvio' and ''/home/USER/.tvio'. 
func Configure() (cfg *UserConfig) {
	v := viper.New()

	v.SetDefault("debug", false)
	v.SetDefault("logger", "none")
	v.SetDefault("interface", "127.0.0.1")

	//Enviroment
	v.SetEnvPrefix("tvio")
	v.AutomaticEnv()

	//Configfile
	v.SetConfigName(".tvio")
	v.AddConfigPath(".") // First look in CWD

	usr, err := user.Current()
	if err != nil {
		v.AddConfigPath(usr.HomeDir) // Then in user home
	}
	cfg = &UserConfig{}
	v.Unmarshal(cfg)
	return
}
