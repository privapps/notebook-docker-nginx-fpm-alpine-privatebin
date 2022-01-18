package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
)

var urls []string

func getUrl() string {
	if urls == nil {
		urls = strings.Split(os.Getenv("URLS"), "|")
		log.Println(urls)
	}
	return urls[rand.Intn(len(urls))]
}

func post(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Bad request - Go away!", http.StatusBadRequest)
		return
	}
	for i := 0; i < 10; i++ {
		url := getUrl()
		rb, code := doPost(url, body)
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
func getHost(url string) string {
	fp := strings.Index(url, "//")
	ep := strings.Index(url[fp+3:], "/")
	if ep < 0 {
		return url[fp+3:]
	} else {
		return url[fp+3 : fp+ep+3]
	}
}
func doPost(url string, body []byte) (string, int) {
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

func main() {
	folder := "./static"
	port := ":8080"
	fs := http.FileServer(http.Dir(folder))
	http.HandleFunc("/relay.php", post)
	http.Handle("/", fs)

	log.Println("Listening on " + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}

}
