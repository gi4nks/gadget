// Copyright © 2017 gi4nks <gi4nks@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gi4nks/quant"

	repos "github.com/streamtune/gadget/repos"
	utils "github.com/streamtune/gadget/utils"
)

var cfgFile string

var Parrot = quant.NewParrot("gadget")
var Utilities = utils.NewUtilities(*Parrot)
var Configuration = utils.NewConfiguration(*Parrot)
var Repository = &repos.Repository{}
var Commands = &repos.Commands{}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gadget",
	Short: "Gadget is a docker insight investigator.",
	Long:  `Gadget is a docker insight investigator.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		Parrot.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gadget.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".gadget") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		Parrot.Println("Using config file:", viper.ConfigFileUsed())
	}

	/* -------------------------- */
	/* initialize the application */
	/* -------------------------- */

	/*
		RepositoryDirectory   string
		RepositoryFile        string
		DebugMode             bool
	*/

	folder, err := quant.ExecutableFolder()

	if err != nil {
		Parrot.Error("Executable folder error", err)
	}

	if viper.GetString("repositoryDirectory") != "" {
		Configuration.RepositoryDirectory = folder + "/" + viper.GetString("repositoryDirectory")
	} else {
		Configuration.RepositoryDirectory = folder + "/" + Configuration.RepositoryDirectory
	}

	if viper.GetString("repositoryFile") != "" {
		Configuration.RepositoryFile = viper.GetString("repositoryFile")
	}

	Configuration.DebugMode = viper.GetBool("debugMode")
	if Configuration.DebugMode {
		Parrot = quant.NewVerboseParrot("gadget")
	}

	/*
		RestServerMode        string
		RestPort              int
		LocalDockerEndpoint   string
		MachineDockerEndpoint string
		UseDockerMachine      bool
		MachineDockerCertFile string
		MachineDockerKeyFile  string
		MachineDockerCAFile   string
	*/

	if viper.GetString("restServerMode") != "" {
		Configuration.RestServerMode = viper.GetString("restServerMode")
	}

	if viper.GetInt("restPort") >= 0 {
		Configuration.RestPort = viper.GetInt("restPort")
	}

	if viper.GetString("localDockerEndpoint") != "" {
		Configuration.LocalDockerEndpoint = viper.GetString("localDockerEndpoint")
	}

	if viper.GetString("machineDockerEndpoint") != "" {
		Configuration.MachineDockerEndpoint = viper.GetString("machineDockerEndpoint")
	}

	Configuration.UseDockerMachine = viper.GetBool("useDockerMachine")

	if viper.GetString("machineDockerCertFile") != "" {
		Configuration.MachineDockerCertFile = viper.GetString("machineDockerCertFile")
	}

	if viper.GetString("machineDockerKeyFile") != "" {
		Configuration.MachineDockerKeyFile = viper.GetString("machineDockerKeyFile")
	}

	if viper.GetString("machineDockerCAFile") != "" {
		Configuration.MachineDockerCAFile = viper.GetString("machineDockerCAFile")
	}

	Repository = repos.NewRepository(*Parrot, *Configuration, *Utilities)
}
