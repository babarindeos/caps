package services

import (
	"fmt"
	"net/http"
)

// contract structure object------------
type contract struct {
	name string
}

func newContractHandler(w http.ResponseWriter, r *http.Request) {
	var recordCount int

	if r.Method != http.MethodPost {

		tmpl.ExecuteTemplate(w, "new_contract.html", nil)
		return
	}

	db = getMySQLDB()
	contractObj := contract{name: r.FormValue("Name")}
	err := db.QueryRow("Select Count(*) from contracts where name=?", contractObj.name).Scan(&recordCount)

	//------------ Check if a contract of same name already exist
	fmt.Println(recordCount)

	if err != nil {
		tmpl.ExecuteTemplate(w, "new_contract.html", struct {
			Success bool
			Status  bool
			Message string
		}{Success: true, Status: false, Message: err.Error()})
		return
	} else {

		if recordCount > 0 {
			tmpl.ExecuteTemplate(w, "new_contract.html", struct {
				Success bool
				Status  bool
				Message string
			}{Success: true, Status: false, Message: "A Contract with that same name already exist!"})
			return
		}

	}

	//------------ Create new contract
	_, err = db.Exec("Insert into contracts(name) values (?)", contractObj.name)

	if err != nil {
		tmpl.ExecuteTemplate(w, "new_contract.html", struct {
			Success bool
			Status  bool
			Message string
		}{Success: true, Status: false, Message: err.Error()})
	} else {
		tmpl.ExecuteTemplate(w, "new_contract.html", struct {
			Success bool
			Status  bool
			Message string
		}{Success: true, Status: true, Message: "The Contract has been successfully created"})
	}

	//fmt.Fprintf(w, "Welcome to golang by url %s", r.FormValue("Name"))
	//fmt.Println(r.FormValue("Name"))

}
