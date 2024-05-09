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
	id               string
	contractId       string
	bidderName       string
	bankStatement    string
	bankReference    string
	cv               string
	certificate      string
	awardletter      string
	completionletter string
	equipment        string
	ownership        string
	auditedAccount   string
	projects         string
}

//---- End of Bid struct -----------------

type BidSectionScore struct {
	id                   string
	contract_id          string
	bidder_name          string
	financial_capability string
	qualification        string
	similar_projects     string
	technical_capability string
	audited_account      string
	projects             string
	total_score          string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	db = getMySQLDB()
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
	contractArray = append(contractArray, "<tr class='bg-blue-600 text-white' style='font-weight:bold'><td style='text-align:center; width:5%'>ID</td><td style='width:75%;'>Name</td><td class='text-center'>Actions</td></tr>")

	for rows.Next() {
		var c Contract
		if err := rows.Scan(&c.id, &c.name); err != nil {
			log.Fatal(err) // Consider logging the error rather than killing the application
		}
		contractArray = append(contractArray, fmt.Sprintf("<tr><td style='text-align:center;'>%s.</td><td>%s</td><td style='text-align:center'>%s</td></tr>", c.id, c.name, "<a style='text-decoration:underline;' class='text-blue-500' href='/contract_home?id="+c.id+"'>Open</a>  |  <a style='text-decoration:underline;' class='text-blue-500' href='#'>Hyperledger</a>"))
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
	var contract_id = r.URL.Query().Get("id")
	if contract_id == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	db = getMySQLDB()
	defer db.Close()

	var contractName string

	err := db.QueryRow("Select name from contracts where id=?", contract_id).Scan(&contractName)

	if err != nil {
		http.Error(w, "Technical error"+err.Error(), http.StatusBadGateway)
		return
	}

	if r.Method != http.MethodPost {
		tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
			Success      bool
			Status       bool
			ContractId   string
			ContractName string
			Message      template.HTML
		}{Success: false, Status: false, ContractId: contract_id, ContractName: contractName, Message: template.HTML("")})
		return
	}

	contract_id = r.URL.Query().Get("id")

	type Bid struct {
		id               string
		contractId       string
		bidderName       string
		bankStatement    string
		bankReference    string
		cv               string
		certificate      string
		awardletter      string
		completionletter string
		equipment        string
		ownership        string
		auditedAccount   string
		projects         string
	}

	var floatBankStatementScore float64
	var floatBankReferenceScore float64
	var floatCVScore float64
	var floatCertificateScore float64
	var floatAwardLetterScore float64
	var floatCompletionLetterScore float64
	var floatEquipmentScore float64
	var floatOwnershipScore float64
	var floatAuditedAccountScore float64
	var floatProjectScore float64

	// floatBankStatementScore, _ = strconv.ParseFloat(r.FormValue("bank_statement"), 64)
	// floatBankReferenceScore, _ = strconv.ParseFloat(r.FormValue("bank_reference"), 64)

	bid := Bid{
		contractId:       contract_id,
		bidderName:       r.FormValue("bidder_name"),
		bankStatement:    r.FormValue("bank_statement"),
		bankReference:    r.FormValue("bank_reference"),
		cv:               r.FormValue("cv"),
		certificate:      r.FormValue("certificate"),
		awardletter:      r.FormValue("award_letter"),
		completionletter: r.FormValue("completion_letter"),
		equipment:        r.FormValue("equipment"),
		ownership:        r.FormValue("ownership"),
		auditedAccount:   r.FormValue("audited_account"),
		projects:         r.FormValue("projects"),
	}

	/* fmt.Println("--------------- Struct Print-------------------")

	fmt.Println("Bid BankStatement: " + bid.bankStatement)
	fmt.Println("Bid bankReference: " + bid.bankReference)
	*/
	floatBankStatement, err := strconv.ParseFloat(bid.bankStatement, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatBankStatementScore = floatBankStatement * 9.0

	//fmt.Println("--------------- Float Bank Statement Computation -------------------")
	//fmt.Println(floatBankStatementScore)

	floatBankReference, err := strconv.ParseFloat(bid.bankReference, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatBankReferenceScore = floatBankReference * 6.0

	floatCV, err := strconv.ParseFloat(bid.cv, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatCVScore = floatCV * 1.0

	floatCertificate, err := strconv.ParseFloat(bid.certificate, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatCertificateScore = floatCertificate * 1.5

	floatAwardLetter, err := strconv.ParseFloat(bid.awardletter, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatAwardLetterScore = floatAwardLetter * 1.0

	floatCompletionLetter, err := strconv.ParseFloat(bid.completionletter, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatCompletionLetterScore = floatCompletionLetter * 1.0

	floatEquipment, err := strconv.ParseFloat(bid.equipment, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatEquipmentScore = floatEquipment * 1.0

	floatOwnership, err := strconv.ParseFloat(bid.ownership, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatOwnershipScore = floatOwnership * 1.5

	floatAuditedAccount, err := strconv.ParseFloat(bid.auditedAccount, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatAuditedAccountScore = floatAuditedAccount * 1.5

	floatProjects, err := strconv.ParseFloat(bid.projects, 64)
	if err != nil {
		fmt.Println(err.Error())
	}
	floatProjectScore = floatProjects * 5

	_, err = db.Exec("Insert into bids(contract_id, bidder_name, bank_statement, bank_statement_score, bank_reference, bank_reference_score, cv, cv_score, certificate, certificate_score, award_letter, award_letter_score, completion_letter, completion_letter_score, equipment, equipment_score, ownership, ownership_score, audited_account, audited_account_score, projects, projects_score) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		bid.contractId, bid.bidderName, bid.bankStatement, floatBankStatementScore, bid.bankReference, floatBankReferenceScore,
		bid.cv, floatCVScore, bid.certificate, floatCertificateScore, bid.awardletter, floatAwardLetterScore,
		bid.completionletter, floatCompletionLetterScore, bid.equipment, floatEquipmentScore,
		bid.ownership, floatOwnershipScore, bid.auditedAccount, floatAuditedAccountScore,
		bid.projects, floatProjectScore)

	if err != nil {
		tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
			Success      bool
			Status       bool
			ContractId   string
			ContractName string
			Message      template.HTML
		}{Success: true, Status: false, ContractId: contract_id, ContractName: contractName, Message: template.HTML(err.Error() + "<div class='py-2 mt-2'><a href='contract_home?id=" + contract_id + "' class='bg-blue-500 hover:bg-blue-400 rounded text-white px-8 py-2 text-sm'>Back to Contract Page</a></div>")})
		return
	}

	tmpl.ExecuteTemplate(w, "contract_bidding.html", struct {
		Success      bool
		Status       bool
		ContractId   string
		ContractName string
		Message      template.HTML
	}{
		Success:      true,
		Status:       true,
		ContractId:   contract_id,
		ContractName: contractName,
		Message:      template.HTML("Bidding has been successfully created. <div class='py-2 mt-2'><a href='contract_home?id=" + contract_id + "' class='bg-blue-500 hover:bg-blue-400 rounded text-white px-8 py-2 text-sm'>Back to Contract Page</a></div>")})

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
	rows, err = db.Query("SELECT sub.id, sub.contract_id, sub.bidder_name,sub.financial_capability, sub.qualification, sub.similar_projects, sub.technical_capability, sub.audited_account, sub.projects, sub.financial_capability + sub.qualification + sub.similar_projects + sub.technical_capability + sub.audited_account + sub.projects AS total_score FROM ( SELECT id, contract_id, bidder_name, bank_statement_score + bank_reference_score AS financial_capability, cv_score + certificate_score AS qualification, award_letter_score + completion_letter_score AS similar_projects, equipment_score + ownership_score AS technical_capability, audited_account_score AS audited_account, projects_score AS projects FROM `bids` ) AS sub ORDER BY total_score DESC;")
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
		"<td>Financial<br>Capability</td><td>Qualifications of<br/>Key Personnels</td><td>Similar Projects<br/>Executed</td><td>Relevant Equipment <br/>Technical Capability</td><td>Audited<br/>Accounts</td><td>Community<br/>Contributions</td>"+
		"<td>Total<br/>Score</td><td>Action</td></tr>")

	for rows.Next() {
		var b BidSectionScore
		if err := rows.Scan(&b.id, &b.contract_id, &b.bidder_name, &b.financial_capability, &b.qualification,
			&b.similar_projects, &b.technical_capability, &b.audited_account, &b.projects, &b.total_score); err != nil {
			log.Fatal(err) // Consider logging the error rather than killing the application
		}
		bidsArray = append(bidsArray, fmt.Sprintf("<tr><td style='text-align:center;'>%s.</td><td class='text-center'>%s</td><td class='text-center'>%s </td><td class='text-center'>%s </td><td class='text-center'>%s </td><td class='text-center'>%s </td><td class='text-center'>%s </td><td class='text-center'>%s </td><td class='text-center bg-yellow-300 font-semibold'>%s </td><td style='text-align:center'>%s</td></tr>",
			b.id, b.bidder_name, b.financial_capability, b.qualification, b.similar_projects, b.technical_capability, b.audited_account, b.projects, b.total_score,
			"<a style='text-decoration:underline;' class='text-blue-500' href='/contract_home?id="+b.contract_id+"'>Open</a>"))
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
