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
// completedTasks - data from the database in an slice form
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

// completedTasks - data from the database in an slice form
type task struct{
	// To convert into json, we need to make all field exportable, ie Capital
	Title string
	ID int64
	Status int 
}
func (p *task) TitleFunc() string {
	return p.Title
}
func (p *task) IdFunc() int64 {
	return p.ID
}
func (p *task) StatusFunc() int {
	return p.Status
}
var lastInsertedID int64

// Taking Existing data from database
var completedTasks []task 
var incompleteTasks []task 

func index(response http.ResponseWriter, request *http.Request) {
	// Connect to database
	db, err := sql.Open("mysql", "fahim:12345@tcp(localhost:3306)/to_do_list")
	checkErr(err)
	defer db.Close()

	// Parse index.html template
	indexTemplate := template.Must(template.ParseFiles("static/index.html"))
	// if didn't enclose the ParseFiles method inside the Must method, we had to handle error separately
	// when using template as string, instead of ParseFiles, we can use below code.
	// templateStr := template.Must(template.New("anyname").Parse(templateSring))

	// declaring for scaning from database
	var title string
	var status int 
	var id int64

	rows, err := db.Query("SELECT * FROM list")
	checkErr(err)
	// only run when the tasks list is empty, tasks are global variables so refreshing the page do not empty the list
	if len(completedTasks) == 0 && len(incompleteTasks) == 0 {
		for rows.Next(){
			err = rows.Scan(&id, &title, &status)
			checkErr(err)
			if status == 1 {
				completedTasks = append(completedTasks, task{ID:id, Title: title, Status:status})
			} else{
				incompleteTasks = append(incompleteTasks, task{ID:id, Title: title, Status:status})
			}
		}
	}
	if request.Method == "GET" {
		// Pass existing data to the index page
		err = indexTemplate.Execute(response, incompleteTasks)
		checkErr(err)
	} else {
		// else it is POST method
		request.ParseForm()
		// If it insert
		title = request.Form.Get("title")
		deleteID := request.Form.Get("deleteID")
		completedID := request.Form.Get("completedID")
		fmt.Println(completedID)
		fmt.Println("d.ID: ",deleteID)
		// r.Form gives a map of values, get gives single value
		// add title to db if not empty
		if len(title) !=0 {
			// INSERT/
			stmt, err := db.Prepare("INSERT list SET title=?, status=?")
			checkErr(err)
			res, _ := stmt.Exec(title, 0)
			id, _ := res.LastInsertId()
			b, _ := json.Marshal(id)
			response.Write(b)
			fmt.Println("ID: ",id,"-",title)
			// Add new data to incompleteTasks slice 
			incompleteTasks = append(incompleteTasks, task{ID:id, Title: title, Status: status})
	
		}
		// Delete from Database
			if deleteID != "" {
			id, err = strconv.ParseInt(deleteID,10,64)
			checkErr(err)
			stmt, err := db.Prepare("DELETE from list WHERE id=?") 
			checkErr(err) 
			_, err = stmt.Exec(id)
			checkErr(err)
			// Delete from the incomplete tasks
			for index, item := range incompleteTasks{
				if item.ID == id {
					incompleteTasks = append(incompleteTasks[:index], incompleteTasks[index+1:]...)
					break
				}
			}
	
		}
		// change status for completed tasks
		// move form incompleteTasks to completedTasks 
			if completedID != "" {
			id, err = strconv.ParseInt(completedID,10,64)
			checkErr(err)
			stmt, err := db.Prepare("UPDATE list SET status=? WHERE id=?")
			checkErr(err) 
			_, err = stmt.Exec(1, id)
			checkErr(err)
			// Delete from the incompleteTasks
			// when an existing item is completed then only completedID will post, title will be empty 
			for index, item := range incompleteTasks{
				if item.ID == id {
					title = item.Title
					incompleteTasks = append(incompleteTasks[:index], incompleteTasks[index+1:]...)
					break
				}
			}
			// Add to completedTasks
			completedTasks = append(completedTasks, task{ID:id, Title: title, Status: status})
		}
	}
}

func getCompletedTasks(w http.ResponseWriter, r *http.Request) {
	// Convert list of data into json
	b, err := json.Marshal(completedTasks)
	checkErr(err)
	fmt.Println("clicked")
	fmt.Println(string(b))
	// instead of above code we can also use
	// json.NewEncoder(w).Encode(lastInsertedID)
	w.Write(b)
	/*
    var html := `<li>
                <div class="row me-4 p-1">
                    <div class="col-8" id="title-col">
                        ${task.ID} ${task.Title}
                    </div>
                    <div class="col-4 btn-group" >
                        <button class="btn" id="delete-button" name="deleteID" value="${task.ID}">
                            <i class="fas fa-trash"></i>
                        </button>
                        <button class="btn" id="status-button" name="completedID" value="${task.ID}">
                            <i class="far fa-check-circle"></i>
                        </button>
                        <button class="btn" id="delete-button" name="deleteID" value="${task.ID}">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                    <hr>
                </div>
            	</li> `
				*/

}

func main(){
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", index)
	http.HandleFunc("/getCompletedTasks", getCompletedTasks)

	err := http.ListenAndServe(":9000", nil)
	checkErr(err)
}
