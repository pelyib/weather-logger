package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pelyib/weather-logger/internal"
)

func add(fs []internal.Forecast) {
  gbd := make(map[string][]internal.Forecast, 0)
  dates := []string{}

  for _, nf := range fs {
    nft, _ := time.Parse(time.RFC3339, nf.At)
    d := nft.Format("2006_01_02")

    if _, ok := gbd[d]; ok {
      gbd[d] = append(gbd[d], nf)
    } else {
      dates = append(dates, d)
      gbd[d] = []internal.Forecast{nf}
    }
  }

  cnf, err := internal.CreateLoggerConf()
  if err != nil {
    fmt.Errorf(err.Error())
  }

  for _, d := range dates {
    fp := fmt.Sprintf("%s%s", cnf.Database.Folder, d)

    if _, err := os.Stat(fp); err != nil {
      fmt.Println(err)
      os.Create(fp)
    }

    if f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777); err == nil {
      nl := func(fs []internal.Forecast) string {
        asString := ""

        for _, fc := range fs {
          fjson, _ := json.Marshal(fc)
          fmt.Println("json serialized fc: >" + string(fjson) + "<")
          asString = asString + "\n" + string(fjson)
        }

        return asString
      }(gbd[d])

      if _, err := f.WriteString(nl); err != nil {
        fmt.Errorf(err.Error())
      }

      f.Sync()
      f.Close()
    }
  }
}
