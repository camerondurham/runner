/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	LANG_ENDPOINT = "/api/v1/languages"
)

var (
	server string = "http://localhost:8080"
)

// langsCmd represents the langs command
var langsCmd = &cobra.Command{
	Use:   "langs",
	Short: "query the server for supported languages",
	Run: func(cmd *cobra.Command, args []string) {
		// implement CLI subcommand logic here
		resp, err := getLangListJSON(server, LANG_ENDPOINT)
		if err.Error() != "" {
			fmt.Println(err)
			return
		}
		fmt.Print("Currently supported languages: ")
		for _, lang := range resp.Languages {
			fmt.Printf("%s ", lang)
		}
	},
}

func init() {
	rootCmd.AddCommand(langsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// langsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// langsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}