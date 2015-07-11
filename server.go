package main

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
//"encoding/json"
)

const (
	Host = "localhost:27017"
	Port = "8081"
)

type observation struct {
	Id int `json:"id"`
	Label int `json:"label"`
}

type observations struct {
	Id     int
	Labels []int
}

func indexHandler(nImages int, c *mgo.Collection) http.HandlerFunc {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}	
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if err := tmpl.Execute(w, rand.Intn(nImages)); err != nil {
			log.Fatal(err)
		}
		if r.Method == "POST" {
			if observation, err := parsePost(r); err != nil {
				log.Println("Error parsing observation classification: ", err)
			} else {
				addObservation(c, observation)
			}
		}
	}
}

func parsePost(r *http.Request) (*observation, error) {
	if r.PostFormValue("label") == "" {
		return nil, errors.New("No image label")
	}
	label, err := strconv.Atoi(r.PostFormValue("label"))
	if err != nil {
		log.Println("Failed to parse classification label, ", err)
	}

	id, err := strconv.Atoi(r.PostFormValue("id"))
	if err != nil {
		log.Println("Failed to parse image id, ", err)
	}
	return &observation{Id: id, Label: label}, nil
}

func addObservation(c *mgo.Collection, o *observation) {
	if err := c.Find(bson.M{"id": o.Id}).One(&observations{}); err != nil {
		log.Println("No records exist with id=", o.Id, ", inserting new document...")
		if err = c.Insert(&observations{Id: o.Id, Labels: []int{o.Label}}); err != nil {
			log.Println("Failed to insert {\"id\": ", o.Id, ", \"labels\": ", []int{o.Label}, "} : ", err)
		}
	} else {
		query := bson.M{"id": o.Id}
		update := bson.M{"$push": bson.M{"labels": o.Label}}
		err = c.Update(query, update)
		if err != nil {
			log.Println("Failed to update {\"id\": ", o.Id, "} : ", err)
		}
	}
}

func main() {
	f, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Error opening log file: ", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Print("Dialing mongo host..........")
	session, err := mgo.Dial(Host)
	if err != nil {
		log.Fatalln("Error dialing mongo at ", Host, ": ", err)
	}
	defer session.Close()
	collection := session.DB("carwash").C("classification_labels")

	log.Print("Loading images.........")
	rand.Seed(time.Now().UnixNano())
	files, err := filepath.Glob("images/*.JPG")
	if err != nil {
		log.Fatalln("Unable to import images: ", err)
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { // serve images
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) { // serve images
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) { // serve css attributes
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) { // serve css attributes
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.HandleFunc("/fonts/", func(w http.ResponseWriter, r *http.Request) { // serve css attributes
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.HandleFunc("/", indexHandler(len(files), collection))

	log.Println("Setup complete, server starting on port ", Port)
	http.ListenAndServe(":"+Port, nil)
}
