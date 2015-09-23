package main

import (
	"crypto/tls"
	"os"

	"flag"
	"fmt"
	"log"

	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
)

func main() {

	user := flag.String("user", "example-nozzle", "user who has access to the firehose")
	password := flag.String("pass", "example-nozzle", "password for the user")
	trafficControllerURL := flag.String("tcurl", "wss://doppler.bosh-lite.com:443", "loggregator traffic controller URL and port")
	uaaURL := flag.String("uaaurl", "https://uaa.bosh-lite.com", "UAA URL")

	flag.Parse()

	uaaClient, err := uaago.NewClient(*uaaURL)
	if err != nil {
		log.Fatalf("Error creating uaa client: %s", err.Error())
	}

	var authToken string
	authToken, err = uaaClient.GetAuthToken(*user, *password, true)
	if err != nil {
		log.Fatalf("Error getting oauth token: %s. Please check your username and password.", err.Error())
	}

	connection := noaa.NewConsumer(*trafficControllerURL, &tls.Config{InsecureSkipVerify: true}, nil)

	fmt.Println("===== Streaming Firehose (will only succeed if you have admin credentials)")

	msgChan := make(chan *events.Envelope)
	go func() {
		defer close(msgChan)
		errorChan := make(chan error)
		const firehoseSubscriptionId = "firehose-a"
		go connection.Firehose(firehoseSubscriptionId, authToken, msgChan, errorChan)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()

	for msg := range msgChan {
		fmt.Printf("%v \n", msg)
	}
}
