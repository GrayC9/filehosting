package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/upload.html", http.StatusSeeOther)
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	// Здесь вы можете отобразить HTML-страницу для загрузки файла
	http.ServeFile(w, r, "interface.html")
}

func handleRequests() {
	// Обработчик для загрузки файлов
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", FileUpload)
	http.HandleFunc("/upload.html", uploadPageHandler)
	http.HandleFunc("/download/", serveFileByKey)
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

	folderPath := "/test_temp/" + key

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

	var folderPath string = "test_temp/" + randFolderName
	err = os.Mkdir(folderPath, 0755) // 0755 - права доступа к папке
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
	_, err = fmt.Fprintf(w, "Ключ файла: %s\n", randFolderName)
	if err != nil {
		fmt.Println("Ошибка")
	}
}
func main() {
	handleRequests()

}
