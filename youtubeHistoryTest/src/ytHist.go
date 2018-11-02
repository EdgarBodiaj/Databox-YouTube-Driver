package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
	libDatabox "github.com/me-box/lib-go-databox"
)

//default addresses to be used in testing mode
const testArbiterEndpoint = "tcp://127.0.0.1:4444"
const testStoreEndpoint = "tcp://127.0.0.1:5555"

var (
	storeClient *libDatabox.CoreStoreClient
	isRun       = false
)

//Playlist contains the array of videos from the users history feed
type Playlist struct {
	Item []Video `json:"entries"`
}

//Video contains the video metadata after cleanup
type Video struct {
	FullTitle   string   `json:"fulltitle"`
	Title       string   `json:"title"`
	AltTitle    string   `json:"alt_title"`
	Dislikes    int      `json:"dislike_count"`
	Views       int      `json:"view_count"`
	AvgRate     float64  `json:"average_rating"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Track       string   `json:"track"`
	ID          string   `json:"id"`
}

func main() {
	DataboxTestMode := os.Getenv("DATABOX_VERSION") == ""
	registerData(DataboxTestMode)
	//The endpoints and routing for the UI
	router := mux.NewRouter()
	router.HandleFunc("/status", statusEndpoint).Methods("GET")
	router.HandleFunc("/ui/info", infoUser)
	router.HandleFunc("/ui/saved", infoUser)
	router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", http.FileServer(http.Dir("./static"))))
	setUpWebServer(DataboxTestMode, router, "8080")
}

func registerData(testMode bool) {
	//Setup store client
	DataboxStoreEndpoint := "tcp://127.0.0.1:5555"
	if testMode {
		DataboxStoreEndpoint = testStoreEndpoint
		ac, _ := libDatabox.NewArbiterClient("./", "./", testArbiterEndpoint)
		storeClient = libDatabox.NewCoreStoreClient(ac, "./", DataboxStoreEndpoint, false)
		//turn on debug output for the databox library
		libDatabox.OutputDebug(true)
	} else {
		DataboxStoreEndpoint = os.Getenv("DATABOX_ZMQ_ENDPOINT")
		storeClient = libDatabox.NewDefaultCoreStoreClient(DataboxStoreEndpoint)
	}
	//Setup authentication datastore
	authDatasource := libDatabox.DataSourceMetadata{
		Description:    "Youtube Login Data",       //required
		ContentType:    libDatabox.ContentTypeTEXT, //required
		Vendor:         "databox-test",             //required
		DataSourceType: "loginData",                //required
		DataSourceID:   "YoutubeHistoryCred",       //required
		StoreType:      libDatabox.StoreTypeKV,     //required
		IsActuator:     false,
		IsFunc:         false,
	}
	err := storeClient.RegisterDatasource(authDatasource)
	if err != nil {
		libDatabox.Err("Error Registering Credential Datasource " + err.Error())
		return
	}
	libDatabox.Info("Registered Credential Datasource")
	//Setup datastore for main data
	testDatasource := libDatabox.DataSourceMetadata{
		Description:    "Youtube History data",     //required
		ContentType:    libDatabox.ContentTypeJSON, //required
		Vendor:         "databox-test",             //required
		DataSourceType: "videoData",                //required
		DataSourceID:   "YoutubeHistory",           //required
		StoreType:      libDatabox.StoreTypeTSBlob, //required
		IsActuator:     false,
		IsFunc:         false,
	}
	err = storeClient.RegisterDatasource(testDatasource)
	if err != nil {
		libDatabox.Err("Error Registering Datasource " + err.Error())
		return
	}
	libDatabox.Info("Registered Datasource")
}

func infoSaved(w http.ResponseWriter, r *http.Request) {
	tempPas, pErr := storeClient.KVText.Read("YoutubeHistoryCred", "password")
	if pErr != nil {
		fmt.Println(pErr.Error())
		return
	}
	if tempPas != nil {
		libDatabox.Info("Saved auth detected")
		fmt.Fprintf(w, "<h1>Authenticated and Working<h1>")
		go doDriverWork()
	} else {
		fmt.Fprintf(w, "<h1>No saved auth detected<h1>")
		return
	}
}

func infoCheck() (success bool) {
	var (
		temp   Playlist
		cmdErr []byte
		er     error
	)

	cmdName := "youtube-dl"
	tempUse, uErr := storeClient.KVText.Read("YoutubeHistoryCred", "username")
	if uErr != nil {
		fmt.Println(uErr.Error())
		return
	}

	tempPas, pErr := storeClient.KVText.Read("YoutubeHistoryCred", "password")
	if pErr != nil {
		fmt.Println(pErr.Error())
		return
	}

	cmdArgs := []string{("-u " + string(tempUse)), ("-p " + string(tempPas)),
		"--skip-download",
		"--dump-single-json",
		"--playlist-items",
		"1",
		"https://www.youtube.com/feed/history"}

	if cmdErr, er = exec.Command(cmdName, cmdArgs[0], cmdArgs[1], cmdArgs[2], cmdArgs[3], cmdArgs[4], cmdArgs[5], cmdArgs[6]).Output(); er != nil {
		fmt.Println(er.Error())
		return
	}

	err := json.Unmarshal(cmdErr, &temp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if len(temp.Item) > 0 {
		success = true
	} else {
		success = false
	}

	return success
}

func infoUser(w http.ResponseWriter, r *http.Request) {
	libDatabox.Info("Obtained auth")
	r.ParseForm()
	//Obtain user login details for their youtube account
	for k, v := range r.Form {
		if k == "email" {
			err := storeClient.KVText.Write("YoutubeHistoryCred", "username", []byte(strings.Join(v, "")))
			if err != nil {
				libDatabox.Err("Error Write Datasource " + err.Error())
			}

		} else {
			err := storeClient.KVText.Write("YoutubeHistoryCred", "password", []byte(strings.Join(v, "")))
			if err != nil {
				libDatabox.Err("Error Write Datasource " + err.Error())
			}
		}

	}
	//If the driver is already running, do not create a new instance
	if isRun {
		fmt.Fprintf(w, "<h1>Already running<h1>")
		libDatabox.Info("Already running")
		fmt.Fprintf(w, " <button onclick='goBack()'>Go Back</button><script>function goBack() {	window.history.back();}</script> ")
	} else {

		if infoCheck() == true {
			fmt.Fprintf(w, "<h1>Authenticated and Working<h1>")
			go doDriverWork()
		} else {
			fmt.Fprintf(w, "<h1>Bad Auth<h1>")
			fmt.Fprintf(w, " <button onclick='goBack()'>Go Back</button><script>function goBack() {	window.history.back();}</script> ")
		}

	}

}

func doDriverWork() {
	libDatabox.Info("starting doDriverWork")
	isRun = true

	cmdName := "youtube-dl"
	tempUse, uErr := storeClient.KVText.Read("YoutubeHistoryCred", "username")
	if uErr != nil {
		fmt.Println(uErr.Error())
		return
	}

	tempPas, pErr := storeClient.KVText.Read("YoutubeHistoryCred", "password")
	if pErr != nil {
		fmt.Println(pErr.Error())
		return
	}

	cmdArgs := []string{("-u " + string(tempUse)), ("-p " + string(tempPas)),
		"--skip-download",
		"--dump-single-json",
		"--playlist-items",
		"1-10",
		"https://www.youtube.com/feed/history"}

	//Create recent store, error object and output
	var (
		hOld   Playlist
		er     error
		cmdOut []byte
	)
	for {
		//Create new var for incoming data
		var hNew Playlist
		if cmdOut, er = exec.Command(cmdName, cmdArgs[0], cmdArgs[1], cmdArgs[2], cmdArgs[3], cmdArgs[4], cmdArgs[5], cmdArgs[6]).Output(); er != nil {
			fmt.Println(er.Error())
			return
		}
		libDatabox.Info("Download Finished")
		err := json.Unmarshal(cmdOut, &hNew)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(hNew.Item) == 0 {
			isRun = false
			break
		}

		//Check to see if the recent store is populated
		//If it has been populated, compare new items with the stored items
		if hOld.Item != nil {
			fmt.Println("New first item is: " + hNew.Item[0].Title)
			fmt.Println("Old first item is: " + hOld.Item[0].Title)
			for i := 0; i < len(hNew.Item); i++ {
				for j := 0; j < len(hOld.Item); j++ {
					//If a duplicate is found in the recent store, do not save item
					if hNew.Item[i].ID == hOld.Item[j].ID {
						break
					}
					//If no duplicates have been found in the store, save the item
					if j == len(hOld.Item)-1 {
						temp, tErr := json.Marshal(hNew.Item[i])
						if tErr != nil {
							fmt.Println(tErr.Error())
							return
						}
						aerr := storeClient.TSBlobJSON.Write("YoutubeHistory", temp)
						if aerr != nil {
							libDatabox.Err("Error Write Datasource " + aerr.Error())
						}
						//libDatabox.Info("Data written to store: " + string(temp))
						libDatabox.Info("Storing data")
					}
				}
			}
			//If its the first time the driver has been run, the recent store will be empty
			//Therefore store the current playlist items
		} else {
			fmt.Println("First case")
			for i := 0; i < len(hNew.Item); i++ {
				temp, tErr := json.Marshal(hNew.Item[i])
				if tErr != nil {
					fmt.Println("" + tErr.Error())
					return
				}
				fmt.Println(string(temp))
				libDatabox.Info(string(temp))
				aerr := storeClient.TSBlobJSON.Write("YoutubeHistory", temp)
				if aerr != nil {
					libDatabox.Err("Error Write Datasource " + aerr.Error())
				}
				//libDatabox.Info("Data written to store: " + string(temp))
				libDatabox.Info("Storing data")
			}
		}

		hOld = hNew

		time.Sleep(time.Second * 30)
		fmt.Println("New Cycle")
	}
}

func statusEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("active\n"))
}

func setUpWebServer(testMode bool, r *mux.Router, port string) {

	//Start up a well behaved HTTP/S server for displying the UI

	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      r,
	}
	if testMode {
		//set up an http server for testing
		libDatabox.Info("Waiting for http requests on port http://127.0.0.1" + srv.Addr + "/ui ....")
		log.Fatal(srv.ListenAndServe())
	} else {
		//configure tls
		tlsConfig := &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
			},
		}

		srv.TLSConfig = tlsConfig

		libDatabox.Info("Waiting for https requests on port " + srv.Addr + " ....")
		log.Fatal(srv.ListenAndServeTLS(libDatabox.GetHttpsCredentials(), libDatabox.GetHttpsCredentials()))
	}
}
