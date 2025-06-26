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
	"os"
	"test/pinger"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	ipFileName = "ips.xlsx"
	firstSheet = "1"
	mainSheet  = "Основной файл"
	infoSheet  = "Информация"
)

type Camera struct {
	Name     string
	Ip       string
	Location string
}

func main() {
	logFile, err := os.OpenFile("./archive/logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("не удалось открыть файл логов: " + err.Error())
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelError,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)

	err = Ping()
	if err != nil {
		slog.Error("Ping error", "error", err.Error())
	}
	ticker := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-ticker.C:
			err := Ping()
			if err != nil {
				slog.Error("Ping error", "error", err.Error())
			}
		}
	}

}

func Ping() error {
	slog.Info("Start ping")
	cameras := make([]Camera, 0)
	ips := make([]string, 0)
	ipCamera := make(map[string][]Camera)
	f, err := excelize.OpenFile(ipFileName)
	if err != nil {
		return fmt.Errorf("can't open file %s, with error %s", ipFileName, err.Error())
	}
	rows, err := f.GetRows(mainSheet)
	if err != nil {
		return fmt.Errorf("can't open read sheet %s, with error %s", mainSheet, err)
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		cameras = append(cameras, Camera{Name: row[1], Ip: row[3], Location: row[2]})
		ipCamera[row[3]] = append(ipCamera[row[3]], Camera{Name: row[1], Ip: row[3], Location: row[2]})
	}
	for ip, _ := range ipCamera {
		ips = append(ips, ip)
	}
	fmt.Println(cameras)
	pinger, err := pinger.NewPingManager(ips)
	if err != nil {
		return fmt.Errorf("can't create ping manager with error %s", err.Error())
	}
	results, comments := pinger.Start()
	for ip, result := range results {
		slog.Info(fmt.Sprintf("%s: %t\n", ip, result))
	}
	for ip, comment := range comments {
		slog.Info(fmt.Sprintf("%s: %s\n", ip, comment))
	}
	if err := infoToExcel(results, comments, ipCamera); err != nil {
		return fmt.Errorf("can't write info to excel with error %s", err.Error())
	}
	return nil
}

type Info struct {
	result  bool
	comment string
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func infoFileName() string {
	now := time.Now()
	return fmt.Sprintf("./archive/архив_%02d.%02d.%d.xlsx", now.Day(), now.Month(), now.Year())
}

func findNextEmptyColumnIndex(f *excelize.File, sheet string, row int) (int, error) {
	for col := 1; col <= 100; col++ {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		val, _ := f.GetCellValue(sheet, cell)
		if val == "" {
			return col, nil
		}
	}
	return 0, fmt.Errorf("нет пустых колонок")
}

func CreateExcelFile(ipCamera map[string][]Camera) error {
	f := excelize.NewFile()
	row := 2
	if _, err := f.NewSheet(infoSheet); err != nil {
		return err
	}
	if err := f.DeleteSheet("Sheet1"); err != nil {
		return err
	}
	for _, cameras := range ipCamera {
		for _, camera := range cameras {
			if err := f.SetCellValue(infoSheet, fmt.Sprintf("A%d", row), camera.Name); err != nil {
				return err
			}

			if err := f.SetCellValue(infoSheet, fmt.Sprintf("B%d", row), camera.Location); err != nil {
				return err
			}
			if err := f.SetCellValue(infoSheet, fmt.Sprintf("C%d", row), camera.Ip); err != nil {
				return err
			}
			row++
		}
	}
	if err := f.SaveAs(infoFileName()); err != nil {
		return err
	}
	return nil
}

func infoToExcel(results map[string]bool, comments map[string]string, ipCamera map[string][]Camera) error {
	IpInfo := make(map[string]Info)
	infoFileName := infoFileName()
	if !fileExists(infoFileName) {
		err := CreateExcelFile(ipCamera)
		if err != nil {
			return fmt.Errorf("can't create file %s, with error %s", infoFileName, err.Error())
		}
	}
	time.Sleep(1 * time.Second)
	infoFile, err := excelize.OpenFile(infoFileName)
	if err != nil {
		return fmt.Errorf("can't open file %s, with error %s", infoFileName, err.Error())
	}
	for ip, cameras := range ipCamera {
		for _, camera := range cameras {
			IpInfo[camera.Ip] = Info{result: results[ip], comment: comments[ip]}
		}
	}
	col, err := findNextEmptyColumnIndex(infoFile, infoSheet, 2)
	if err != nil {
		return fmt.Errorf("can't find next empty column index, with error %s", err.Error())
	}
	timeCell1, err := excelize.CoordinatesToCellName(col, -1)
	if err != nil {
		return fmt.Errorf("can't convert coordinates to cell name, with error %s", err.Error())
	}
	timeCell2, err := excelize.CoordinatesToCellName(col+1, 1)
	if err != nil {
		return fmt.Errorf("can't convert coordinates to cell name, with error %s", err.Error())
	}
	slog.Info("Insert into column", "column", col)

	if err := infoFile.MergeCell(infoSheet, timeCell1, timeCell2); err != nil {
		return fmt.Errorf("can't merge cells, with error %s", err.Error())
	}
	if err := infoFile.SetCellValue(infoSheet, timeCell1, time.Now().Format("15:04:05")); err != nil {
		return fmt.Errorf("can't set cell value, with error %s", err.Error())
	}
	for row := 2; row <= 5000; row++ {
		ip, err := infoFile.GetCellValue(infoSheet, fmt.Sprintf("C%d", row))
		slog.Info("Processing IP", "ip", ip)
		if err != nil {
			return fmt.Errorf("can't get cell value, with error %s", err.Error())
		}
		if ip == "" {
			break
		}
		cellResult, err := excelize.CoordinatesToCellName(col+1, row)
		if err != nil {
			return fmt.Errorf("can't convert coordinates to cell name, with error %s", err.Error())
		}
		cellComment, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			return fmt.Errorf("can't convert coordinates to cell name, with error %s", err.Error())
		}
		slog.Info("Ip result", "result", IpInfo[ip].result)
		if err := infoFile.SetCellValue(infoSheet, cellResult, IpInfo[ip].result); err != nil {
			return fmt.Errorf("can't set cell value, with error %s", err.Error())
		}
		slog.Info("Ip comment", "comment", IpInfo[ip].comment)
		if err := infoFile.SetCellValue(infoSheet, cellComment, IpInfo[ip].comment); err != nil {
			return fmt.Errorf("can't set cell value, with error %s", err.Error())
		}
	}
	infoFile.Save()
	return nil
}
