package main

import (
	"fmt"
	"net/http"
	"html/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
} 

func index(response http.ResponseWriter, request *http.Request) {
	indexTemplate, err := template.ParseFiles("index.html")
	checkErr(err)
	if request.Method == "GET" {
		err = indexTemplate.Execute(response, nil)
		checkErr(err)
	} else {
		// else it is POST method
		request.ParseForm()
		title := request.Form.Get("title")
		// r.Form gives a map of values, get gives single value
		// add title to db if not empty
		if len(title) !=0 {
			db, err := sql.Open("mysql", "fahim:12345@tcp(localhost:3306)/to_do_list")
			checkErr(err)
			// INSERT
			stmt, err := db.Prepare("INSERT list SET title=?")
			checkErr(err)
			res, _ := stmt.Exec(title)
			id, _ := res.LastInsertId()
			fmt.Println("ID: ",id,"-",title)
			indexTemplate.Execute(response, nil)
			stmt, err = db.Prepare("DELETE from list where id=?")
			checkErr(err)
			var i int64
			for i = 0; i < 10; i++ {
				stmt.Exec(id-i)
			}
			// db.QueryRow("SELECT * FROM list")
		}
	}
}
func main(){
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":9000", nil)
	checkErr(err)
}