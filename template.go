package main

import (
	"html/template"
	"io"
	"log"
)

func parse(name, content string) *template.Template {
	t, err := template.New(name).Parse(content)
	if err != nil || t == nil {
		log.Fatalf("can't load template %v, err: %v", name, err)
	}
	return t
}

func templateParse(name string) *template.Template {
	layout := parse("layout", Files["layout.html"])
	t, err := layout.New(name).Parse(Files[name])
	if err != nil || t == nil {
		log.Fatalf("can't load template %v", name)
	}
	return t
}

var monthStatReportTemplate *template.Template

func init() {
	monthStatReportTemplate = templateParse("month-stat-report.html")
}

func genMonthStatReport(wr io.Writer, data interface{}) {
	err := monthStatReportTemplate.ExecuteTemplate(wr, "layout", data)
	if err != nil {
		log.Printf("The template layout exec error:%v", err)
	}
}
