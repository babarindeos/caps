package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var tmpl *template.Template
var recordCount int

var tmplDashboard *template.Template
var tmplContractBid *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
	tmplDashboard = template.Must(template.ParseFiles("templates/dashboard.html", "templates/header.html", "templates/footer.html"))
	tmplContractBid = template.Must(template.ParseFiles("templates/contract_bidding.html", "templates/header.html", "templates/footer.html"))
}

func getMySQLDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@(127.0.0.1)/caps?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// contract structure object------------
type Contract struct {
	id   string
	name string
}

// end of contract structure object ------

// --- Bid struct ------------------------
type Bid struct {
	id                             string
	contract_id                    string
	bidder_name                    string
	bank_statement_submission      string
	bank_statement_score           string
	bank_reference_submission      string
	bank_reference_score           string
	cv_submission                  string
	cv_score                       string
	certificate_submission         string
	certificate_score              string
	awardletter_submission         string
	awardletter_score              string
	completionletter_submission    string
	completionletter_score         string
	equipment_submission           string
	equipment_score                string
	equipment_ownership_submission string
	equipment_ownership_score      string
	deviation_submission           string
	deviation_score                string
	projects_submission            string
	projects_score                 string
}

//---- End of Bid struct -----------------

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	db := getMySQLDB()
	defer db.Close() // Ensure the database connection is closed when the function completes

	rows, err := db.Query("SELECT id, name FROM contracts order by id desc")
	if err != nil {
		tmpl.ExecuteTemplate(w, "dashboard.html", struct {
			Success bool
			Message template.HTML
		}{Success: false, Message: template.HTML("An error occurred loading the records")})
		return
	}
	defer rows.Close() // Ensure rows are closed after processing

	var contractArray []string
	contractArray = append(contractArray, "<table border='1' cellpadding='2' style='border-collapse:collapse; border:1px solid blue;'>")
	contractArray = append(contractArray, "<tr class='bg-blue-600 text-white' style='font-weight:bold'><td style='text-align:center'>ID</td><td>Name</td><td></td></tr>")

	for rows.Next() {
		var c Contract
		if err := rows.Scan(&c.id, &c.name); err != nil {
			log.Fatal(err) // Consider logging the error rather than killing the application
		}
		contractArray = append(contractArray, fmt.Sprintf("<tr><td style='text-align:center;'>%s.</td><td>%s</td><td style='text-align:center'>%s</td></tr>", c.id, c.name, "<a style='text-decoration:underline;' class='text-blue-500' href='/contract_home?id="+c.id+"'>Open</a>"))
	}
	contractArray = append(contractArray, "</table>")

	htmlOutput := template.HTML(strings.Join(contractArray, "")) // Convert slice to string and then to template.HTML

	tmpl.ExecuteTemplate(w, "dashboard.html", struct {
		Success bool
		Message template.HTML
	}{Success: true, Message: htmlOutput})
}

func newContractHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

		tmpl.ExecuteTemplate(w, "new_contract.html", nil)
		return
	}

	db = getMySQLDB()
	contractObj := Contract{name: r.FormValue("Name")}
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

func contractBiddingHandler(w http.ResponseWriter, r *http.Request) {

	contractId := r.URL.Query().Get("id")

	if r.Method != http.MethodPost {

		if contractId == "" {
			// contractId is not set or is empty
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		db = getMySQLDB()

		// Query the database for the contract by ID
		rows, err := db.Query("SELECT id, name FROM contracts WHERE id=?", contractId)
		if err != nil {
			tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
				Success bool
				Status  bool
				Title   string
				Message template.HTML
			}{Success: false, Status: false, Title: "Contract Bidding", Message: template.HTML("An error occurred loading the record")})
			return
		}
		defer rows.Close()

		// Check if there is a result
		if !rows.Next() {
			http.NotFound(w, r)
			return
		}

		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning rows: %v", err)
			http.Error(w, "Error reading data", http.StatusInternalServerError)
			return
		}

		// Render the contract data using a template
		tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
			Success bool
			Status  bool
			Title   string
			Message string
		}{Success: false, Status: true, Title: name, Message: contractId})

		return

	}

	// if isPost implementation
	bid := Bid{
		contract_id:                    r.FormValue("contract_id"),
		bidder_name:                    r.FormValue("bidder_name"),
		bank_statement_submission:      r.FormValue("bank_statement_submission"),
		bank_statement_score:           r.FormValue("bank_statement_score"),
		bank_reference_submission:      r.FormValue("bank_reference_submission"),
		cv_submission:                  r.FormValue("cv_submission"),
		cv_score:                       r.FormValue("cv_score"),
		certificate_submission:         r.FormValue("certificate_submission"),
		certificate_score:              r.FormValue("certificate_score"),
		awardletter_submission:         r.FormValue("awardletter_submission"),
		awardletter_score:              r.FormValue("awardletter_score"),
		completionletter_submission:    r.FormValue("completionletter_submission"),
		completionletter_score:         r.FormValue("completionletter_score"),
		equipment_submission:           r.FormValue("equipment_submission"),
		equipment_score:                r.FormValue("equipment_score"),
		equipment_ownership_submission: r.FormValue("equipment_ownership_submission"),
		equipment_ownership_score:      r.FormValue("equipment_ownership_score"),
		deviation_submission:           r.FormValue("deviation_submission"),
		deviation_score:                r.FormValue("deviation_score"),
		projects_submission:            r.FormValue("project_submission"),
		projects_score:                 r.FormValue("projects_score"),
	}

	db = getMySQLDB()
	contract_id, _ := strconv.Atoi(bid.contract_id)
	bank_statement_submission, _ := strconv.Atoi(bid.bank_statement_submission)
	bank_statement_score, _ := strconv.Atoi(bid.bank_statement_score)
	bank_reference_submission, _ := strconv.Atoi(bid.bank_reference_submission)
	bank_reference_score, _ := strconv.Atoi(bid.bank_reference_score)
	cv_submission, _ := strconv.Atoi(bid.cv_submission)
	cv_score, _ := strconv.Atoi(bid.cv_score)
	certificate_submission, _ := strconv.Atoi(bid.certificate_submission)
	certificate_score, _ := strconv.Atoi(bid.certificate_score)
	awardletter_submission, _ := strconv.Atoi(bid.awardletter_submission)
	awardletter_score, _ := strconv.Atoi(bid.awardletter_score)
	completionletter_submission, _ := strconv.Atoi(bid.completionletter_submission)
	completionletter_score, _ := strconv.Atoi(bid.completionletter_score)
	equipment_submission, _ := strconv.Atoi(bid.equipment_submission)
	equipment_score, _ := strconv.Atoi(bid.equipment_score)
	equipment_ownership_submission, _ := strconv.Atoi(bid.equipment_ownership_submission)
	equipment_ownership_score, _ := strconv.Atoi(bid.equipment_ownership_score)
	deviation_submission, _ := strconv.Atoi(bid.deviation_submission)
	deviation_score, _ := strconv.Atoi(bid.deviation_score)
	projects_submission, _ := strconv.Atoi(bid.projects_submission)
	projects_score, _ := strconv.Atoi(bid.projects_score)

	//_, err := db.Exec("Insert into bids(contract_id) values (?)", contract_id)

	_, err := db.Exec("Insert into bids(contract_id, bidder_name, bank_statement_submission, bank_statement_score, bank_reference_submission, bank_reference_score, cv_submission, cv_score, certificate_submission, certificate_score, awardletter_submission, awardletter_score, completionletter_submission, completionletter_score, equipment_submission, equipment_score, equipment_ownership_submission, equipment_ownership_score, deviation_submission, deviation_score, projects_submission, projects_score) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		contract_id, bid.bidder_name, bank_statement_submission, bank_statement_score, bank_reference_submission, bank_reference_score, cv_submission, cv_score, certificate_submission, certificate_score, awardletter_submission, awardletter_score, completionletter_submission, completionletter_score, equipment_submission, equipment_score, equipment_ownership_submission, equipment_ownership_score, deviation_submission, deviation_score, projects_submission, projects_score)

	if err != nil {
		tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
			Success bool
			Status  bool
			Title   string
			Message string
		}{Success: true, Status: false, Title: "Contract", Message: err.Error()})

		return
	} else {
		tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
			Success bool
			Status  bool
			Title   string
			Message string
		}{Success: true, Status: true, Title: "Contract", Message: "The Bid Information has been successfully created and saved"})
		return
	}

	//fmt.Println(r.FormValue("certificate_submission"))
}

func contractHomeHandler(w http.ResponseWriter, r *http.Request) {
	contractId := r.URL.Query().Get("id")

	if contractId == "" {
		// contractId is not set or is empty
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	db = getMySQLDB()

	// Query the database for the contract by ID
	rows, err := db.Query("SELECT id, name FROM contracts WHERE id=?", contractId)
	if err != nil {
		tmpl.ExecuteTemplate(w, "contract_home.html", struct {
			Success bool
			Title   string
			Id      string
			Message template.HTML
		}{Success: false, Message: template.HTML("An error occurred loading the record")})
		return
	}
	defer rows.Close()

	// Check if there is a result
	if !rows.Next() {
		http.NotFound(w, r)
		return
	}

	var id, name string
	if err := rows.Scan(&id, &name); err != nil {
		log.Printf("Error scanning rows: %v", err)
		http.Error(w, "Error reading data", http.StatusInternalServerError)
		return
	}

	// -------------   Get Bid Information --------------------
	rows, err = db.Query("SELECT id, contract_id, bidder_name, bank_statement_submission, bank_statement_score, bank_reference_submission, "+
		"bank_reference_score, cv_submission, cv_score, certificate_submission, certificate_score FROM bids where contract_id=? order by id desc", id)
	if err != nil {
		tmpl.ExecuteTemplate(w, "contract_home.html", struct {
			Success bool
			Title   string
			Id      string
			Message template.HTML
		}{Success: true, Title: name, Id: id, Message: template.HTML("An error occurred loading the records")})
		return
	}
	defer rows.Close() // Ensure rows are closed after processing

	var bidsArray []string

	bidsArray = append(bidsArray, "<table border='1' cellpadding='2' style='border-collapse:collapse; border:1px solid blue;'>")
	bidsArray = append(bidsArray, "<tr class='bg-blue-600 text-white' style='font-weight:bold'><td style='text-align:center'>ID</td><td>Bidder</td>"+
		"<td>Bank<br/>Statement</td><td>Bank<br/>Reference</td><td>CV</td><td>Certificate</td>"+
		"<td>Action</td></tr>")

	for rows.Next() {
		var b Bid
		if err := rows.Scan(&b.id, &b.contract_id, &b.bidder_name, &b.bank_statement_submission,
			&b.bank_statement_score, &b.bank_reference_submission, &b.bank_reference_score, &b.cv_submission, &b.cv_score, &b.certificate_submission,
			&b.certificate_score); err != nil {
			log.Fatal(err) // Consider logging the error rather than killing the application
		}
		bidsArray = append(bidsArray, fmt.Sprintf("<tr><td style='text-align:center;'>%s.</td><td>%s</td><td>%s / %s</td><td>%s / %s</td><td>%s / %s</td><td>%s / %s</td><td style='text-align:center'>%s</td></tr>",
			b.id, b.bidder_name, b.bank_statement_submission, b.bank_statement_score, b.bank_reference_submission, b.bank_reference_score, b.cv_submission, b.cv_score, b.certificate_submission, b.certificate_score,
			"<a style='text-decoration:underline;' class='text-blue-500' href='/contract_home?id="+b.id+"'>Open</a>"))
	}
	bidsArray = append(bidsArray, "</table>")

	htmlOutput := template.HTML(strings.Join(bidsArray, "")) // Convert slice to string and then to template.HTML

	// Render the contract data using a template
	tmpl.ExecuteTemplate(w, "contract_home.html", struct {
		Success bool
		Title   string
		Id      string
		Message template.HTML
	}{Success: true, Title: name, Id: id, Message: htmlOutput})

}

func main() {
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/new_contract", newContractHandler)
	http.HandleFunc("/contract_home", contractHomeHandler)
	http.HandleFunc("/contract_bidding", contractBiddingHandler)
	http.ListenAndServe(":999", nil)

}
