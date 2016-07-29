// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func message(msg string) {
	t := time.Now()
	fmt.Println(fmt.Sprintf("[%s]\t%s", t.Local(), msg))
}

func generatePdf(input io.Reader, output io.Writer) error {
	message("Generating PDF")
	filename := fmt.Sprintf("%s/temp.pdf", os.TempDir())
	defer os.Remove(filename)

	cmd := exec.Command("wkhtmltopdf", "-", filename)
	cmd.Stdin = input

	err := cmd.Run()
	// TODO Log output somewhere
	if err != nil {
		return errors.New("Could not execute wkhtmltopdf command")
	}

	pdf, oErr := os.Open(filename)
	if oErr != nil {
		return errors.New("Could not open temp file for writing")
	}

	_, cErr := io.Copy(output, pdf)
	if cErr != nil {
		return errors.New("Could not return PDF body")
	}

	return nil
}

func serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")

	switch r.Method {
	case http.MethodPost:
		if err := generatePdf(r.Body, w); err != nil {
			message(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gowkhtmltopdf",
	Short: "Serves an html to pdf converter web service",
	Long: `Convert html to PDF by sending HTML markup as the request body.

Returns the markup converted to PDF in an application/pdf response.

e.g. curl -d "Hello, World" http://localhost:8000`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", serve)
		listenAddress := fmt.Sprintf("%s:%s", cmd.Flag("address").Value, cmd.Flag("port").Value)
		message(fmt.Sprintf("Listening on %s", listenAddress))
		http.ListenAndServe(listenAddress, nil)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gowkhtmltopdf.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.Flags().String("address", "127.0.0.1", "Address on which to listen for requests")
	RootCmd.Flags().String("port", "8000", "Port on which to listen for requests")
	// TODO Allow configuration by filename
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".gowkhtmltopdf") // name of config file (without extension)
	viper.AddConfigPath("$HOME")          // adding home directory as first search path
	viper.AutomaticEnv()                  // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
