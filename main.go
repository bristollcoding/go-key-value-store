package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	//----Routing http REST----- (TODO: Move to RestApi.go)
	routerMux := http.NewServeMux()

	//Put http endpoint to call store.Put
	routerMux.HandleFunc("PUT /api/v1/{key}", storePutHandler)
	// Get http endpoint to call store.Get
	routerMux.HandleFunc("GET /api/v1/{key}", storeGetHandler)
	// Delete http endpoint to call store.Get
	routerMux.HandleFunc("DELETE /api/v1/{key}", storeDeleteHandler)

	port := "8080"
	fmt.Printf("Server Startet on port: %v\n", port)
	//start listening on port 8080 with routerMux as Handler
	log.Fatal(http.ListenAndServe("localhost:"+port, routerMux))
}

// Function to send Put request to store.Put ()
func storePutHandler(w http.ResponseWriter, r *http.Request) {
	//get key from path variable
	key := r.PathValue("key")
	fmt.Printf("Put Key: %v\n", key)

	//Read value from body
	value, err := io.ReadAll(r.Body)

	if err != nil {
		//Write error to Response with a 500 http status and return
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Force close body on exit
	defer r.Body.Close()

	//Call store.Put method with key:value from request (Body value is still []byte!)
	err = Put(key, string(value))
	fmt.Printf("Put saved value: %v\n", string(value))
	if err != nil {
		//Write error to Response with a 500 http status and return
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//If nothing has failed return 201 created
	w.WriteHeader(http.StatusCreated)
}

// Function to send Put request to store.Put ()
func storeGetHandler(w http.ResponseWriter, r *http.Request) {
	//get key from path variable
	key := r.PathValue("key")
	fmt.Printf("Get Key: %v\n", key)

	//Call store.Get method with key
	value, err := Get(key)
	if err != nil {
		if errors.Is(err, NoSuchKeyError) {
			//If error is NoSuchKey set status 404 not found
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			//Write error to Response with a 500 http status and return
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	fmt.Printf("Get Retrieved value: %v \n", value)
	//If nothing has failed write value to response ( default status 200 will be sended)
	_, err = fmt.Fprint(w, value)

	if err != nil {
		fmt.Printf("Error writing response: %v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
}

// Function to send Put request to store.Put ()
func storeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//get key from path variable
	key := r.PathValue("key")
	fmt.Printf("Delete Key: %v\n", key)
	// Call store.Delete
	Delete(key)

}

// ----Store operations (TODO: move to store.go)
// Store will start as a simple map of string:string (Later --> struct with mutex )
var store = make(map[string]string)

// Define Custom error
var NoSuchKeyError = errors.New("No Such Key")

// Define Put function to add elements to store
func Put(key, value string) error {

	//Idempotent so no checks for already existing value with given key ( ALWAYS OVERWRITE)
	store[key] = value

	return nil
}

// Define Get function to retrieve element by key
func Get(key string) (string, error) {
	value, ok := store[key]

	if !ok {
		return "", NoSuchKeyError
	}

	return value, nil
}

// Define Delete function to delete element by key
func Delete(key string) {

	delete(store, key)

}
