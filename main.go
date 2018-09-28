package main

// Copyright (C) 2018 Joas Schilling <coding@schilljs.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

var (
	config            *viper.Viper
	sqlInsertPresence *sql.Stmt
)

func main() {
	arguments := os.Args[1:]
	if len(arguments) != 1 || (arguments[0] != "test" && arguments[0] != "fetch") {
		log.Println("./my-cfhn-presence-stats [argument]")
		log.Println(" * \"test\" the config, database connection and connection to the presence API")
		log.Println(" * \"fetch\" query the presence API and store the data in the database")
		return
	}

	config = viper.New()
	config.SetConfigName("config")
	config.AddConfigPath(".")
	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("[✘ ] Fatal error config file: %s \n", err)
		return
	}
	log.Println("[✓ ] Config file loaded")

	// Open database connection
	var err error
	db, err := sql.Open("mysql",
		config.GetString("database.user")+":"+config.GetString("database.password")+
			"@tcp("+config.GetString("database.host")+")/"+config.GetString("database.name"))
	if err != nil {
		log.Fatalf("[✘ ] Fatal error database connection: %s \n", err)
		return
	}
	log.Println("[✓ ] Database connection established")
	defer db.Close()

	sqlInsertPresence, err = db.Prepare("INSERT INTO `presences` (`username`, `datetime`) VALUES (?,?)")
	if err != nil {
		log.Fatalf("[✘ ] Fatal error database could not prepare insert: %s \n", err)
		return
	}

	if arguments[0] == "test" {
		client := &http.Client{}
		_, err = client.Get(config.GetString("presence_server"))
		if err != nil {
			log.Fatalf("[✘ ] Could not connect to resence API: %s", err)
			return
		}
		log.Println("[✓ ] Presence API is reachable")
	} else if arguments[0] == "fetch" {
		fetchPresencesFromAPI()
	}
}