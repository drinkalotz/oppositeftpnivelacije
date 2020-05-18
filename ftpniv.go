package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	// "net/mail"
	// "net/smtp"
	// "time"

	_ "github.com/denisenkom/go-mssqldb"
	// "github.com/scorredoira/email"
	"github.com/secsy/goftp"
)

var (
	debug    = flag.Bool("debug", false, "debuging")
	password = flag.String("password", "Ivadp35", "password")
	server   = flag.String("server", "192.168.1.205\\DATALAB", "server")
	user     = flag.String("user", "sa", "user")
)

// Nivelacije struktura
type Nivelacije struct {
	sku  string
	cena string
}

func main() {
	fmt.Println("-> Eksport nivelacija za webshop.")
	ftpConfig := goftp.Config{
		User:     "opposite_import",
		Password: "DhTh43Zk9Hs9BfWBBxKu",
	}

	file, err := os.Create("nivelacije.csv")
	checkError("Greska prilikom kreiranja fajla: ", err)
	// defer file.Close()

	writer := csv.NewWriter(file)
	// defer writer.Flush()

	var nivelacije Nivelacije
	flag.Parse()

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s", *server, *user, *password)
	if *debug {
		fmt.Printf(" connString:%s\n", connString)
	}

	conn, err := sql.Open("mssql", connString)
	if err != nil {
		checkError("Greska prilikom konekcije na SQL: ", err)
	} else {
		fmt.Println("-> Uspesno konektovan na SQL Server.")
	}
	defer conn.Close()

	rows, err := conn.Query("exec oppositemp..__exportNivelacija")
	if err != nil {
		checkError("Greska prilikom pozivanja stored procedure: ", err)
	} else {
		fmt.Println("-> Uspesno pokrenuta stored procedura.")
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&nivelacije.sku, &nivelacije.cena)

		if err != nil {
			log.Fatal(err)
		}
		var csvData = []string{nivelacije.sku, nivelacije.cena}
		err = writer.Write(csvData)
		checkError("Greska prilikom pisanja u CSV fajl: ", err)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	client, err := goftp.DialConfig(ftpConfig, "138.201.225.130")
	if err != nil {
		checkError("Greska prilikom konekcije na FTP server: ", err)
	} else {
		fmt.Println("-> Uspesno kreiran client ka FTP serveru.")
		writer.Flush()
		file.Close()
	}

	uploadFile, err := os.Open("nivelacije.csv")
	if err != nil {
		checkError("Greska prilikom citanja fajla: ", err)
	} else {
		fmt.Println("-> Uspesno procitan fajl.")
	}

	err = client.Store("/niv_export.csv", uploadFile)
	if err != nil {
		checkError("Greska prilikom upload-a na FTP server: ", err)
	} else {
		fmt.Println("-> Uspesno upload-ovan fajl na FTP server.")
	}

	os.Remove("nivelacije.csv")
	fmt.Println("-> Obrisan fajl.")

	fmt.Print("Pritisni Enter za nastavak...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

}

func checkError(message string, err error) {
	if err != nil {
		fmt.Println(message)
		fmt.Println(err.Error())
	}
}
