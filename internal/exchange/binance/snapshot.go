package binance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func getSnapshot(symbol string) (depthSnapshot, error) {
	//keep the limit 50
	symbol = strings.ToUpper(symbol)
	url := fmt.Sprintf("https://api.binance.com/api/v3/depth?symbol=%s&limit=50", symbol)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		err = fmt.Errorf("error while closing the response body %w", err)
		log.Fatal(err)
		return depthSnapshot{}, err
	}
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("request failed with status code %d", resp.StatusCode)
		return depthSnapshot{}, errors.New(err.Error())
	}
	snapshot := depthSnapshot{}
	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		err = fmt.Errorf("could not parse the response body to json %w", err)
		return depthSnapshot{}, err
	}
	snapshot.Symbol = symbol //set the symbol
	return snapshot, nil
}
