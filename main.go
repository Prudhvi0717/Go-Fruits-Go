package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "521235"
	dbname   = "FruitBase"
)

type Fruit struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Qty   int    `json:"quantity"`
}

func respGen(text string) string {
	m := make(map[string]string)
	m["Response"] = text
	jsonStr, _ := json.Marshal(m)
	return string(jsonStr)
}

func addNewFruit(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var fruit Fruit
	_ = json.NewDecoder(req.Body).Decode(&fruit)

	err := db.QueryRow(`SELECT id FROM fruits WHERE name = $1`, fruit.Name).Scan(&fruit.Id)

	if err == sql.ErrNoRows {
		err := db.QueryRow(`INSERT INTO fruits (name, price, qty)
				VALUES ($1, $2, $3)
				RETURNING id`, fruit.Name, fruit.Price, fruit.Qty).Scan(&fruit.Id)

		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(fruit)

	} else {

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(respGen("The Fruit Already Exists! You can Update the details instead of Add!"))
		return
	}
}

func updateFruit(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var updFrt Fruit
	_ = json.NewDecoder(req.Body).Decode(&updFrt)
	err := db.QueryRow(`UPDATE fruits SET name = $1,
			price = $2, qty = $3 WHERE id = $4 RETURNING id`,
		updFrt.Name, updFrt.Price, updFrt.Qty, updFrt.Id).Scan(&updFrt.Id)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(respGen("The mentioned Fruit ID Doesnt exist!"))
	} else {
		json.NewEncoder(w).Encode(updFrt)
	}
}

func buyFruit(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(req)
	name := params["name"]

	qty, _ := strconv.Atoi(params["qty"])
	var curQuant int
	var price int
	err := db.QueryRow(`SELECT qty, price FROM fruits WHERE name = $1`, name).Scan(&curQuant, &price)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(respGen("The Fruit Doesn't Exist! Please check the purchase request"))
		return
	}
	if curQuant >= qty {
		_, _ = db.Query(`UPDATE fruits SET qty = $1 WHERE name = $2`, curQuant-qty, name)
		json.NewEncoder(w).Encode(respGen(fmt.Sprintf("Your Total Bill is %d", (price * qty))))

	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(respGen("The Purchase Quantity exceeds the available Quantity"))
	}
}

func getFruitsMenu(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var fruits []Fruit
	rows, _ := db.Query("SELECT * FROM fruits")
	for rows.Next() {
		var fruit Fruit
		_ = rows.Scan(&fruit.Id, &fruit.Name, &fruit.Price, &fruit.Qty)
		fruits = append(fruits, fruit)
	}
	json.NewEncoder(w).Encode(fruits)
}

func deleteFruit(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, _ := strconv.Atoi(mux.Vars(req)["id"])

	_, _ = db.Query(`DELETE FROM fruits WHERE id = $1`, id)
	json.NewEncoder(w).Encode(respGen(fmt.Sprintf("The Fruit with ID : %v is Deleted", id)))
}

func main() {
	//creating a mux router
	router := mux.NewRouter()

	//connecting the DataBase
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, _ = sql.Open("postgres", psqlInfo)
	defer db.Close()
	err := db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	// EndPoints
	router.HandleFunc("/menu", getFruitsMenu).Methods("GET")
	router.HandleFunc("/buyFruit/{name}/{qty}", buyFruit).Methods("GET")
	router.HandleFunc("/addFruit", addNewFruit).Methods("POST")
	router.HandleFunc("/updateFruit", updateFruit).Methods("POST")
	router.HandleFunc("/deleteFruit/{id}", deleteFruit).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}
