package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", 8080)

	viper.SetDefault("mysql.user", "root")
	viper.SetDefault("mysql.password", "")
	viper.SetDefault("mysql.host", "localhost")
	viper.SetDefault("mysql.port", 3306)
	viper.SetDefault("mysql.protocol", "tcp")
	viper.SetDefault("mysql.db", "")

	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@%s(%s:%d)/%s",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.protocol"),
		viper.GetString("mysql.host"),
		viper.GetInt("mysql.port"),
		viper.GetString("mysql.db"),
	))
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	findAccountByUuidStmt, err := db.Prepare("SELECT username FROM accounts WHERE uuid = ? LIMIT 1")
	if err != nil {
		panic(err)
	}

	router := httprouter.New()
	router.GET("/api/minecraft/session/profile/:uuid", logRequestHandler(func(
		response http.ResponseWriter,
		request *http.Request,
		params httprouter.Params,
	) {
		uuid, err := formatUuid(params.ByName("uuid"))
		if err != nil {
			response.WriteHeader(204)
			return
		}

		var username string
		err = findAccountByUuidStmt.QueryRow(uuid).Scan(&username)
		if errors.Is(err, sql.ErrNoRows) {
			response.WriteHeader(204)
			return
		} else if err != nil {
			panic(err)
		}

		profileUrl := "http://skinsystem.ely.by/profile/" + username
		if request.FormValue("unsigned") == "false" {
			profileUrl += "?unsigned=false"
		}

		profileResp, err := http.Get(profileUrl)
		if err != nil {
			// TODO: log reason
			response.WriteHeader(500)
			return
		}

		response.Header().Set("Content-Type", "application/json")

		_, err = io.Copy(response, profileResp.Body)
		if err != nil {
			response.WriteHeader(500)
			return
		}
	}))

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", viper.GetString("http.host"), viper.GetInt("http.port")), router)
	if err != nil {
		panic(err)
	}
}

func formatUuid(input string) (string, error) {
	uuid := strings.ReplaceAll(input, "-", "")
	if len(uuid) != 32 {
		return "", errors.New("invalid uuid")
	}

	return uuid[0:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:], nil
}

func logRequestHandler(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		lrw := negroni.NewResponseWriter(w)

		h(lrw, r, p) // Call the original handler
		log.Printf("\"%s %s\" %d", r.Method, r.URL.String(), lrw.Status())
	}
}
