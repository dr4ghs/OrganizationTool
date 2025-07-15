package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

type Activity struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	User   string `json:"user"`
	Points int    `json:"points"`
	Goal   int    `json:"goal"`
	Type   string `json:"type"`
}

func (a *Activity) Validate() (err error) {
	fields := make([]string, 0)

	if a.Name == "" {
		fields = append(fields, "name")
	}

	if a.User == "" {
		fields = append(fields, "user")
	}

	if a.Points < 1 {
		fields = append(fields, "points")
	}

	if a.Goal < 1 {
		fields = append(fields, "goal")
	}

	if a.Type == "" ||
		!slices.Contains(
			[]string{"daily", "weekly", "monthly", "yearly"},
			strings.ToLower(a.Type),
		) {
		fields = append(fields, "type")
	}

	if len(fields) > 0 {
		err = fmt.Errorf("Invalid request: %s", strings.Join(fields, ", "))
	}

	return err
}

func (a *Activity) Save(token string) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/collections/activities/records/%s", os.Getenv("BACKEND_URL"), a.Id)
	req, _ := http.NewRequest(http.MethodPatch, url, io.NopCloser(bytes.NewReader(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNotFound {
		a.create(token)
	}

	return nil
}

func (a *Activity) create(token string) error {
	data, err := json.Marshal(
		Activity{Name: a.Name, User: a.User, Points: a.Points, Goal: a.Goal, Type: a.Type},
	)
	if err != nil {
		return err
	}

	log.Println(string(data))
	url := fmt.Sprintf("%s/api/collections/activities/records", os.Getenv("BACKEND_URL"))
	req, _ := http.NewRequest(http.MethodPost, url, io.NopCloser(bytes.NewReader(data)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(a)
}
