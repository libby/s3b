package main

import (
	"crypto/rand"
	"fmt"
	"gopkg.in/alexcesaro/statsd.v2"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"github.com/urfave/cli"
)

func initialize_test_element(agent_id string, tag_size_string string, debug bool) test_element {

	// Parse the string
	element := strings.Split(tag_size_string, "=")
	element_tag := strings.TrimSpace(element[0])

	// Create a temporary file
	element_file, err := ioutil.TempFile("", fmt.Sprintf("%s_%s", agent_id, element_tag))
	if err != nil {
		log.Fatalln("Unable to create temp file!")
	} else {
		if debug {
			log.Printf("Created temp file [%s]", element_file)
		}
	}

	// Parse the element size
	element_size, err := strconv.ParseUint(element[1], 0, 0)
	if err != nil {
		log.Fatalf("Unable to convert [%s] to integer\n", element[1])
	}

	// Create a byte array and fill it with random data
	b := make([]byte, element_size)
	_, err = rand.Read(b)
	if err != nil {
		log.Fatalln("Error reading random data")
	}

	// Fill the temp file with random data
	_, err = element_file.Write(b)
	if err != nil {
		log.Fatalln("Could not write temporary file")
	}

	// Close the temporary file
	err = element_file.Close()
	if err != nil {
		log.Fatalln("Could not close the temporary file object!")
	}

	if debug {
		log.Printf("Test Element: Tag: [%s], Filename: [%s], Size: [%s]", element_tag,
			element_file.Name(), element_size)
	}

	return test_element{
		tag:          element_tag,
		tmp_filename: element_file.Name(),
		file_size:    element_size}
}

func initialize_test_matrix(agent_id string, connection_object s3_connection, c *cli.Context) test_matrix {
	// Debug
	debug := c.Bool("debug")

	// StatsD host
	statsd_host := c.String("statsd")
	if len(statsd_host) < 1 {
		log.Fatalln("StatsD host not defined!")
	}
	if debug {
		log.Printf("StatsD host: %s", statsd_host)
	}

	// StatsD prefix
	statsd_app_prefix := c.String("prefix")
	if len(statsd_app_prefix) < 1 {
		statsd_app_prefix = "s3b"
		log.Println("StatsD prefix not defined!  Using standard [%s] prefix", statsd_app_prefix)
	} else {
		if debug {
			log.Printf("StatsD prefix: [%s]", statsd_app_prefix)
		}
	}

	statsd_client, err := statsd.New(statsd.Address(statsd_host),
		statsd.Prefix(statsd_app_prefix), statsd.ErrorHandler(func(err error) {
			log.Fatalf("StatsD Error: %v", err)
		}))
	if err != nil {
		log.Fatal(err)
	}

	// Register this agent_id with StatsD
	statsd_client.Increment("agent_id")

	var test_elements []test_element

	// This variable holds a set of key value pairs.
	// The format is meant to be tag=size, tag=size, etc.
	// tag is used as a measurement data point
	// size is the size of the object to be used for testing
	matrix_string := c.String("matrix")
	if len(matrix_string) < 1 {
		log.Fatalln("Test matrix not defined!")
	}

	for _, element := range strings.Split(matrix_string, ",") {
		test_elements = append(test_elements, initialize_test_element(agent_id, element, debug))
	}

	return test_matrix{
		agent_id:          agent_id,
		connection_object: connection_object,
		test_elements:     test_elements,
		statsd_client:     statsd_client,
		debug:             debug}
}