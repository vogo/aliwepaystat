// Copyright 2019 wongoo. All rights reserved.

package aliwepaystat

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
		log.Fatalf("can't load template %v, err: %v", name, err)
	}
	return t
}

var indexStatTemplate *template.Template
var monthStatTemplate *template.Template

func init() {
	monthStatTemplate = templateParse("month-stat.html")
	indexStatTemplate = templateParse("index.html")
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
