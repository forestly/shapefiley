package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	TmpLocation = "/tmp/shapefiley"
)

const (
	Started  = "started"
	Finished = "finished"
)

var (
	db gorm.DB
)

type Shapefile struct {
	Id        int64
	Status    string
	Filename  string    `json:"-"`
	CreatedAt time.Time `json:"-"`
	Geom      string    `sql:"-"`
}

func (t *Shapefile) GetGeodata() {
	t.Geom = "yay done"
	// rows, err := db.Table("tree_geoms").Select("ST_AsGeoJSON(ST_CollectionExtract(geom, 3)) as geom2").Where("latin_name = ?", t.LatinName).Rows()
	// if err != nil {
	// 	log.Println(err)
	// }

	// for rows.Next() {
	// 	var geodata string
	// 	rows.Scan(&geodata)
	// 	t.GeomData = append(t.GeomData, geodata)
	// }
}

func renderJson(w http.ResponseWriter, page interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	b, err := json.Marshal(page)
	if err != nil {
		log.Println("error:", err)
		fmt.Fprintf(w, "")
	}

	w.Write(b)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseMultipartForm(100000)
		if err != nil {
			// http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm

		log.Println(m)

		//get the *fileheaders
		files := m.File["file"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			shapefile := Shapefile{
				Status: Started,
			}
			db.Create(&shapefile)

			filename := TmpLocation + "/" + strconv.FormatInt(shapefile.Id, 10) + "_" + files[i].Filename
			log.Println(filename)

			dst, err := os.Create(filename)
			defer dst.Close()
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			shapefile.Filename = dst.Name()
			db.Save(&shapefile)
			go processFile(shapefile)
		}

		log.Println("upload")
		renderJson(w, nil)
	}
}

func processFile(shapefile Shapefile) {
}

func showShapefileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shapefileId, _ := strconv.ParseInt(vars["shapefileId"], 10, 64)

	shapefile := Shapefile{
		Id: int64(shapefileId),
	}
	db.First(&shapefile)

	if shapefile.Status == Finished {
		shapefile.GetGeodata()
	}

	renderJson(w, shapefile)
}

func init() {
	databaseUrl := os.Getenv("SHAPEFILEY_DATABASE_URL")
	if databaseUrl == "" {
		databaseUrl = "user=ayerra dbname=shapefiley_development sslmode=disable"
	}

	log.Println("Database:", databaseUrl)

	var err error
	db, err = gorm.Open("postgres", databaseUrl)
	if err != nil {
		log.Println(err)
	}

	db.AutoMigrate(&Shapefile{})
}

func main() {
	r := mux.NewRouter()
	// r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/upload", uploadHandler)
	r.HandleFunc("/shapefiles/{shapefileId}", showShapefileHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)
	http.ListenAndServe(":3002", nil)
}
