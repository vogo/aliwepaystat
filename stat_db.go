package aliwepaystat

import (
    "database/sql"
    "log"
)

func BuildStatsFromDB(db *sql.DB) {
    monthStatsMap = make(map[string]*MonthStat)
    yearMonths = nil

    yms, err := QueryYearMonths(db)
    if err != nil {
        log.Fatal(err)
    }
    for _, ym := range yms {
        transList, err := QueryTransByYearMonth(db, ym)
        if err != nil {
            log.Fatal(err)
        }
        ms := getMonthStat(ym)
        for _, t := range transList {
            ms.add(t)
        }
    }
}