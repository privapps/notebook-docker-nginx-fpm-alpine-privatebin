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
	"strconv"
	"strings"
	"time"
)

var urls []string
var dataPath string

const REMOTE_TIMEOUT_SECONDS = 10

func getRemoteUrl() string {
	if urls == nil {
		urls = strings.Split(getEnv("URLS", "https://paste.rosset.net|https://bin.0xfc.de|https://privatepastebin.com|https://p.darklab.sh"), "|")
		log.Println(urls)
	}
	return urls[rand.Intn(len(urls))]
}

func getDataPath() string {
	if len(dataPath) <= 0 {
		dataPath = getEnv("NOTE_DATA_PATH", "./notes")
		if !strings.HasSuffix(dataPath, "/") {
			dataPath += "/"
		}
	}

	return dataPath
}

func relay(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
		return
	} else if req.Method == http.MethodPost {
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
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Write([]byte(rb))
				log.Print("Paste remote OK:\t", url)
				return
			} else {
				log.Println("Paste remote failed: ", code, url, "\t", rb)
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
		return
	}
}

func back(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		pasteid, ok := req.URL.Query()["pasteid"]
		if !ok || len(pasteid[0]) < 4 {
			http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
			return
		}
		fileName := getDataPath() + "data/" + pasteid[0]
		epoch := fileEpoch(fileName)
		if epoch > 0 {
			fileBytes, err := ioutil.ReadFile(fileName)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Timestamp", strconv.FormatInt(epoch, 10))
			w.Write(fileBytes)
			log.Print("Paste served: ", pasteid)
			return
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
	} else if req.Method == http.MethodPut {
		xhash := req.Header.Get("X-Hash")
		var xTimestamp = req.Header.Get("X-Timestamp")
		pasteid, ok := req.URL.Query()["pasteid"]
		if !ok || len(pasteid[0]) < 4 || len(xhash) < 60 {
			http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
			return
		}
		fileName := getDataPath() + pasteid[0]
		if fileEpoch(fileName) == 0 {
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
		lHash := fmt.Sprintf("%x", bHash[:])
		if xhash != lHash {
			log.Println("Paste hashes: ", xhash, lHash)
			http.Error(w, "Server token missmatch", http.StatusForbidden)
			return
		}
		dataFileName := getDataPath() + "data/" + pasteid[0]
		if len(xTimestamp) == 0 {
			xTimestamp = "0"
		}
		epoch := strconv.FormatInt(fileEpoch(dataFileName), 10)
		if epoch != xTimestamp && epoch != "0" {
			http.Error(w, "Conflict", http.StatusConflict)
			return
		}

		os.WriteFile(dataFileName, rbody, 0600)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Timestamp", strconv.FormatInt(fileEpoch(dataFileName), 10))
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
		log.Println("Paste updated: ", pasteid)
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
		log.Print("Paste registed: ", pasteid)
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
	client := &http.Client{
		Timeout: time.Second * REMOTE_TIMEOUT_SECONDS,
	}
	rreq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", 0
	}
	// rreq.Header.Set("origin", origin_host)
	rreq.Header.Set("X-Requested-With", "JSONHttpRequest")
	rreq.Header.Set("origin", host)
	resp, err := client.Do(rreq)
	if err != nil {
		return err.Error(), http.StatusServiceUnavailable
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

func fileEpoch(fileName string) int64 {
	f, err := os.Stat(fileName)
	if err != nil {
		return 0
	}
	return f.ModTime().Unix()
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
		next = fileEpoch(getDataPath()+pasteid) > 0
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
	folder := getEnv("NOTE_WEB_PATH", "./static")
	port := ":" + getEnv("PORT", "8080")
	if fileEpoch(getDataPath()+"data") == 0 {
		panic("ENV NOTE_DATA_PATH not configured right. Missing folder $NOTE_DATA_PATH/data !")
	}
	if fileEpoch(folder) == 0 || fileEpoch(folder+"/index.html") == 0 {
		panic("Web folder NOTE_WEB_PATH not found !")
	}

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
