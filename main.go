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

type task struct{
	// To convert into json, we need to make all field exportable, ie Capital
	ID int64
	Title string
	Description string
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
func (p *task) DescriptionFunc() string {
	return p.Description
}

type tasks struct{
	Completed []task
	Incomplete []task
}
var lastInsertedID int64

// Taking Existing data from database

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
	var id int64
	var title string
	var description string
	var status int 

	var toDo tasks

	rows, err := db.Query("SELECT * FROM list")
	checkErr(err)
	// only run when the tasks list is empty, tasks are global variables so refreshing the page do not empty the list
	for rows.Next(){
		err = rows.Scan(&id, &title, &description, &status)
		checkErr(err)
		if status == 1 {
			toDo.Completed = append(toDo.Completed, task{ID:id, Title: title, Description: description, Status:status})
		} else{
			toDo.Incomplete = append(toDo.Incomplete, task{ID:id, Title: title, Description: description, Status:status})
		}
	}
	if request.Method == "GET" {
		// Pass existing data to the index page
		err = indexTemplate.Execute(response, toDo)
		checkErr(err)
	} else {
		// else it is POST method
		request.ParseForm()
		title = request.Form.Get("title")
		description = request.Form.Get("description")
		deleteID := request.Form.Get("deleteID")
		completedID := request.Form.Get("completedID")
		editID := request.Form.Get("editID")
		uId := request.Form.Get("uId")
		uTitle := request.Form.Get("uTitle")
		uDescription := request.Form.Get("uDescription")

		fmt.Println(completedID)
		fmt.Println("d.ID: ",deleteID)
		// r.Form gives a map of values, get gives single value
		// add title to db if not empty

		// Insert to database
		if len(title) !=0 {
			// INSERT/
			stmt, err := db.Prepare("INSERT list SET title=?, description=?, status=?")
			checkErr(err)
			// Status is 0 for new item. ie incomplete
			res, _ := stmt.Exec(title, description, 0)
			id, _ := res.LastInsertId()
			idStr := strconv.FormatInt(id,10)
			m := map[string]string{"id":idStr, "description": description}
			b, _ := json.Marshal(m)
			response.Write(b)
			fmt.Println("ID: ",id,"-",title,"-", description)
			// Add new data to incompleteTasks slice 
			toDo.Incomplete = append(toDo.Incomplete, task{ID:id, Title: title, 
				Description: description, Status: status})
	
		}
		// Delete from Database
		if deleteID != "" {
			id, err = strconv.ParseInt(deleteID,10,64)
			checkErr(err)
			stmt, err := db.Prepare("DELETE from list WHERE id=?") 
			checkErr(err) 
			_, err = stmt.Exec(id)
			checkErr(err)
			// To delete from the slices, 1st we have to find in which slice the item is present
			itemfound := false
			for index, item := range toDo.Incomplete{
				if item.ID == id {
					toDo.Incomplete = append(toDo.Incomplete[:index], toDo.Incomplete[index+1:]...)
					itemfound = true
					break
				}
			}
			if ! itemfound {
				for index, item := range toDo.Completed{
					if item.ID == id {
						toDo.Completed = append(toDo.Completed[:index], toDo.Completed[index+1:]...)
						itemfound = false 
						break
					}
				}
	
			}
		}
		// if status changed to completed, then
		// move task form incompleteTasks to completedTasks 
		if completedID != "" {
			id, err = strconv.ParseInt(completedID,10,64)
			checkErr(err)
			stmt, err := db.Prepare("UPDATE list SET status=? WHERE id=?")
			checkErr(err) 
			_, err = stmt.Exec(1, id)
			checkErr(err)
			// Delete from the incompleteTasks
			// when an existing item is completed then only completedID will post, title will be empty 
			for index, item := range toDo.Incomplete{
				if item.ID == id {
					title = item.Title
					toDo.Incomplete = append(toDo.Incomplete[:index], toDo.Incomplete[index+1:]...)
					break
				}
			}
			// Add to completedTasks
			toDo.Completed = append(toDo.Completed, task{ID:id, Title: title, Status: status})
		}
		if editID != ""{
			id, err = strconv.ParseInt(editID,10,64)
			checkErr(err)
			// Cancel button is only for incomplete tasks. search for id and return a map of title and description.
			for _, item := range toDo.Incomplete{
				if item.ID == id {
					title = item.Title
					description = item.Description
					break
				}
			}
			m := map[string]string{"title":title, "description":description}
			b, _ := json.Marshal(m)
			response.Write(b)
		}

		// Update data
		if uId !="" {
			id, err = strconv.ParseInt(uId,10,64)
			checkErr(err)
			stmt, err := db.Prepare("UPDATE list SET title=?, description=? WHERE id=?")
			checkErr(err) 
			_, err = stmt.Exec(uTitle, uDescription, id)
			checkErr(err)
			// Update from the incompleteTasks
			// when an existing item is completed then only completedID will post, title will be empty 
			for _, item := range toDo.Incomplete{
				if item.ID == id {
					item.Title = uTitle
					item.Description = uDescription
					break
				}
			}
		}
	}
}

/*
func getCompletedTasks(w http.ResponseWriter, r *http.Request) {
	// Convert list of data into json
	b, err := json.Marshal()
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

}
				*/

func main(){
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", index)
	// http.HandleFunc("/getCompletedTasks", getCompletedTasks)

	err := http.ListenAndServe(":9090", nil)
	checkErr(err)
}
