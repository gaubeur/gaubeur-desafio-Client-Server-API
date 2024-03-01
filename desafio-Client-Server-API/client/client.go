package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	data  string
	valor string
}

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		ErrorLogger.Println("Error Create Arquivo cotacao.txt")
		return
	}

	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		ErrorLogger.Println("Error http.NewRequestWithContext")
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		ErrorLogger.Println("Time Exceeded to http://localhost:8080")
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		ErrorLogger.Println("Error ReadAll")
		return
	}

	println(string(body))

	tamanho, err := f.Write([]byte(body))

	if err != nil {
		ErrorLogger.Println("Error na Gravação do arquivo cotacao.txt")
		return
	}

	if tamanho > 0 {
		InfoLogger.Println("Registro inserido no arquivo cotacao.txt com sucesso")
		return
	}

}
