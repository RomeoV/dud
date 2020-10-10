package cmd

import (
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "duc",
		Short: "DUC is a tool for storing, versioning, and reproducing big data files",
		Long: `Data Under Control (duc) is a tool to store, version, and reproduce big
		data files alongside the source code that creates it.
		Inspired by Git.`,
	}
)

// Execute is the main entry point to the cobra cli
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	rootDir, err := getProjectRootDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(rootDir, ".duc"))
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	defaultCachePath, err := filepath.Abs(".duc/cache")
	if err != nil {
		panic(err)
	}
	viper.SetDefault("cache", defaultCachePath)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}
	}
}