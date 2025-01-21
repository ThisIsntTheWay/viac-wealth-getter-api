package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thisisnttheway/viac-wealth-getter/wealth"
)

type WealthResponse struct {
	Timestamp time.Time     `json:"timestamp"` // Timestamp of last succesfully retrieved wealth
	Wealth    wealth.Wealth `json:"wealth"`
	UpToDate  bool          `json:"up_to_date"` // False = GetWealth() failed
	Error     string        `json:"error"`      // Empty if UpToDate == true
}

type WealthEntry struct {
	ID        int
	Timestamp time.Time
	Wealth    float32
}

// Cache wealth response from viac-wealth-getter
func cacheWealth(w wealth.Wealth) error {
	db, err := sql.Open("sqlite3", "./wealth.db")
	if err != nil {
		return fmt.Errorf("failed to open DB: %v", err)
	}
	defer db.Close()

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS wealth (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		total_wealth REAL NOT NULL,
		timestamp TEXT NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO wealth (total_wealth, timestamp) VALUES (?, ?);`,
		w.TotalValue,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to to insert values: %v", err)
	}

	return nil
}

// Gets most recent cached wealth entry
func getMostRecentCachedWealth() (*WealthEntry, error) {
	db, err := sql.Open("sqlite3", "./wealth.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %v", err)
	}
	defer db.Close()

	query := `
	SELECT id, total_wealth, timestamp 
	FROM wealth 
	ORDER BY id DESC 
	LIMIT 1;
	`
	row := db.QueryRow(query)

	var entry WealthEntry
	var timestampStr string

	err = row.Scan(&entry.ID, &entry.Wealth, &timestampStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no entries found in the database")
		}
		return nil, fmt.Errorf("error querying the youngest entry by ID: %v", err)
	}

	timestampStr = strings.Replace(timestampStr, " ", "T", 1)
	entry.Timestamp, err = time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp: %v", err)
	}

	return &entry, nil
}

// /wealth handler
func getWealth(w http.ResponseWriter, r *http.Request) {
	var response WealthResponse
	wealth, err := wealth.GetWealth()
	if err != nil {
		slog.Error("WEALTH", "action", "getWealth", "error", err)
		response.Error = err.Error()

		cachedEntry, err := getMostRecentCachedWealth()
		if err != nil {
			slog.Error("WEALTH", "action", "getMostRecentCachedWealth", "error", err)
			response.Error = fmt.Sprintf("Could neither get wealth nor a cached entry (%v)", err)

			out, _ := json.Marshal(response)
			http.Error(w, string(out), http.StatusInternalServerError)
			return
		}

		response.Wealth.TotalValue = cachedEntry.Wealth
		response.Timestamp = cachedEntry.Timestamp

		out, _ := json.Marshal(response)
		fmt.Fprintf(w, "%s", string(out))
		return
	}

	response.Timestamp = time.Now()
	response.UpToDate = true
	response.Wealth = wealth

	// Cache wealth
	err = cacheWealth(wealth)
	if err != nil {
		slog.Error("WEALTH", "action", "cacheWealth", "error", err)
	}

	out, _ := json.Marshal(response)
	fmt.Fprintf(w, "%s", string(out))
}

func main() {
	port := "8080"
	slog.Info("MAIN", "action", "serve", "port", port)

	http.HandleFunc("/wealth", getWealth)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
