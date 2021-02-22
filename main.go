package main

import (
	"fmt"
	"log"
	"net/http"
	"html/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		title := r.Form.Get("title")
		// r.Form gives a map of values, get gives single value
		// add title to db if not empty
		if len(title) !=0 {
			fmt.Println(title, "To Database")
			db, _ := sql.Open("mysql", "fahim:12345@tcp(localhost:3306)/to_do_list")
			// INSERT
			stmt, _ := db.Prepare("INSERT list SET title=?")
			res, _ := stmt.Exec(title)
			id, _ := res.LastInsertId()
			fmt.Println("ID: ",id,"-",title)
					t, _ := template.ParseFiles("index.html")
		t.Execute(w, nil)
		db.QueryRow("SELECT * FROM list")
		}
	}
}
func main(){
	http.HandleFunc("/", index)
	fmt.Println("Server running...")
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}