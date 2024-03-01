package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	// driver for sqlite3 database
	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
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

	mux := http.NewServeMux()

	mux.HandleFunc("/cotacao", BuscaCotacaoHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("sqlite3", "posgolang")

	if err != nil {
		ErrorLogger.Println("Error drive sqlite3")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer db.Close()

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		ErrorLogger.Println("Time Exceeded to API https://economia.awesomeapi.com.br/json/last/USD-BRL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, error := io.ReadAll(resp.Body)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var cotacao Cotacao
	error = json.Unmarshal(body, &cotacao)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "Application/json")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Data da Cotação: " + cotacao.USDBRL.CreateDate + "\nValor da Cotação: " + cotacao.USDBRL.Bid + "\n"))

	err = InsereCotacao(db, &cotacao)

	if err != nil {
		ErrorLogger.Println("Erro InsereCotacao")
		return
	}

}

func InsereCotacao(db *sql.DB, cotacao *Cotacao) error {

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)

	defer cancel()

	stmt, err := db.PrepareContext(ctx, "insert into cotacao(data, resultado) values (?,?)")

	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := json.Marshal(cotacao)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(time.Now(), res)

	if err != nil {
		ErrorLogger.Println("Time Exceeded to InsereCotacao")
		return err
	}

	return nil

}
