// package main

// import (
// 	"log/slog"

// 	"github.com/go-ping/ping"
// 	"github.com/xuri/excelize/v2"
// )

// const(
// 	ipFileName="ips.xlsx"
// 	firstSheet="1"
// )

// func main() {
// 	f,err:= excelize.OpenFile(ipFileName)
// 	if err!=nil{
// 		slog.Error("can't open file %s, with error %s",ipFileName,err.Error() )
// 	}
// 	rows,err:=f.GetRows("")
// 	if err!=nil{
// 		slog.Error("can't open read sheet %s, with error %s", firstSheet,err.Error() )
// 	}
// 	for i,row:= range rows{
// 		if i==0{
// 			continue
// 		}

// 	}

// }
package main

import (
	"fmt"
	"log/slog"
	"test/pinger"

	"github.com/xuri/excelize/v2"
)

const (
	ipFileName = "ips.xlsx"
	firstSheet = "1"
)

type Camera struct {
	Name     string
	Ip       string
	Location string
}

func main() {
	cameras := make([]Camera, 0)
	ips := make([]string, 0)
	f, err := excelize.OpenFile(ipFileName)
	if err != nil {
		slog.Error("can't open file %s, with error %s", ipFileName, err.Error())
	}
	rows, err := f.GetRows("Основной файл")
	if err != nil {
		slog.Error("can't open read sheet %s, with error %s", firstSheet, err.Error())
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		cameras = append(cameras, Camera{Name: row[1], Ip: row[3], Location: row[2]})
		ips = append(ips, row[3])

	}
	fmt.Println(cameras)

	pinger, err := pinger.NewPingManager(ips)
	if err != nil {
		fmt.Println(err)
		return
	}
	results, comments := pinger.Start()
	for ip, result := range results {
		fmt.Printf("%s: %t\n", ip, result)
	}
	for ip, comment := range comments {
		fmt.Printf("%s: %s\n", ip, comment)
	}
	fmt.Println(len(results))
	fmt.Println(len(comments))
}
