package main

import (
	apiLib "github.com/hornbill/goApiLib"
	logrus "github.com/sirupsen/logrus"
)

const (
	version   = "1.0.0"
	logPrefix = "docimport"
)

var (
	counters       counterStruct
	csvContent     []csvStruct
	csvShares      = make(map[string][]sharesStruct)
	csvCollections = make(map[string][]int)
	csvTags        = make(map[string][]string)
	espXmlmc       *apiLib.XmlmcInstStruct
	flags          flagsStruct
	logFile        = logrus.New()
	logStdOut      = logrus.New()
)

type counterStruct struct {
	session struct {
		addSuccess    uint16
		addFailed     uint16
		deleteSuccess uint16
		deleteFailed  uint16
	}
	documents struct {
		addSuccess uint16
		addFailed  uint16
	}
	collections struct {
		addSuccess uint16
		addFailed  uint16
	}
	shares struct {
		addSuccess uint16
		addFailed  uint16
	}
	tags struct {
		addSuccess uint16
		addFailed  uint16
	}
}

type flagsStruct struct {
	configAPIKey         string
	configAPITimeout     int
	configCSVMain        string
	configCSVShares      string
	configCSVCollections string
	configCSVTags        string
	configDebug          bool
	configDryRun         bool
	configInstanceID     string
	configVersion        bool
}

type csvStruct struct {
	Filepath          string
	Title             string
	Status            string
	Description       string
	VersioningEnabled bool
	ReviewDate        string
	Collections       string
	Shares            []sharesStruct
	Owner             string
	Tags              string
	Filename          string
	SessionPath       string
	ContentType       string
	DocumentID        string
	ActivityStreamID  string
}

type sharesStruct struct {
	URN            string
	Read           bool
	ModifyContent  bool
	ModifyMetaData bool
}

type xmlmcResponseStruct struct {
	MethodResult     string       `xml:"status,attr"`
	State            stateStruct  `xml:"state"`
	DocumentID       string       `xml:"params>documentId"`
	ActivityStreamID string       `xml:"params>activityStreamId"`
	HPKID            string       `xml:"params>h_pk_id"`
	TagsFound        []tagsStruct `xml:"params>name"`
	TagID            int          `xml:"params>tagId"`
}

type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}

type tagsStruct struct {
	ID   int    `xml:"tagId"`
	Name string `xml:"text"`
}
