package main

import (
	"fmt"
	"os"
	"time"

	apiLib "github.com/hornbill/goApiLib"
	logrus "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {
	//Setup logging
	cwd, _ := os.Getwd()
	logPath := cwd + "/log"
	//-- If Folder Does Not Exist then create it
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err := os.Mkdir(logPath, 0777)
		if err != nil {
			fmt.Println("Error Creating Log Folder ", logPath, ": ", err)
			os.Exit(101)
		}
	}
	logFileName := logPath + "/" + logPrefix + "_" + time.Now().Format("20060102150405") + ".log"
	f, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		os.Exit(1)
	}

	//Setup logrus to FILE ONLY
	logFile = &logrus.Logger{
		Out:   f,
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "%time% [%lvl%] %msg%\n",
		},
	}

	//Setup logrus to STDOUT
	logStdOut.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	logStdOut.SetOutput(os.Stdout)
}

func main() {
	//Process flags
	procFlags()

	//If configVersion just output version number and die
	if flags.configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//Hornbill Session
	espXmlmc = apiLib.NewXmlmcInstance(flags.configInstanceID)
	espXmlmc.SetAPIKey(flags.configAPIKey)
	espXmlmc.SetTimeout(flags.configAPITimeout)

	//Grab CSV Data
	getCSVDocuments()
	if len(csvContent) > 0 {
		if flags.configCSVShares != "" {
			getCSVShares()
		}
		if flags.configCSVCollections != "" {
			getCSVCollections()
		}
		if flags.configCSVTags != "" {
			getCSVTags()
		}
		processDocuments()
	} else {
		logError("No rows found in "+flags.configCSVMain, true)
	}
	logInfo("Processing Complete!", true)

	logInfo("🟢 Files added to Hornbill Session: "+fmt.Sprint(counters.session.addSuccess), true)
	if counters.session.addFailed > 0 {
		logInfo("🔴 Errors adding files to Hornbill Session: "+fmt.Sprint(counters.session.addFailed), true)
	}

	logInfo("🟢 Documents successfully added: "+fmt.Sprint(counters.documents.addSuccess), true)
	if counters.documents.addFailed > 0 {
		logInfo("🔴 Errors adding Documents: "+fmt.Sprint(counters.documents.addFailed), true)
	}

	logInfo("🟢 Documents Collections successfully associated: "+fmt.Sprint(counters.collections.addSuccess), true)
	if counters.collections.addFailed > 0 {
		logInfo("🔴 Errors adding Documents to Collections: "+fmt.Sprint(counters.collections.addFailed), true)
	}

	logInfo("🟢 Document Shares successfully created: "+fmt.Sprint(counters.shares.addSuccess), true)
	if counters.shares.addFailed > 0 {
		logInfo("🔴 Errors Sharing Documents: "+fmt.Sprint(counters.shares.addFailed), true)
	}

	logInfo("🟢 Document Tags successfully applied: "+fmt.Sprint(counters.tags.addSuccess), true)
	if counters.tags.addFailed > 0 {
		logInfo("🔴 Errors Tagging Documents: "+fmt.Sprint(counters.tags.addFailed), true)
	}

	logInfo("🟢 Files cleaned from Hornbill Session: "+fmt.Sprint(counters.session.deleteSuccess), true)
	if counters.session.addFailed > 0 {
		logInfo("🔴 Errors cleaning files from Hornbill Session: "+fmt.Sprint(counters.session.deleteFailed), true)
	}
}
