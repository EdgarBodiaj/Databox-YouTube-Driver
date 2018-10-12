package main


import (
	"fmt"
	"net/http"
	"log"
	"time"
	"crypto/tls"
	"github.com/gorilla/mux"
	"os"
	"os/exec"
	"strings"
	"encoding/json"
	libDatabox "github.com/me-box/lib-go-databox"
)

//default addresses to be used in testing mode
const testArbiterEndpoint = "tcp://127.0.0.1:4444"
const testStoreEndpoint = "tcp://127.0.0.1:5555"

var (
	cmdOut []byte
	er error
	History Playlist
	Indiv Video
	username string
	password string
)


type Playlist struct{
	Item []Video `json:"entries"`
} 

type Video struct{
	FullTitle string `json:"fulltitle"`
	Title string `json:"title"`
	AltTitle string `json:"alt_title"`
	Dislikes int `json:"dislike_count"`
	Views int `json:"view_count"`
	AvgRate float64 `json:"average_rating"`
	Description string `json:"description"`
	Tags []string `json:"tags"`
	Track string `json:"track"`
	ID string `json:"id"`
	
}

func main(){
	libDatabox.Info("Starting ....")

	//Are we running inside databox?
	DataboxTestMode := os.Getenv("DATABOX_VERSION") == ""

	// Read in the store endpoint provided by databox
	// this is a driver so you will get a core-store
	// and you are responsible for registering datasources
	// and writing in data.
	var DataboxStoreEndpoint string
	var storeClient *libDatabox.CoreStoreClient
	httpServerPort := "8080"
	if DataboxTestMode {
		DataboxStoreEndpoint = testStoreEndpoint
		ac, _ := libDatabox.NewArbiterClient("./", "./", testArbiterEndpoint)
		storeClient = libDatabox.NewCoreStoreClient(ac, "./", DataboxStoreEndpoint, false)
		//turn on debug output for the databox library
		libDatabox.OutputDebug(true)
	} else {
		DataboxStoreEndpoint = os.Getenv("DATABOX_STORE_ENDPOINT")
		storeClient = libDatabox.NewDefaultCoreStoreClient(DataboxStoreEndpoint)
	}

	if checkInstall(){go doDriverWork(DataboxTestMode, storeClient)}
	
	//The endpoints and routing for the UI
	router := mux.NewRouter()
	router.HandleFunc("/status", statusEndpoint).Methods("GET")
	router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", http.FileServer(http.Dir("./htmlPage"))))
	router.HandleFunc("/info", infoUser)
	setUpWebServer(true, router, httpServerPort)
}

func infoUser(w http.ResponseWriter, r *http.Request, ) {
	r.ParseForm()
	//Obtain user login details for their youtube account
	for k, v := range r.Form {
        	if k == "email" {username = strings.Join(v, "")
		} else {password = strings.Join(v, "")}
        	
    	}
	
}

func checkInstall()(exist bool){
	//Check to see if youtube-dl is installed
	if _, err := os.Stat("/usr/local/bin/youtube-dl"); os.IsNotExist(err) {
		fmt.Println("youtube-dl does not exist")
		os.Exit(3)
		return false
	} else {return true}
}

func doDriverWork(testMode bool, storeClient *libDatabox.CoreStoreClient) {
	libDatabox.Info("starting doDriverWork")

	//register our datasources
	//we only need to do this once at start up
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
	arr := storeClient.RegisterDatasource(testDatasource)
	if arr != nil {
		libDatabox.Err("Error Registering Datasource " + arr.Error())
		return
	}
	libDatabox.Info("Registered Datasource")

	cmdName :="youtube-dl"
	
	tempUse := ""
	tempPas := ""
	libDatabox.Info("Waiting for authentication")
	for {   tempPas = "-p " + password 
		tempUse = "-u " + username
		if tempPas != "-p " && tempUse != "-u "{break}}

	cmdArgs := []string{tempUse,tempPas,
				"--skip-download",
				"-o'%(playlist)s/%(playlist_index)s - %(title)s.%(ext)s'",
				"--dump-single-json",
				"--playlist-items",
				"1-3",
				"https://www.youtube.com/feed/history"}
	if cmdOut, er = exec.Command(cmdName, cmdArgs[0], cmdArgs[1], cmdArgs[2], cmdArgs[3], cmdArgs[4], cmdArgs[5], cmdArgs[6], cmdArgs[7]).Output(); er != nil {log.Fatal(er)}
	//fmt.Println(string(cmdOut))

	err := json.Unmarshal(cmdOut, &History)
	if err != nil {fmt.Println(err)}
	
	temp, tErr := json.Marshal(History)
	if tErr != nil{log.Fatal(tErr)}
	libDatabox.Info("Converting data")
	aerr := storeClient.TSBlobJSON.Write("YoutubeHistory", temp)
		if aerr != nil {
			libDatabox.Err("Error Write Datasource " + aerr.Error())
		}
		libDatabox.Info("Data written to store: " + string(temp))	
	libDatabox.Info("Storing data")
	/*fmt.Println(History.Item[0].Title)
	fmt.Println(History.Item[0].Tags)
	fmt.Println(History.Item[0].ID)
	fmt.Println(History.Item[0].Views)
	
	fmt.Println(History.Item[1].Title)
	fmt.Println(History.Item[1].Tags)
	fmt.Println(History.Item[1].ID)
	fmt.Println(History.Item[1].Views)

	fmt.Println(History.Item[2].Title)
	fmt.Println(History.Item[2].Tags)
	fmt.Println(History.Item[2].ID)
	fmt.Println(History.Item[2].Views)*/
		
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
