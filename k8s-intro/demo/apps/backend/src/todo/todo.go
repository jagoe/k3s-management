package todo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"../util"
)

// List is a list of Items
type List struct {
	todos []Item
	file  *os.File
}

// ItemList is a list of Items solely for the purpose of marshaling
type ItemList struct {
	// Todo list
	Todos []Item `json:"todos"`
}

// Item is a todo
type Item struct {
	// Id of the todo
	ID string `json:"id"`
	// Text of the todo
	Text string `json:"text"`
	// Done state of the todo
	Done bool `json:"done"`
}

// NewList creates a new todo list
func NewList(path string) *List {
	_ = os.Mkdir(filepath.Dir(path), os.ModeDir|0770)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)
	util.HandleError(err)

	bytes, err := ioutil.ReadAll(file)
	util.HandleError(err)

	if len(bytes) == 0 {
		bytes = []byte("{\"todos\": []}")
	}

	var itemList ItemList

	err = json.Unmarshal(bytes, &itemList)
	util.HandleError(err)

	list := &List{file: file, todos: itemList.Todos}

	return list
}

// Close closes the file associated with the todo list
func (list *List) Close() {
	list.file.Close()
}

// All returns all todos in the list
func (list *List) All() []Item {
	return list.todos
}

// Add a new todo to the list
func (list *List) Add(text string) {
	todo := Item{ID: util.RandomString(32), Text: text}
	list.todos = append(list.todos, todo)

	list.save()
}

// Delete a todo from the list
func (list *List) Delete(id string) error {
	index, err := list.find(id)
	if err != nil {
		return err
	}

	list.todos = append(list.todos[:index], list.todos[index+1:]...)

	list.save()

	return nil
}

// SetState sets the done state of a todo
func (list *List) SetState(id string, done bool) error {
	index, err := list.find(id)
	if err != nil {
		return err
	}

	list.todos[index].Done = done

	list.save()

	return nil
}

func (list *List) find(id string) (int, error) {
	for i := range list.todos {
		if list.todos[i].ID == id {
			return i, nil
		}
	}

	return -1, fmt.Errorf("Can't find todo with id %s", id)
}

func (list *List) save() {
	itemList := ItemList{Todos: list.todos}

	// truncate file, otherwise JSON will just be appended
	list.file.Truncate(0)
	list.file.Seek(0, 0)

	// write JSON
	encoder := json.NewEncoder(list.file)
	encoder.SetIndent("", "")
	err := encoder.Encode(itemList)
	util.HandleError(err)
}
