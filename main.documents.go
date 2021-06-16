package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func processDocuments() {
	logInfo("Processing "+strconv.Itoa(len(csvContent))+" files", true)
	for _, file := range csvContent {
		//Process filename and title
		logInfo("Processing: "+file.Filepath, true)
		file.Filename = filepath.Base(file.Filepath)
		file.SessionPath = "session/" + file.Filename
		if file.Title == "" {
			file.Title = strings.Replace(file.Filename, filepath.Ext(file.Filename), "", 1)
		}

		//Add the file to the session
		err := putFileInSession(&file)
		if err != nil {
			logError(err.Error(), true)
			counters.session.addFailed++
			continue
		}
		counters.session.addSuccess++

		//documentAdd API to create doc from session file
		docID, err := documentAdd(&file)
		if err != nil {
			logError(err.Error(), true)
			counters.documents.addFailed++
		} else {
			counters.documents.addSuccess++
			if file.Owner != "" {
				err = documentSetOwner(docID, file.Owner)
				if err != nil {
					logError(err.Error(), true)
				}
			}
			//Process Collections
			if _, ok := csvCollections[file.Filepath]; ok {
				for _, collectionID := range csvCollections[file.Filepath] {
					err = addToCollection(file.DocumentID, collectionID)
					if err != nil {
						counters.collections.addFailed++
						logError(err.Error(), true)
					} else {
						counters.collections.addSuccess++
					}
				}
			}

			//Process Shares
			if _, ok := csvShares[file.Filepath]; ok {
				for _, share := range csvShares[file.Filepath] {
					err = shareDocument(file.DocumentID, share)
					if err != nil {
						counters.shares.addFailed++
						logError(err.Error(), true)
					} else {
						counters.shares.addSuccess++
					}
				}
			}

			//Process Tags
			if _, ok := csvTags[file.Filepath]; ok {
				for _, tag := range csvTags[file.Filepath] {
					err = processTag(file.DocumentID, tag)
					if err != nil {
						counters.tags.addFailed++
						logError(err.Error(), true)
					} else {
						counters.tags.addSuccess++
					}
				}
			}
		}

		//Delete the processed file from the session
		err = deleteFileFromSession(&file)
		if err != nil {
			logError(err.Error(), true)
			counters.session.deleteFailed++
		} else {
			counters.session.deleteSuccess++
		}
	}
}

func putFileInSession(file *csvStruct) error {
	logInfo("Uploading: "+file.Filepath, false)
	//Get file content
	fileContent, err := getFileContent(file.Filepath)
	if err != nil {
		return err
	}

	//Work out file content type
	file.ContentType = http.DetectContentType(fileContent)
	logDebug("Content Type: "+file.ContentType, false)
	if err != nil {
		return err
	}

	//Work out destination
	endpoint := espXmlmc.DavEndpoint + file.SessionPath
	logDebug("Destination: "+endpoint, false)

	//PUT file in to API Key users session
	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(fileContent))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", file.ContentType)
	req.Header.Set("Authorization", "ESP-APIKEY "+flags.configAPIKey)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}
	logInfo("Upload Success: "+endpoint, false)
	return nil
}

func deleteFileFromSession(file *csvStruct) error {
	//Work out file for deletion
	endpoint := espXmlmc.DavEndpoint + file.SessionPath
	logInfo("Deleting: "+endpoint, false)

	//Perform PUT
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "ESP-APIKEY "+flags.configAPIKey)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 204 {
		return errors.New(res.Status)
	}
	logInfo("Delete Success", false)
	return nil
}

func documentAdd(file *csvStruct) (string, error) {
	logInfo("Creating Document "+file.Title, false)
	docID := ""
	espXmlmc.SetParam("title", file.Title)
	if file.Description != "" {
		espXmlmc.SetParam("description", file.Description)
	}
	espXmlmc.SetParam("status", file.Status)
	if file.ReviewDate != "" {
		espXmlmc.SetParam("reviewDate", file.ReviewDate)
	}
	if file.VersioningEnabled {
		espXmlmc.SetParam("enableRevisionTracking", strconv.FormatBool(file.VersioningEnabled))
	}
	espXmlmc.OpenElement("serverFile")
	espXmlmc.SetParam("fileName", file.Filename)
	espXmlmc.SetParam("fileSource", "/"+file.SessionPath)
	espXmlmc.SetParam("mimeType", file.ContentType)
	espXmlmc.CloseElement("serverFile")

	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[library::documentAdd] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("library", "documentAdd")
		if err != nil {
			return docID, err
		}
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)
		var xmlmcResponse xmlmcResponseStruct

		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return docID, err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return docID, errors.New(xmlmcResponse.State.ErrorRet)
		}
		file.DocumentID = xmlmcResponse.DocumentID
		file.ActivityStreamID = xmlmcResponse.ActivityStreamID
		logInfo("Document "+file.DocumentID+" Created Successfully", false)
		docID = file.DocumentID
	} else {
		logInfo("[DRYRUN] library::documentAdd:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return docID, nil
}
func documentSetOwner(documentID, owner string) error {
	logInfo("Setting Owner "+owner+" against Document "+documentID, false)
	espXmlmc.SetParam("documentId", documentID)
	espXmlmc.SetParam("owner", "urn:sys:user:"+owner)
	espXmlmc.SetParam("reason", "Owner set during import process")
	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[library::documentChangeOwner] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("library", "documentChangeOwner")
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)
		if err != nil {
			return err
		}

		var xmlmcResponse xmlmcResponseStruct
		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return errors.New(xmlmcResponse.State.ErrorRet)
		}
		logInfo("Document Owner Set Successfully", false)
	} else {
		logInfo("[DRYRUN] library::documentChangeOwner:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return nil
}

func addToCollection(documentID string, collectionID int) error {
	logInfo("Adding Document "+documentID+" to Collection "+strconv.Itoa(collectionID), false)
	espXmlmc.SetParam("collectionId", strconv.Itoa(collectionID))
	espXmlmc.SetParam("documentId", documentID)
	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[apps/com.hornbill.docmanager/Collection::addToCollection] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("apps/com.hornbill.docmanager/Collection", "addToCollection")
		if err != nil {
			return err
		}
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)

		var xmlmcResponse xmlmcResponseStruct
		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return errors.New(xmlmcResponse.State.ErrorRet)
		}
		logInfo("Document Added to Collection Successfully", false)
	} else {
		logInfo("[DRYRUN] apps/com.hornbill.docmanager/Collection::addToCollection:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return nil
}

func shareDocument(documentID string, shareDetails sharesStruct) error {
	logInfo("Sharing Document "+documentID+" with "+shareDetails.URN, false)
	espXmlmc.SetParam("documentId", documentID)
	espXmlmc.SetParam("share", shareDetails.URN)
	espXmlmc.OpenElement("permissions")
	espXmlmc.SetParam("read", strconv.FormatBool(shareDetails.Read))
	espXmlmc.SetParam("modifyContent", strconv.FormatBool(shareDetails.ModifyContent))
	espXmlmc.SetParam("modifyMetaData", strconv.FormatBool(shareDetails.ModifyMetaData))
	espXmlmc.CloseElement("permissions")
	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[library::documentShare] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("library", "documentShare")
		if err != nil {
			return err
		}
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)

		var xmlmcResponse xmlmcResponseStruct
		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return errors.New(xmlmcResponse.State.ErrorRet)
		}
		logInfo("Document Shared Successfully: "+xmlmcResponse.HPKID, false)
	} else {
		logInfo("[DRYRUN] library::documentShare:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return nil
}

func processTag(documentID, tag string) error {
	//Does tag exist
	tagExists, tagID, err := findTag(tag)
	if err != nil {
		return err
	}
	if !tagExists {
		tagID, err = addTag(tag)
		if err != nil {
			return err
		}
	}
	err = linkTag(documentID, tagID)
	return err
}

func findTag(tag string) (bool, int, error) {
	logInfo("Searching For Tag: "+tag, false)
	tagID := 0
	tagExists := false
	if tagKey, ok := foundTags[tag]; ok {
		logInfo("Tag Found In Cache: "+strconv.Itoa(tagKey), false)
		return true, tagKey, nil
	}
	espXmlmc.SetParam("tagGroup", "urn:tagGroup:library")
	//Escape backslash in tag
	tagregex := regexp.MustCompile(`\\`)
	tagSearch := tagregex.ReplaceAllString(tag, "\\\\")
	espXmlmc.SetParam("nameFilter", tagSearch)

	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[library::tagGetList] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("library", "tagGetList")
		if err != nil {
			return tagExists, tagID, err
		}
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)

		var xmlmcResponse xmlmcResponseStruct
		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return tagExists, tagID, err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return tagExists, tagID, errors.New(xmlmcResponse.State.ErrorRet)
		}
		for _, v := range xmlmcResponse.TagsFound {
			if strings.EqualFold(v.Name, tag) {
				logInfo("Tag Found: "+strconv.Itoa(v.ID), false)
				tagExists = true
				tagID = v.ID
				foundTags[tag] = v.ID
			}
		}
		if !tagExists {
			logInfo("Tag Not Found", false)
		}
	} else {
		logInfo("[DRYRUN] library::tagGetList:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return tagExists, tagID, nil
}

func addTag(tag string) (int, error) {
	logInfo("Creating Tag: "+tag, false)
	tagID := 0
	espXmlmc.SetParam("tagGroup", "urn:tagGroup:library")
	espXmlmc.OpenElement("tag")
	espXmlmc.SetParam("text", tag)
	espXmlmc.CloseElement("tag")

	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[library::tagCreate] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("library", "tagCreate")
		if err != nil {
			return tagID, err
		}
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)

		var xmlmcResponse xmlmcResponseStruct
		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return tagID, err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return tagID, errors.New(xmlmcResponse.State.ErrorRet)
		}
		tagID = xmlmcResponse.TagID
		foundTags[tag] = tagID
		logInfo("Tag Created Successfully: "+strconv.Itoa(tagID), false)
	} else {
		logInfo("[DRYRUN] library::tagCreate:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return tagID, nil
}

func linkTag(documentID string, tagID int) error {
	logInfo("Linking Tag: "+strconv.Itoa(tagID)+" to Document: "+documentID, false)
	espXmlmc.SetParam("tagGroup", "urn:tagGroup:library")
	espXmlmc.SetParam("tagID", strconv.Itoa(tagID))
	espXmlmc.SetParam("objectRefUrn", "urn:lib:document:"+documentID)
	//-- Check for Dry Run
	if !flags.configDryRun {
		logDebug("[library::tagLinkObject] "+espXmlmc.GetParam(), false)
		XMLResponse, err := espXmlmc.Invoke("library", "tagLinkObject")
		if err != nil {
			return err
		}
		logDebug("[RESPONSE] "+flattenXML(XMLResponse), false)

		var xmlmcResponse xmlmcResponseStruct
		err = xml.Unmarshal([]byte(XMLResponse), &xmlmcResponse)
		if err != nil {
			return err
		}
		if xmlmcResponse.MethodResult != "ok" {
			return errors.New(xmlmcResponse.State.ErrorRet)
		}
		logInfo("Tag Linked Successfully", false)
	} else {
		logInfo("[DRYRUN] library::tagLinkObject:"+espXmlmc.GetParam(), false)
		espXmlmc.ClearParam()
	}
	return nil
}

func flattenXML(source string) string {
	re := regexp.MustCompile(`(\r?\n)|\t`)
	return re.ReplaceAllString(source, "")
}

func getFileContent(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}
