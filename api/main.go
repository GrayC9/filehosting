package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "interface.html")
}

func handleRequests() {
	http.HandleFunc("/", serveFileByKey)
	http.HandleFunc("/upload", FileUpload)
	http.HandleFunc("/upload.html", uploadPageHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func randomString(length int) string {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	str := base64.URLEncoding.EncodeToString(bytes)

	return str[:length]
}

func serveFileByKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	//folderPath := "/home/teleg/key_from_israel/temp_test/" + key
	folderPath := "/home/trd12/GolangProjects/file_exchange/temp_test/" + key

	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		http.Error(w, "Папка не найдена", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		http.Error(w, "Ошибка чтения файлов", http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		http.Error(w, "Файлы не найдены", http.StatusNotFound)
		return
	}

	file, err := os.Open(filepath.Join(folderPath, files[0].Name()))
	if err != nil {
		http.Error(w, "Ошибка открытия файла", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+files[0].Name())
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Ошибка передачи файла", http.StatusInternalServerError)
		return
	}
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	var err error
	var randFolderName string = randomString(5)
	if r.Method != "POST" {
		http.Error(w, "Метод не добавлен", http.StatusMethodNotAllowed)
		return
	}

	if r.ContentLength > 1<<11 {
		http.Error(w, "Слишком большой файл", http.StatusInternalServerError)
		return
	}

	err = r.ParseMultipartForm(1 << 10)
	if err != nil {
		fmt.Println("Ошибка")
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	var folderPath string = "temp_test/" + randFolderName
	err = os.Mkdir(folderPath, 0755)
	if err != nil {
		fmt.Println("Ошибка при создании папки:", err)
		return
	}

	f, err := os.OpenFile(folderPath+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	io.Copy(f, file)
	fmt.Println("Файл загружен")

	http.Redirect(w, r, fmt.Sprintf("/%s", randFolderName), http.StatusSeeOther)

	//_, err = fmt.Fprintf(w, "http://89.169.96.194:10000/%s\n", randFolderName)
	//if err != nil {
	//	fmt.Println("Ошибка")
	//}
	saveKeyToDb(randFolderName)
}

func saveKeyToDb(randFolderName string) {
	dbPassword, exists := os.LookupEnv("MARIADB_PASSWORD")
	if !exists {
		log.Fatal("No MARIADB_PASSWORD")
	}
	dbHost, exists := os.LookupEnv("MARIADB_HOST")
	if !exists {
		log.Fatal("No MARIADB_HOST")
	}
	var connectionString string = fmt.Sprintf("root:%s@tcp(%s:3306)/filehosting", dbPassword, dbHost)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO file_storage (file_key, start_time, end_time) VALUES (?, ?, ?)",
		randFolderName, time.Now(), time.Now().Add(time.Hour))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	handleRequests()
}
