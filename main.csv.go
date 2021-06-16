package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
)

func getCSVDocuments() {
	lines, err := readCSV(flags.configCSVMain)
	if err != nil {
		logError(err.Error(), true)
		os.Exit(1)
	}
	for _, line := range lines {
		if strings.ToLower(line[0]) == "filepath" || line[0] == "" {
			continue
		}
		csvData := csvStruct{
			Filepath:          line[0],
			Title:             line[1],
			Status:            line[2],
			Description:       line[3],
			ReviewDate:        line[4],
			VersioningEnabled: false,
			Owner:             line[6],
		}
		versioningEnabled, err := strconv.ParseBool(line[5])
		if err == nil {
			csvData.VersioningEnabled = versioningEnabled
		}
		csvContent = append(csvContent, csvData)
	}
}

func getCSVShares() {
	lines, err := readCSV(flags.configCSVShares)
	if err != nil {
		logError(err.Error(), true)
		os.Exit(1)
	}
	for _, line := range lines {
		if strings.ToLower(line[0]) == "filepath" || line[0] == "" {
			continue
		}
		csvData := sharesStruct{
			URN:            line[1],
			Read:           false,
			ModifyContent:  false,
			ModifyMetaData: false,
		}
		read, err := strconv.ParseBool(line[2])
		if err == nil {
			csvData.Read = read
		}
		modifyContent, err := strconv.ParseBool(line[3])
		if err == nil {
			csvData.ModifyContent = modifyContent
		}
		modifyMetaData, err := strconv.ParseBool(line[4])
		if err == nil {
			csvData.ModifyMetaData = modifyMetaData
		}
		csvShares[line[0]] = append(csvShares[line[0]], csvData)
	}
}

func getCSVCollections() {
	lines, err := readCSV(flags.configCSVCollections)
	if err != nil {
		logError(err.Error(), true)
		os.Exit(1)
	}
	for _, line := range lines {
		if strings.ToLower(line[0]) == "filepath" || line[0] == "" {
			continue
		}
		collID, err := strconv.Atoi(line[1])
		if err == nil {
			csvCollections[line[0]] = append(csvCollections[line[0]], collID)
		}
	}
}

func getCSVTags() {
	lines, err := readCSV(flags.configCSVTags)
	if err != nil {
		logError(err.Error(), true)
		os.Exit(1)
	}
	for _, line := range lines {
		if strings.ToLower(line[0]) == "filepath" || line[0] == "" {
			continue
		}
		csvTags[line[0]] = append(csvTags[line[0]], line[1])
	}
}

func readCSV(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return lines, nil
}
