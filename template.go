// Copyright 2019 vogo. All rights reserved.

package aliwepaystat

import (
	_ "embed"
	"html/template"
	"io"
	"log"
)

//go:embed layout.html
var layoutTemplateData string

//go:embed index.html
var indexTemplateData string

//go:embed month-stat.html
var monthStatTemplateData string

func parse(name, content string) *template.Template {
	t, err := template.New(name).Parse(content)
	if err != nil || t == nil {
		log.Fatalf("can't load template %v, err: %v", name, err)
	}
	return t
}

func templateParse(name, pageData string) *template.Template {
	layout := parse("layout", layoutTemplateData)
	t, err := layout.New(name).Parse(pageData)
	if err != nil || t == nil {
		log.Fatalf("can't load template %v, err: %v", name, err)
	}
	return t
}

var indexStatTemplate *template.Template
var monthStatTemplate *template.Template

func init() {
	monthStatTemplate = templateParse("month-stat.html", monthStatTemplateData)
	indexStatTemplate = templateParse("index.html", indexTemplateData)
}

func genMonthStatReport(wr io.Writer, data interface{}) {
	err := monthStatTemplate.ExecuteTemplate(wr, "layout", data)
	if err != nil {
		log.Printf("The template layout exec error:%v", err)
	}
}

func genIndexStatReport(wr io.Writer) {
	data := make(map[string]interface{})
	data["yearMonths"] = yearMonths
	data["monthStatsMap"] = monthStatsMap

	err := indexStatTemplate.ExecuteTemplate(wr, "layout", data)
	if err != nil {
		log.Printf("The template layout exec error:%v", err)
	}
}
