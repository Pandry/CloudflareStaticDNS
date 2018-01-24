package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	//Check args
	args := os.Args[1:]
	//No check on the content (for now)
	if len(args) != 4 {
		log.Panic("Invalid arguments!\nExpected usage: cfddns <email> <API Key> <zone name> <subdomain>")
		return
	}

	//New instance of Cloudflare API
	api, err := cloudflare.New(args[1], args[0])
	if err != nil {
		log.Fatal(err)
	}

	//Get public IP
	ipresp, err := http.Get("http://ipinfo.io/ip")
	if err != nil {
		log.Fatal(err)
	}
	defer ipresp.Body.Close()
	var publicIP string
	if ipresp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(ipresp.Body)
		if err != nil {
			log.Fatal(err)
		}
		publicIP = string(bodyBytes)
	}
	var validID = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
	publicIP = validID.FindString(publicIP)

	// Fetch the zone ID
	id, err := api.ZoneIDByName(args[2])
	if err != nil {
		log.Fatal(err)
	}

	//Search for the domain
	var createRecord bool
	zones, err := api.DNSRecords(id, cloudflare.DNSRecord{Name: args[3] + "." + args[2], Type: "A"})
	if err != nil {
		log.Fatal(err)
	}
	if len(zones) > 1 {
		//Something's wrong here!
		//Delete subdomain?
	} else if len(zones) == 1 {
		//Subdomain exist!
		if zones[0].Content == publicIP {
			//The IP is the same, no need to edit
			return
		}
	} else {
		//Subdomain dosn't exist
		createRecord = true
	}

	var record cloudflare.DNSRecord
	if createRecord {
		record.Proxied = false
		record.Name = args[3]
		record.Type = "A"
		record.Content = publicIP
		record.Locked = false
		record.Proxiable = true
		_, err = api.CreateDNSRecord(id, record)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		record = zones[0]
		record.Content = publicIP
		record.Name = args[3]
		err = api.UpdateDNSRecord(id, record.ID, record)
	}

}
