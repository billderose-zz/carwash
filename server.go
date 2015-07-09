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
)

const (
	Host = "localhost:27017"
	Port = "8081"
)

type observation struct {
	id    int
	label int
}

type observations struct {
	Id     int
	Labels []int
}

func mainHandler(nImages int, c *mgo.Collection) http.HandlerFunc {
	tmpl, err := template.ParseFiles("main.html")
	if err != nil {
		log.Fatal(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, rand.Intn(nImages)); err != nil {
			log.Fatal(err)
		}
		if r.Method == "POST" {
			if observation, err := parsePost(r); err == nil {
				addObservation(c, observation)
			}
		}
	}
}

func parsePost(r *http.Request) (*observation, error) {
	r.ParseForm()
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
	return &observation{id: id, label: label}, nil
}

func addObservation(c *mgo.Collection, o *observation) {
	if err := c.Find(bson.M{"id": o.id}).One(&observations{}); err != nil {
		log.Println("No records exist with id=", o.id, ", inserting new document...")
		if err = c.Insert(&observations{Id: o.id, Labels: []int{o.label}}); err != nil {
			log.Println("Failed to insert {\"id\": ", o.id, ", \"labels\": ", []int{o.label}, "} : ", err)
		}
	} else {
		query := bson.M{"id": o.id}
		update := bson.M{"$push": bson.M{"labels": o.label}}
		err = c.Update(query, update)
		if err != nil {
			log.Println("Failed to update {\"id\": ", o.id, "} : ", err)
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

	http.HandleFunc("/", mainHandler(len(files), collection))
	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) { // serve images
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	http.ListenAndServe(":"+Port, nil)
	log.Println("Setup complete, server running on port ", Port)
}
