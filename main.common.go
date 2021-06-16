package main

import (
	"flag"
	"fmt"
	"os"
)

func logInfo(s string, outputToCLI bool) {
	logFile.Info(s)
	if outputToCLI {
		logStdOut.Info(s)
	}
}

func logError(s string, outputToCLI bool) {
	logFile.Error(s)
	if outputToCLI {
		logStdOut.Error(s)
	}
}

func logDebug(s string, outputToCLI bool) {
	if flags.configDebug {
		logFile.Debug(s)
		if outputToCLI {
			logStdOut.Debug(s)
		}
	}
}

//-- Process Input Flags
func procFlags() {
	//-- Grab Flags
	flag.BoolVar(&flags.configDryRun, "dryrun", false, "Allow the Import to run without Creating Documents")
	flag.StringVar(&flags.configInstanceID, "instanceid", "", "ID of the Hornbill Instance to connect to")
	flag.StringVar(&flags.configAPIKey, "apikey", "", "API Key to use as Authentication when connecting to Hornbill Instance")
	flag.StringVar(&flags.configCSVMain, "csvd", "", "CSV file containing main document data")
	flag.StringVar(&flags.configCSVShares, "csvs", "", "CSV file containing document sharing data")
	flag.StringVar(&flags.configCSVCollections, "csvc", "", "CSV file containing document collection data")
	flag.StringVar(&flags.configCSVTags, "csvt", "", "CSV file containing document tag data")
	flag.IntVar(&flags.configAPITimeout, "apitimeout", 60, "Number of Seconds to Timeout an API Connection")
	flag.BoolVar(&flags.configDebug, "debug", false, "Log extended debug information")
	flag.BoolVar(&flags.configVersion, "version", false, "Output Version")

	//-- Parse flags
	flag.Parse()

	//-- Output config
	if !flags.configVersion {
		logInfo("---- Hornbill Document Import Utility V"+fmt.Sprintf("%v", version)+" ----", true)

		//Check mandatory flags
		required := []string{"instanceid", "apikey", "csvd"}
		seen := make(map[string]bool)
		flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
		missingFlags := false
		for _, req := range required {
			if !seen[req] {
				logError("Mandatory argument not provided: -"+req, true)
				missingFlags = true
			}
		}
		if missingFlags {
			os.Exit(2) // the same exit code flag.Parse uses
		}

		logInfo(" -dryrun     "+fmt.Sprint(flags.configDryRun), true)
		logInfo(" -instanceid "+flags.configInstanceID, true)
		logDebug("-apikey     "+flags.configAPIKey, true)
		logInfo(" -csvd        "+flags.configCSVMain, true)
		logInfo(" -csvs        "+flags.configCSVShares, true)
		logInfo(" -csvc        "+flags.configCSVCollections, true)
		logInfo(" -csvt        "+flags.configCSVTags, true)
		logInfo(" -apitimeout "+fmt.Sprint(flags.configAPITimeout), true)
		logInfo(" -debug      "+fmt.Sprint(flags.configDebug), true)
		logInfo(" -version    "+fmt.Sprint(flags.configVersion), true)
	}
}
