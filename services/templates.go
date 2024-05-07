package services

import "html/template"

var tmpl *template.Template

//var tmplNewContract *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
	//tmplNewContract = template.Must(template.ParseFiles("templates/new_contract.html"))
}
