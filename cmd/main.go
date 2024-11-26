package main

import (
	"bufio"
	"dbf-column-filler/internal/models"
	"dbf-column-filler/internal/translator"
	"fmt"
	"github.com/Riznyk01/dbf"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	LogFileName  = "error_log.txt"
	FailedToOpen = "Failed to open log file"
)

var logger *log.Logger

func init() {
	file, err := os.OpenFile(LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		fmt.Printf("%s: %v", FailedToOpen, err)
		os.Exit(1)
	}
	logger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	var wg sync.WaitGroup
	var params []string

	lang, err := translator.LoadTranslations()
	if err != nil {
		logger.Println(err)
	}

	fmt.Printf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n\n",
		lang["StartMessageRow1"],
		lang["StartMessageRow2"],
		lang["StartMessageRow3"],
		lang["StartMessageRow4"],
		lang["StartMessageRow5"],
		lang["StartMessageRow6"],
		lang["StartMessageRow7"])

	firstRun := alreadyRan()
	if !firstRun {
		<-time.After(2 * time.Second)
	} else {
		<-time.After(20 * time.Second)
	}

	checkForDroppedFiles(os.Args, lang)
	checkForOtherFormats(os.Args, lang)

	scanner := bufio.NewScanner(os.Stdin)

	var line string
	for {
		scanner.Scan()
		line = scanner.Text()

		params = strings.Split(line, " ")
		if len(params) != 3 {
			fmt.Printf("%s\n", lang["EnteredNotThree"])
		} else if len(params) == 0 {
			fmt.Printf("%s\n", lang["DidntEnter"])
		} else {
			break
		}

	}
	wg.Add(len(os.Args) - 1)
	go func() {
		for _, filePath := range os.Args[1:] {
			go processDBFFile(filePath, params, &wg, lang)
		}
	}()

	wg.Wait()
	fmt.Println(lang["SuccessMessage"])
	<-time.After(15 * time.Second)
}
func processDBFFile(filePath string, par []string, wg *sync.WaitGroup, lang models.Translations) {
	defer wg.Done()
	fmt.Printf("%s %s\n", lang["text.Working"], filePath)
	_, fileName := filepath.Split(filePath)
	dirPath := filepath.Dir(filePath)
	changedFilesDir := filepath.Join(dirPath, lang["OutputFolder"])
	pathForTheChangedFiles := filepath.Join(dirPath, lang["OutputFolder"], fileName)

	t, err := dbf.LoadFile(filePath)
	if err != nil {
		logger.Println(err)
		return
	}

	for i := 0; i < t.NumRecords(); i++ {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf(lang["PanicMessage"], r)
			}
		}()

		for j := 0; j < len(par); j += 3 {
			columnName := par[j]
			value1 := par[j+1]
			value2 := par[j+2]

			currentValue := t.FieldValueByName(i, columnName)
			if currentValue == value1 {
				newRowIndex := t.InsertRecord()
				// copy the data from the current processing row to the inserted row
				for _, fieldName := range t.Fields() {
					fieldValue := t.FieldValueByName(i, fieldName.Name)
					t.SetFieldValueByName(newRowIndex, fieldName.Name, fieldValue)
				}
				// change the value of column value1 to value2 in the new row
				t.SetFieldValueByName(newRowIndex, columnName, value2)
			}
		}
	}

	_, err = os.Stat(changedFilesDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(changedFilesDir, os.ModePerm)
		if err != nil {
			logger.Println(lang["CreatingFoldersError"], err)
			return
		}
	}

	err = t.SaveFile(pathForTheChangedFiles)
	if err != nil {
		logger.Println(err)
	}
	fmt.Printf("%s\n%s\n", lang["FileSavedMessage"], pathForTheChangedFiles)
}

func checkForDroppedFiles(files []string, lang models.Translations) {
	for len(files) < 2 {
		fmt.Printf("%s", lang["DropTheFiles"])
		<-time.After(2 * time.Second)
		os.Exit(0)
	}
}
func checkForOtherFormats(files []string, lang models.Translations) {
	for _, filePath := range files[1:] {
		if !strings.HasSuffix(filePath, lang["FileExt"]) {
			fmt.Printf("%s", lang["DropDBF"])
			<-time.After(2 * time.Second)
			os.Exit(0)
		}
	}
}
func alreadyRan() bool {
	filename := "already_ran"
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			println("Error creating file:", err)
		}
		defer file.Close()
		return true
	}
	return false
}
