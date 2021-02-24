package main

import (
	"database/sql"
	"fmt"
	"text/template"
	"strconv"
	// html/template is mainly used for cross site scripting, to escape html syntax
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
} 

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
		//li += fmt.Sprintf(`<li>%v</li>`, title)
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
		// r.Form gives a map of values, get gives single value
		// add title to db if not empty
		if len(title) !=0 {
			// INSERT
			stmt, err := db.Prepare("INSERT list SET title=?")
			checkErr(err)
			res, _ := stmt.Exec(title)
			id, _ := res.LastInsertId()
			fmt.Println("ID: ",id,"-",title)
			// Add new data to tasks slice 
			tasks = append(tasks, task{id:id, title: title})
		}

		// If it delete 
		deleteID := request.Form.Get("deleteID")
		if deleteID != "" {
			id, err = strconv.ParseInt(deleteID,10,64)
			checkErr(err)
			stmt, err := db.Prepare("DELETE from list WHERE id=?") 
			checkErr(err) 
			_, err = stmt.Exec(id)
			checkErr(err)
			for index, item := range tasks{
				if item.id == id {
					tasks = append(tasks[:index], tasks[index+1:]...)
					break
				}
			}
		}

		err = indexTemplate.Execute(response, tasks)
		checkErr(err)
	}
}
func delete(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	fmt.Println(request.Form.Get("ID"))
	fmt.Fprintf(response, "hi")
}

func main(){
	// To add external files
	/*
	Added "/assets/" in both here and inside the template style path also, and strip it out.
	or we have to add "/" instead "/assets/" in Handle   to avoid using http.StripPrefix function,
	but when we do so it throws a panic error which says that "/" with different handler, it overcome this we have to change "/" to "/index" in oru handleFunc. 
	*/
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", index)
	http.HandleFunc("/delete", delete)

	err := http.ListenAndServe(":9000", nil)
	checkErr(err)
}