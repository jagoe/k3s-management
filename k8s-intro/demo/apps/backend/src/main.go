package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"./todo"
	"./util"
)

var todoList *todo.List

func main() {
	port, storage := parseFlags()

	todoList = todo.NewList(storage + "/todos.json")
	defer todoList.Close()

	var badRequestError error
	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		fmt.Println(r.Method, r.RequestURI)

		switch r.Method {
		case "GET":
			list(w)
		case "POST":
			add(w, r)
		case "DELETE":
			badRequestError = delete(w, r)
		case "PUT":
			badRequestError = setState(w, r)
		case "OPTIONS":
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
		}

		if badRequestError != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(badRequestError.Error()))
		}
	})

	addr := "0.0.0.0:" + strconv.Itoa(port)
	fmt.Println("Listening on http://" + addr)
	err := http.ListenAndServe(addr, nil)
	util.HandleError(err)
}

func parseFlags() (int, string) {
	port := flag.Int(
		"p", int(80), "Port to listen on",
	)
	storage := flag.String(
		"s", "/storage", "Storage directory",
	)
	flag.Parse()

	return *port, *storage
}

func list(w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(todoList.All())
	util.HandleError(err)
}

func add(w http.ResponseWriter, r *http.Request) error {
	text, err := ioutil.ReadAll(r.Body)
	util.HandleError(err)

	if len(text) == 0 {
		return fmt.Errorf("Please provide a todo text")
	}

	todoList.Add(string(text))

	list(w)

	return nil
}

func delete(w http.ResponseWriter, r *http.Request) error {
	id, err := ioutil.ReadAll(r.Body)
	util.HandleError(err)

	if len(id) == 0 {
		return fmt.Errorf("Please provide a todo id")
	}

	err = todoList.Delete(string(id))
	if err != nil {
		return err
	}

	list(w)

	return nil
}

func setState(w http.ResponseWriter, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	util.HandleError(err)

	if len(body) == 0 {
		return fmt.Errorf("Please provide a todo state of either 'done' or 'open'")
	}

	content := strings.Split(string(body), ":")
	id := content[0]
	state := content[1]
	var done bool
	switch state {
	case "done":
		done = true
	case "open":
		done = false
	default:
		return fmt.Errorf("Please provide a todo state of either 'done' or 'open'")
	}

	err = todoList.SetState(id, done)
	if err != nil {
		return err
	}

	list(w)

	return nil
}
