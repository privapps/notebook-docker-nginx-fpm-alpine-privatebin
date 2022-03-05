package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

var urls []string
var dataPath string

func getRemoteUrl() string {
	if urls == nil {
		urls = strings.Split(getEnv("URLS", "https://bin.0xfc.de|https://bin.bus-hit.me|https://p.darklab.sh"), "|")
		log.Println(urls)
	}
	return urls[rand.Intn(len(urls))]
}

func getDataPath() string {
	if len(dataPath) <= 0 {
		dataPath = getEnv("NOTE_DATA_PATH", "/var/www/notes/")
	}
	return dataPath
}

func relay(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
		return
	}
	for i := 0; i < 10; i++ {
		url := getRemoteUrl()
		rb, code := doRelay(url, body)
		if code == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(rb))
			return
		} else {
			log.Println(url, code, rb)
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func back(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		pasteid, ok := req.URL.Query()["pasteid"]
		if !ok || len(pasteid[0]) < 4 {
			http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
			return
		}
		fileName := getDataPath() + "data/" + pasteid[0]
		if fileExists(fileName) {
			fileBytes, err := ioutil.ReadFile(fileName)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(fileBytes)
			return
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	} else if req.Method == http.MethodPut {
		xhash := req.Header.Get("X-Hash")
		pasteid, ok := req.URL.Query()["pasteid"]
		if !ok || len(pasteid[0]) < 4 || len(xhash) < 60 {
			http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
			return
		}
		fileName := getDataPath() + pasteid[0]
		if !fileExists(fileName) {
			http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
			return
		}
		key, err2 := ioutil.ReadFile(fileName)

		rbody, err := ioutil.ReadAll(req.Body)
		if err != nil || err2 != nil {
			http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
			return
		}
		bHash := sha256.Sum256([]byte(string(rbody) + string(key)))
		if xhash != fmt.Sprintf("%x", bHash[:]) {
			log.Println(xhash, string(bHash[:]))
			http.Error(w, "Server token missmatch", http.StatusForbidden)
			return
		}
		os.WriteFile(getDataPath()+"data/"+pasteid[0], rbody, 0600)
		w.WriteHeader(http.StatusCreated)
		resp := struct {
			Status int64  `json:"status"`
			Id     string `json:"id"`
			Url    string `json:"url"`
		}{
			Status: 0,
			Id:     pasteid[0],
			Url:    pasteid[0],
		}
		log.Print("Paste updated: ", pasteid)
		json.NewEncoder(w).Encode(resp)
	} else if req.Method == http.MethodPost { // register
		pasteid, pwd := createPwd()
		resp := struct {
			Id  string `json:"id"`
			Key string `json:"key"`
		}{
			Id:  pasteid,
			Key: pwd,
		}
		w.Header().Set("Content-Type", "application/json")
		log.Print("Create new: paste: ", pasteid)
		json.NewEncoder(w).Encode(resp)
		return
	} else {
		http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
	}
}

func getHost(url string) string {
	fp := strings.Index(url, "//")
	ep := strings.Index(url[fp+3:], "/")
	if ep < 0 {
		return url[fp+2:]
	} else {
		return url[fp+2 : fp+ep+3]
	}
}

func doRelay(url string, body []byte) (string, int) {
	host := getHost(url)
	client := &http.Client{}
	rreq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", 0
	}
	// rreq.Header.Set("origin", origin_host)
	rreq.Header.Set("X-Requested-With", "JSONHttpRequest")
	rreq.Header.Set("origin", host)
	resp, err := client.Do(rreq)
	if err != nil {
		return "Bad response - Data?", http.StatusServiceUnavailable
	}
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", -1
	}
	var result map[string]interface{}
	json.Unmarshal([]byte(rbody), &result)

	if result["status"] != nil && reflect.ValueOf(result["status"]).Kind() == reflect.Float64 {
		if result["status"] == float64(0) {
			result["url"] = url + fmt.Sprint(result["url"])
			jsonString, err := json.Marshal(result)
			if err != nil {
				return string(jsonString), -2
			}
			return string(jsonString), http.StatusOK
		}
	}
	log.Println(result["status"] != nil, reflect.ValueOf(result["status"]).Kind(), result["status"], result)
	return "", -3
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
func createPwd() (string, string) {
	var next = true
	var pasteid string
	var pwd string
	for next {
		pasteid = randomString(6)
		pwd = randomString(6)
		next = fileExists(getDataPath() + pasteid)
	}
	data := []byte(pwd)
	os.WriteFile(getDataPath()+pasteid, data, 0444)
	return pasteid, pwd
}
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	folder := "./static"
	port := ":" + getEnv("PORT", "8000")
	fs := http.FileServer(http.Dir(folder))
	http.HandleFunc("/relay.php", relay)
	http.HandleFunc("/back.php", back)
	http.Handle("/", fs)

	log.Println("Listening on " + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
