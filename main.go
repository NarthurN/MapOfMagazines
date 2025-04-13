package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Magazine struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

type MapOfMagazines struct {
	client *http.Client
}

func newMapOfMagazines() *MapOfMagazines {
	return &MapOfMagazines{client: &http.Client{Timeout: 1 * time.Second}}
}

func (m *MapOfMagazines) findMagazinesInCity(city string) ([]Magazine, error) {
	baseURL := "http://localhost:8080/getMagazinesByCity/"
	requestedURL := baseURL + city
	fmt.Println(requestedURL)

	req, err := http.NewRequest(http.MethodGet, requestedURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response, err := m.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var magazine []Magazine
	if err = json.NewDecoder(response.Body).Decode(&magazine); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return magazine, nil
}

func main() {
	mag := newMapOfMagazines()
	storiesOfMoscow, err := mag.findMagazinesInCity("Moscow")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(storiesOfMoscow)
}

