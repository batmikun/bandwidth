package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// SUBACCOUNT STRUCT XML ----------------------------------------------------------------------

type SitesResponse struct {
	XMLName xml.Name `xml:"SitesResponse"`
	Sites   Sites    `xml:"Sites"`
}

type Sites struct {
	XMLName xml.Name `xml:"Sites"`
	Site    []Site   `xml:"Site"`
}

type Site struct {
	XMLName xml.Name `xml:"Site"`
	Id      int      `xml:"Id"`
}

// END SUBACCOUNT XML --------------------------------------------------------------------------

// LOCATION STRUCT XML -------------------------------------------------------------------------

type TNSipPeersResponsestruct struct {
	XMLNAme  xml.Name `xml:"TNSipPeersResponse"`
	SipPeers SipPeers `xml:"SipPeers"`
}

type SipPeers struct {
	XMLName xml.Name  `xml:"SipPeers"`
	SipPeer []SipPeer `xml:"SipPeer"`
}

type SipPeer struct {
	XMLName xml.Name `xml:"SipPeer"`
	PeerId  string   `xml:"PeerId"`
}

// END LOCATION XML ----------------------------------------------------------------------------

// NUMBER STRUCT XML ---------------------------------------------------------------------------
type SipPeerTelephoneNumbersResponse struct {
	XMLName                 xml.Name                `xml:"SipPeerTelephoneNumbersResponse"`
	SipPeerTelephoneNumbers SipPeerTelephoneNumbers `xml:"SipPeerTelephoneNumbers"`
}

type SipPeerTelephoneNumbers struct {
	XMLName                xml.Name                 `xml:"SipPeerTelephoneNumbers"`
	SipPeerTelephoneNumber []SipPeerTelephoneNumber `xml:"SipPeerTelephoneNumber"`
	FullNumber             int                      `xml:"FullNumber"`
}

type SipPeerTelephoneNumber struct {
	XMLName    xml.Name `xml:"SipPeerTelephoneNumber"`
	FullNumber int      `xml:"FullNumber"`
}

// END NUMBER XML ------------------------------------------------------------------------------

// SUBACCOUNT STRUCTS --------------------------------------------------------------------------

type Subaccount struct {
	Id            int
	LocationCount int
	Location      []Location
}

type Location struct {
	Id      string
	Numbers []int
}

var subaccounts []Subaccount

// END SUBACCOUNT ------------------------------------------------------------------------------

const BANDWIDTH_USERNAME = ""
const BANDWIDTH_PASSWORD = ""

const BANDWIDTH_GET_SITES = "https://dashboard.bandwidth.com/api/accounts/5008946/sites"

var BANDWIDTH_GET_SIPPERS_FOR_SITE = "https://dashboard.bandwidth.com/api/accounts/5008946/sites/%d/sippeers"
var BANDWIDTH_GET_PHONES_FOR_SIPPERS = "https://dashboard.bandwidth.com/api/accounts/5008946/sites/%d/sippeers/%s/tns"

func main() {
	client := &http.Client{}
	fill_subaccounts(client, &subaccounts)

	for index, subaccount := range subaccounts {
		fill_locations(client, &subaccounts, index, subaccount.Id)
	}

	f, err := os.OpenFile("./data.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	for _, subaccount := range subaccounts {
		subaccount, _ := json.MarshalIndent(subaccount, "", " ")
		n, err := f.Write(subaccount)
		if err != nil {
			fmt.Println(n, err)
		}

		if n, err = f.WriteString("\n"); err != nil {
			fmt.Println(n, err)
		}
	}
}

func fill_subaccounts(client *http.Client, sub *[]Subaccount) {
	req, err := http.NewRequest("GET", BANDWIDTH_GET_SITES, nil)
	req.SetBasicAuth(BANDWIDTH_USERNAME, BANDWIDTH_PASSWORD)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	var subaccounts SitesResponse
	xml.Unmarshal(body, &subaccounts)

	for i := 0; i < len(subaccounts.Sites.Site); i++ {
		id := int(subaccounts.Sites.Site[i].Id)

		subaccount := Subaccount{Id: id}
		*sub = append(*sub, subaccount)
	}
}

func fill_locations(client *http.Client, sub *[]Subaccount, subaccount_index int, subaccount_id int) {

	req, err := http.NewRequest("GET", fmt.Sprintf(BANDWIDTH_GET_SIPPERS_FOR_SITE, subaccount_id), nil)
	req.SetBasicAuth(BANDWIDTH_USERNAME, BANDWIDTH_PASSWORD)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	var locations TNSipPeersResponsestruct
	xml.Unmarshal(body, &locations)

	for location_index, location := range locations.SipPeers.SipPeer {
		(*sub)[subaccount_index].Location = append((*sub)[subaccount_index].Location, Location{Id: location.PeerId})
		fill_numbers(client, &subaccounts, subaccount_index, subaccount_id, location_index, location.PeerId)
	}
}

func fill_numbers(client *http.Client, sub *[]Subaccount, subaccount_index int, subaccount_id int, location_index int, location_id string) {
	req, err := http.NewRequest("GET", fmt.Sprintf(BANDWIDTH_GET_PHONES_FOR_SIPPERS, subaccount_id, location_id), nil)
	req.SetBasicAuth(BANDWIDTH_USERNAME, BANDWIDTH_PASSWORD)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	var numbers SipPeerTelephoneNumbersResponse
	xml.Unmarshal(body, &numbers)

	for _, e := range numbers.SipPeerTelephoneNumbers.SipPeerTelephoneNumber {
		(*sub)[subaccount_index].Location[location_index].Numbers = append((*sub)[subaccount_index].Location[location_index].Numbers, e.FullNumber)
	}

}
