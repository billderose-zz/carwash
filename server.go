package main

import (
	"encoding/json"
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
	"strings"
	"time"
)

const (
	Host = "localhost:27017"
	Port = "8081"
)

type observation struct {
	Id    int `json:"id"`
	Label int `json:"label"`
}

type observations struct {
	Id     int   `json:"id"`
	Labels []int `json:"label"`
}

func randomLabel(imagePaths []string) int {
	n := rand.Intn(len(imagePaths))
	imageLab := strings.TrimPrefix(imagePaths[n], "images/")
	imageLab = strings.TrimSuffix(imageLab, ".JPG")
	if i, err := strconv.Atoi(imageLab); err != nil {
		log.Println("Unable to generate image label, ", err, ". Loading 47.JPG.")
		return 47
	} else {
		return i
	}
}

func indexHandler(imagePaths []string, c *mgo.Collection) http.HandlerFunc {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if err := tmpl.Execute(w, randomLabel(imagePaths)); err != nil {
				log.Fatal(err)
			}
		} else if r.Method == "POST" {
			if observation, err := parsePost(r); err != nil {
				log.Println("Error parsing observation classification: ", err)
			} else {
				addObservation(c, observation)
			}
			json.NewEncoder(w).Encode(observation{Id: randomLabel(imagePaths)})
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
	files, err := filepath.Glob("images/[0-9]*.JPG")
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
	http.HandleFunc("/", indexHandler(files, collection))

	log.Println("Setup complete, server starting on port ", Port)
	http.ListenAndServe(":"+Port, nil)
}
