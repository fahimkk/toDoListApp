package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	//"go/printer"
	"strconv"
	"text/template"

	//"strconv"
	// html/template is mainly used for cross site scripting, to escape html syntax
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
} 
/*
// tasks - data from the database in an slice form
type task struct{
	title string
	id int64
}
func (p *task) Title() string {
	return p.title
}
func (p *task) ID() int64 {
	return p.id
}
*/

// tasks - data from the database in an slice form
type task struct{
	title string
	id int64
}
func (p *task) Title() string {
	return p.title
}
func (p *task) ID() int64 {
	return p.id
}

var lastInsertedID int64

func index(response http.ResponseWriter, request *http.Request) {
	// Connect to database
	db, err := sql.Open("mysql", "fahim:12345@tcp(localhost:3306)/to_do_list")
	checkErr(err)
	// Parse index.html template
	indexTemplate := template.Must(template.ParseFiles("static/index.html"))
	// if didn't enclose the ParseFiles method inside the Must method, we had to handle error separately
	// when using template as string, instead of ParseFiles, we can use below code.
	// templateStr := template.Must(template.New("anyname").Parse(templateSring))

	var title string
	var id int64

	// Taking Existing data from database
	var tasks []task 
	rows, err := db.Query("SELECT * FROM list")
	checkErr(err)
	// var li string
	for rows.Next(){
		err = rows.Scan(&id, &title)
		checkErr(err)
		tasks = append(tasks, task{id:id, title: title})
	}
	if request.Method == "GET" {
		// Pass existing data to the index page
		err = indexTemplate.Execute(response, tasks)
		checkErr(err)
	} else {
		// else it is POST method
		request.ParseForm()
		// If it insert
		title = request.Form.Get("title")
		deleteID := request.Form.Get("deleteID")
		fmt.Println("d.ID: ",deleteID)
		// r.Form gives a map of values, get gives single value
		// add title to db if not empty
		if len(title) !=0 {
			// INSERT
			stmt, err := db.Prepare("INSERT list SET title=?")
			checkErr(err)
			res, _ := stmt.Exec(title)
			id, _ := res.LastInsertId()
			b, _ := json.Marshal(id)
			response.Write(b)
			fmt.Println("ID: ",id,"-",title)
		}
	}
}


func delete(w http.ResponseWriter, r *http.Request) {
	// Delete
	r.ParseForm()
	deleteID := r.Form.Get("deleteID")
		if deleteID != "" {
			id, err := strconv.ParseInt(deleteID,10,64)
			checkErr(err)
			fmt.Println(id)
		}
		http.Redirect(w, r, "/",http.StatusSeeOther)

	// instead of above code we can also use
	// json.NewEncoder(w).Encode(lastInsertedID)
}

func main(){
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", index)
	http.HandleFunc("/delete", delete)

	err := http.ListenAndServe(":9000", nil)
	checkErr(err)
}