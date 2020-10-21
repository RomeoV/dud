package cmd

import (
	"log"
	"os"

	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the given Stage and all upstream Stages.",
	Run: func(cmd *cobra.Command, args []string) {

		ch, err := cache.NewLocalCache(viper.GetString("cache"))
		if err != nil {
			log.Fatal(err)
		}

		idx, err := index.FromFile(".dud/index")
		if os.IsNotExist(err) {
			idx = make(index.Index)
		} else if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			for path := range idx {
				args = append(args, path)
			}
		}

		ran := make(map[string]bool)
		for _, path := range args {
			inProgress := make(map[string]bool)
			err := idx.Run(path, ch, ran, inProgress)
			if err != nil {
				log.Fatal(err)
			}
		}

		for path, wasRun := range ran {
			description := "not run"
			if wasRun {
				description = "run"
			}
			log.Printf("%s: %s\n", path, description)
		}
	},
}
