package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	. "github.com/dongweiming/eshop-prices/models"

)

var rootCmd = &cobra.Command{
	Use:   "eshop-prices",
	Short: "Manage Data",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Initialize the database. Test environment or use it for the first time,
execute it will delete data, please note`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Init Finished!!")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "database migration",
	Long: `Database migration is required whenever database fields are changed!`,
	Run: func(cmd *cobra.Command, args []string) {
		DB.AutoMigrate(&Game{}, &Publisher{}, &Developer{}, &Genre{}, &GamePublisher{},
			&GameDeveloper{}, &GameGenre{})
		log.Info("Database Migrated!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(migrateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
