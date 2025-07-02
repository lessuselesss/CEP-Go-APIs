package main

import (
	"fmt"
	"log"

	"circular-api/services"
	"circular-api/models"
)

func main() {
	log.Println("--- Testing CCertificate ---")
	cert := models.NewCCertificate()
	cert.SetData("Hello, Circular World!")

	fmt.Println("Certificate Data (Plain):", cert.GetData())

	jsonCert, err := cert.GetJSONCertificate()
	if err != nil {
		log.Fatalf("Error getting JSON certificate: %v", err)
	}
	fmt.Println("JSON Certificate:", jsonCert)

	size, err := cert.GetCertificateSize()
	if err != nil {
		log.Fatalf("Error getting certificate size: %v", err)
	}
	fmt.Printf("Certificate Size: %d bytes\n", size)

	log.Println("\n--- Testing CEPAccount ---")
	account := services.NewCEPAccount()

	// Open an account
	testAddress := "0x1234567890abcdef1234567890abcdef12345678"
	account.Open(testAddress)

	// The following calls demonstrate usage but require a live network.
	// They are commented out to allow the example to run without network access.
	// fmt.Println("Updating account...")
	// if !account.UpdateAccount() {
	// 	log.Println("Failed to update account")
	// }

	// fmt.Println("Setting network...")
	// if !account.SetNetwork("mainnet") {
	// 	log.Println("Failed to set network")
	// }

	fmt.Println("Example run complete.")
}
