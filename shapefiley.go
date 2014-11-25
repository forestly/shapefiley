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
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	TmpLocation = "/tmp/shapefiley"
)

const (
	Started  = "started"
	Finished = "finished"
	Failed   = "failed"
)

var (
	WorkDatabase string
	db           gorm.DB
	workDb       gorm.DB
)

type Shapefile struct {
	Id     int64
	Name   string
	Status string

	ZipFilename string `json:"-"`

	CreatedAt time.Time `json:"-"`
	Geom      []string  `sql:"-"`
}

func (t *Shapefile) GetGeodata() {
	rows, err := workDb.Table(t.Name).Select("ST_AsGeoJSON(ST_CollectionExtract(geom, 3)) as geom2").Rows()
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var geodata string
		rows.Scan(&geodata)
		t.Geom = append(t.Geom, geodata)
	}
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

		//get the *fileheaders
		files := m.File["file"]
		for i, _ := range files {
			log.Println("Iter", i)

			name := strings.Split(files[i].Filename, ".")
			shapefile := Shapefile{
				Name:   name[0],
				Status: Started,
			}

			db.Create(&shapefile)

			log.Println("First", shapefile)

			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filename := TmpLocation + "/" + strconv.FormatInt(shapefile.Id, 10) + "_" + files[i].Filename

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

			shapefile.ZipFilename = dst.Name()
			db.Save(&shapefile)

			log.Println("Last", shapefile)

			go processFile(shapefile)
		}

		log.Println("upload")
		renderJson(w, nil)
	}
}

func processFile(shapefile Shapefile) {
	var err error

	dir, _ := os.Getwd()
	log.Println(dir)

	s2pcmd := exec.Command(dir+"/worker.sh", shapefile.ZipFilename, shapefile.Name)

	err = s2pcmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Waiting for command to finish for:", shapefile)
	err = s2pcmd.Wait()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		shapefile.Status = Failed
		db.Save(&shapefile)
		return
	}

	shapefile.Status = Finished
	db.Save(&shapefile)
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

	workDatabaseUrl := os.Getenv("SHAPEFILEY_WORK_DATABASE_NAME")
	if workDatabaseUrl == "" {
		workDatabaseUrl = "user=ayerra dbname=shapefiley_work_development sslmode=disable"
	}

	var err error
	db, err = gorm.Open("postgres", databaseUrl)
	if err != nil {
		log.Println(err)
	}

	db.AutoMigrate(&Shapefile{})

	workDb, err = gorm.Open("postgres", workDatabaseUrl)
	if err != nil {
		log.Println(err)
	}

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
