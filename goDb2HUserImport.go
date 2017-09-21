package main

//----- Packages -----
import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"text/template"
	/* DAV inclusion */
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"

	"crypto/rand"
	"github.com/hornbill/color" //-- CLI Colour
	"github.com/hornbill/goApiLib"
	"github.com/hornbill/pb" //--Hornbil Clone of "github.com/cheggaaa/pb"
	"strconv"
	"strings"
	"sync"
	"time"
	//SQL Package
	"github.com/hornbill/sqlx"
	//SQL Drivers
	_ "github.com/alexbrainman/odbc"
	_ "github.com/hornbill/go-mssqldb"
	_ "github.com/hornbill/mysql"
	_ "github.com/jnewmano/mysql320" //MySQL v3.2.0 to v5 driver - Provides SWSQL (MySQL 4.0.16) support
)

//----- Constants -----
const (
	letterBytes  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	version      = "1.2.2"
	constOK      = "ok"
	updateString = "Update"
	createString = "Create"
)

var (
	SQLImportConf       SQLImportConfStruct
	xmlmcInstanceConfig xmlmcConfig
	sites               []siteListStruct
	managers            []managerListStruct
	groups              []groupListStruct
	counters            counterTypeStruct
	configFileName      string
	configZone          string
	configLogPrefix     string
	configDryRun        bool
	configVersion       bool
	configWorkers       int
	configMaxRoutines   string
	timeNow             string
	startTime           time.Time
	endTime             time.Duration
	errorCount          uint64
	noValuesToUpdate    = "There are no values to update"
	mutexBar            = &sync.Mutex{}
	mutexCounters       = &sync.Mutex{}
	mutexSites          = &sync.Mutex{}
	mutexGroups         = &sync.Mutex{}
	mutexManagers       = &sync.Mutex{}
	maxGoroutines       = 6
	loggerApi           *apiLib.XmlmcInstStruct
	once                sync.Once
	onceLog             sync.Once
	mutexLogger         = &sync.Mutex{}
	mutexLog            = &sync.Mutex{}
	f                   *os.File
	client              = http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1,
		},
		Timeout: time.Duration(10 * time.Second),
	}

	userProfileArray = []string{
		"MiddleName",
		"JobDescription",
		"Manager",
		"WorkPhone",
		"Qualifications",
		"Interests",
		"Expertise",
		"Gender",
		"Dob",
		"Nationality",
		"Religion",
		"HomeTelephone",
		"SocialNetworkA",
		"SocialNetworkB",
		"SocialNetworkC",
		"SocialNetworkD",
		"SocialNetworkE",
		"SocialNetworkF",
		"SocialNetworkG",
		"SocialNetworkH",
		"PersonalInterests",
		"HomeAddress",
		"PersonalBlog",
		"Attrib1",
		"Attrib2",
		"Attrib3",
		"Attrib4",
		"Attrib5",
		"Attrib6",
		"Attrib7",
		"Attrib8"}

	userUpdateArray = []string{
		"userId",
		"UserType",
		"Name",
		"Password",
		"FirstName",
		"LastName",
		"JobTitle",
		"Site",
		"Phone",
		"Email",
		"Mobile",
		"AbsenceMessage",
		"TimeZone",
		"Language",
		"DateTimeFormat",
		"DateFormat",
		"TimeFormat",
		"CurrencySymbol",
		"CountryCode"}

	userCreateArray = []string{
		"userId",
		"Name",
		"Password",
		"UserType",
		"FirstName",
		"LastName",
		"JobTitle",
		"Site",
		"Phone",
		"Email",
		"Mobile",
		"AbsenceMessage",
		"TimeZone",
		"Language",
		"DateTimeFormat",
		"DateFormat",
		"TimeFormat",
		"CurrencySymbol",
		"CountryCode"}
)

type siteListStruct struct {
	SiteName string
	SiteID   int
}
type xmlmcSiteListResponse struct {
	MethodResult string               `xml:"status,attr"`
	Params       paramsSiteListStruct `xml:"params"`
	State        stateStruct          `xml:"state"`
}
type paramsSiteListStruct struct {
	RowData paramsSiteRowDataListStruct `xml:"rowData"`
}
type paramsSiteRowDataListStruct struct {
	Row siteObjectStruct `xml:"row"`
}
type siteObjectStruct struct {
	SiteID      int    `xml:"h_id"`
	SiteName    string `xml:"h_site_name"`
	SiteCountry string `xml:"h_country"`
}

type managerListStruct struct {
	UserName string
	UserID   string
}
type groupListStruct struct {
	GroupName string
	GroupID   string
}

type xmlmcConfig struct {
	instance string
	zone     string
	url      string
}

type counterTypeStruct struct {
	updated        uint16
	created        uint16
	profileUpdated uint16
	updatedSkipped uint16
	createskipped  uint16
	profileSkipped uint16
}
type userAccountStatusStruct struct {
	Action  string
	Enabled bool
	Status  string
}
type userManagerStruct struct {
	Action  string
	Enabled bool
}

type siteLookupStruct struct {
	Action    string
	Enabled   bool
	Attribute string
}
type imageLinkStruct struct {
	Action     string
	Enabled    bool
	UploadType string
	ImageType  string
	URI        string
}
type orgLookupStruct struct {
	Action   string
	Enabled  bool
	OrgUnits []OrgUnitStruct
}
type OrgUnitStruct struct {
	Attribute   string
	Type        int
	Membership  string
	TasksView   bool
	TasksAction bool
}
type SQLImportConfStruct struct {
	APIKey             string
	InstanceID         string
	URL                string
	DAVURL             string
	UpdateUserType     bool
	UserRoleAction     string
	UserIdentifier     string
	SQLConf            sqlConfStruct
	UserMapping        map[string]string //userMappingStruct
	UserAccountStatus  userAccountStatusStruct
	UserProfileMapping map[string]string //userProfileMappingStruct
	UserManagerMapping userManagerStruct
	SQLAttributes      []string
	Roles              []string
	SiteLookup         siteLookupStruct
	ImageLink          imageLinkStruct
	OrgLookup          orgLookupStruct
}
type xmlmcResponse struct {
	MethodResult string       `xml:"status,attr"`
	Params       paramsStruct `xml:"params"`
	State        stateStruct  `xml:"state"`
}
type xmlmcCheckUserResponse struct {
	MethodResult string                 `xml:"status,attr"`
	Params       paramsCheckUsersStruct `xml:"params"`
	State        stateStruct            `xml:"state"`
}
type xmlmcUserListResponse struct {
	MethodResult string                     `xml:"status,attr"`
	Params       paramsUserSearchListStruct `xml:"params"`
	State        stateStruct                `xml:"state"`
}
type paramsUserSearchListStruct struct {
	RowData paramsUserRowDataListStruct `xml:"rowData"`
}
type paramsUserRowDataListStruct struct {
	Row userObjectStruct `xml:"row"`
}
type userObjectStruct struct {
	UserID   string `xml:"h_user_id"`
	UserName string `xml:"h_name"`
}

type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}
type paramsCheckUsersStruct struct {
	RecordExist bool `xml:"recordExist"`
}
type paramsStruct struct {
	SessionID string `xml:"sessionId"`
}

//###
type sqlConfStruct struct {
	Driver   string
	Server   string
	UserName string
	Password string
	Port     int
	Query    string
	Database string
	Encrypt  bool
	UserID   string
}

//### organisation units structures
type xmlmcuserSetGroupOptionsResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
}
type xmlmcprofileSetImageResponse struct {
	MethodResult string                `xml:"status,attr"`
	Params       paramsGroupListStruct `xml:"params"`
	State        stateStruct           `xml:"state"`
}
type xmlmcGroupListResponse struct {
	MethodResult string                `xml:"status,attr"`
	Params       paramsGroupListStruct `xml:"params"`
	State        stateStruct           `xml:"state"`
}

type paramsGroupListStruct struct {
	RowData paramsGroupRowDataListStruct `xml:"rowData"`
}

type paramsGroupRowDataListStruct struct {
	Row groupObjectStruct `xml:"row"`
}

type groupObjectStruct struct {
	GroupID   string `xml:"h_id"`
	GroupName string `xml:"h_name"`
}

func initVars() {
	//-- Start Time for Durration
	startTime = time.Now()
	//-- Start Time for Log File
	timeNow = time.Now().Format(time.RFC3339)
	//-- Remove :
	timeNow = strings.Replace(timeNow, ":", "-", -1)
	//-- Set Counter
	errorCount = 0
}

//----- Main Function -----
func main() {

	//-- Initiate Variables
	initVars()

	//-- Process Flags
	procFlags()

	//-- If configVersion just output version number and die
	if configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//-- Load Configuration File Into Struct
	SQLImportConf = loadConfig()

	//-- Validation on Configuration File
	err := validateConf()
	if err != nil {
		logger(4, fmt.Sprintf("%v", err), true)
		logger(4, "Please Check your Configuration File: "+configFileName, true)
		return
	}

	//-- Set Instance ID
	var boolSetInstance = setInstance(configZone, SQLImportConf.InstanceID)
	if boolSetInstance != true {
		return
	}

	//-- Generate Instance XMLMC Endpoint
	SQLImportConf.URL = getInstanceURL()
	SQLImportConf.DAVURL = getInstanceDAVURL()
	logger(1, "Instance Endpoint "+fmt.Sprintf("%v", SQLImportConf.URL), true)
	//-- Once we have loaded the config write to hornbill log file
	logged := espLogger("---- XMLMC SQL Import Utility V"+fmt.Sprintf("%v", version)+" ----", "debug")

	if !logged {
		logger(4, "Unable to Connect to Instance", true)
		return
	}

	//Set SWSQLDriver to mysql320
	if SQLImportConf.SQLConf.Driver == "swsql" {
		SQLImportConf.SQLConf.Driver = "mysql320"
	}

	//Get asset types, process accordingly
	var boolSQLUsers, arrUsers = queryDatabase()
	if boolSQLUsers {
		processUsers(arrUsers)
	} else {
		logger(4, "No Results found", true)
		return
	}

	outputEnd()
}

func outputEnd() {
	//-- End output
	if errorCount > 0 {
		logger(4, "Error encountered please check the log file", true)
		logger(4, "Error Count: "+fmt.Sprintf("%d", errorCount), true)
		//logger(4, "Check Log File for Details", true)
	}
	logger(1, "Updated: "+fmt.Sprintf("%d", counters.updated), true)
	logger(1, "Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), true)

	logger(1, "Created: "+fmt.Sprintf("%d", counters.created), true)
	logger(1, "Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), true)

	logger(1, "Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), true)
	logger(1, "Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), true)

	//-- Show Time Takens
	endTime = time.Since(startTime)
	logger(1, "Time Taken: "+fmt.Sprintf("%v", endTime), true)
	//-- complete
	complete()
	logger(1, "---- XMLMC SQL Import Complete ---- ", true)
}
func procFlags() {
	//-- Grab Flags
	flag.StringVar(&configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&configZone, "zone", "eur", "Override the default Zone the instance sits in")
	flag.StringVar(&configLogPrefix, "logprefix", "", "Add prefix to the logfile")
	flag.BoolVar(&configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating users")
	flag.BoolVar(&configVersion, "version", false, "Output Version")
	flag.IntVar(&configWorkers, "workers", 1, "Number of Worker threads to use")
	flag.StringVar(&configMaxRoutines, "concurrent", "1", "Maximum number of requests to import concurrently.")

	//-- Parse Flags
	flag.Parse()

	//-- Output config
	if !configVersion {
		outputFlags()
	}

	//Check maxGoroutines for valid value
	maxRoutines, err := strconv.Atoi(configMaxRoutines)
	if err != nil {
		color.Red("Unable to convert maximum concurrency of [" + configMaxRoutines + "] to type INT for processing")
		return
	}
	maxGoroutines = maxRoutines

	if maxGoroutines < 1 || maxGoroutines > 10 {
		color.Red("The maximum concurrent requests allowed is between 1 and 10 (inclusive).\n\n")
		color.Red("You have selected " + configMaxRoutines + ". Please try again, with a valid value against ")
		color.Red("the -concurrent switch.")
		return
	}
}
func outputFlags() {
	//-- Output
	logger(1, "---- XMLMC SQL Import Utility V"+fmt.Sprintf("%v", version)+" ----", true)

	logger(1, "Flag - Config File "+configFileName, true)
	logger(1, "Flag - Zone "+configZone, true)
	logger(1, "Flag - Log Prefix "+configLogPrefix, true)
	logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", configDryRun), true)
	logger(1, "Flag - Workers "+fmt.Sprintf("%v", configWorkers), false)
}

//-- Check Latest
//-- Function to Load Configruation File
func loadConfig() SQLImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + configFileName
	logger(1, "Loading Config File: "+configurationFilePath, false)
	if _, fileCheckErr := os.Stat(configurationFilePath); os.IsNotExist(fileCheckErr) {
		logger(4, "No Configuration File", true)
		os.Exit(102)
	}

	//-- Load Config File
	file, fileError := os.Open(configurationFilePath)
	//-- Check For Error Reading File
	if fileError != nil {
		logger(4, "Error Opening Configuration File: "+fmt.Sprintf("%v", fileError), true)
	}
	//-- New Decoder
	decoder := json.NewDecoder(file)

	eldapConf := SQLImportConfStruct{}

	//-- Decode JSON
	err := decoder.Decode(&eldapConf)
	//-- Error Checking
	if err != nil {
		logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true)
	}

	//-- Return New Congfig
	return eldapConf
}

func validateConf() error {

	//-- Check for API Key
	if SQLImportConf.APIKey == "" {
		err := errors.New("API Key is not set")
		return err
	}
	//-- Check for Instance ID
	if SQLImportConf.InstanceID == "" {
		err := errors.New("InstanceID is not set")
		return err
	}

	//-- Process Config File

	return nil
}

//-- Worker Pool Function
func loggerGen(t int, s string) string {

	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
	case 1:
		errorLogPrefix = "[DEBUG] "
	case 2:
		errorLogPrefix = "[MESSAGE] "
	case 3:
		errorLogPrefix = "[WARN] "
	case 4:
		errorLogPrefix = "[ERROR] "
	}
	currentTime := time.Now().UTC()
	time := currentTime.Format("2006/01/02 15:04:05")
	return time + " " + errorLogPrefix + s + "\n"
}

//-- Logging function
func logger(t int, s string, outputtoCLI bool) {

	mutexLog.Lock()
	defer mutexLog.Unlock()

	onceLog.Do(func() {
		//-- Curreny WD
		cwd, _ := os.Getwd()
		//-- Log Folder
		logPath := cwd + "/log"
		//-- Log File
		logFileName := logPath + "/" + configLogPrefix + "SQL_User_Import_" + timeNow + ".log"
		//red := color.New(color.FgRed).PrintfFunc()
		//orange := color.New(color.FgCyan).PrintfFunc()
		//-- If Folder Does Not Exist then create it
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			err := os.Mkdir(logPath, 0777)
			if err != nil {
				fmt.Printf("Error Creating Log Folder %q: %s \r", logPath, err)
				os.Exit(101)
			}
		}

		//-- Open Log File
		var err error
		f, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			fmt.Printf("Error Creating Log File %q: %s \n", logFileName, err)
			os.Exit(100)
		}
		// assign it to the standard logger
		log.SetOutput(f)
	})
	// don't forget to close it
	//defer f.Close()
	red := color.New(color.FgRed).PrintfFunc()
	orange := color.New(color.FgCyan).PrintfFunc()
	var errorLogPrefix = ""
	//-- Create Log Entry
	switch t {
	case 0:
		errorLogPrefix = ""
	case 1:
		errorLogPrefix = "[DEBUG] "
	case 2:
		errorLogPrefix = "[MESSAGE] "
	case 3:
		errorLogPrefix = "[WARN] "
	case 4:
		errorLogPrefix = "[ERROR] "
	}
	if outputtoCLI {
		if t == 3 {
			orange(errorLogPrefix + s + "\n")
		} else if t == 4 {
			red(errorLogPrefix + s + "\n")
		} else {
			fmt.Printf(errorLogPrefix + s + "\n")
		}

	}
	log.Println(errorLogPrefix + s)
}

//-- complete
func complete() {
	//-- End output
	espLogger("Errors: "+fmt.Sprintf("%d", errorCount), "error")
	espLogger("Updated: "+fmt.Sprintf("%d", counters.updated), "debug")
	espLogger("Updated Skipped: "+fmt.Sprintf("%d", counters.updatedSkipped), "debug")
	espLogger("Created: "+fmt.Sprintf("%d", counters.created), "debug")
	espLogger("Created Skipped: "+fmt.Sprintf("%d", counters.createskipped), "debug")
	espLogger("Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), "debug")
	espLogger("Profiles Skipped: "+fmt.Sprintf("%d", counters.profileSkipped), "debug")
	espLogger("Time Taken: "+fmt.Sprintf("%v", endTime), "debug")
	espLogger("---- XMLMC SQL User Import Complete ---- ", "debug")
}

// Set Instance Id
func setInstance(strZone string, instanceID string) bool {
	//-- Set Zone
	setZone(strZone)
	//-- Check for blank instance
	if instanceID == "" {
		logger(4, "InstanceId Must be Specified in the Configuration File", true)
		return false
	}
	//-- Set Instance
	xmlmcInstanceConfig.instance = instanceID
	return true
}

// Set Instance Zone to Overide Live
func setZone(zone string) {
	xmlmcInstanceConfig.zone = zone
}

//-- Log to ESP
func espLogger(message string, severity string) bool {

	// We lock the whole function so we dont reuse the same connection for multiple logging attempts
	mutexLogger.Lock()
	defer mutexLogger.Unlock()

	// We initilaise the connection pool the first time the function is called and reuse it
	// This is reuse the connections rather than creating a pool each invocation
	once.Do(func() {
		loggerApi = apiLib.NewXmlmcInstance(SQLImportConf.URL)
	})
	loggerApi.SetAPIKey(SQLImportConf.APIKey)
	//We set a 5 second timeout as anything else calling espLogger will be locked waiting for completion so dont wait too long
	loggerApi.SetTimeout(5)
	loggerApi.SetParam("fileName", "SQL_User_Import")
	loggerApi.SetParam("group", "general")
	loggerApi.SetParam("severity", severity)
	loggerApi.SetParam("message", message)

	XMLLogger, xmlmcErr := loggerApi.Invoke("system", "logMessage")
	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", xmlmcErr), true)
		return false
	}
	err := xml.Unmarshal([]byte(XMLLogger), &xmlRespon)
	if err != nil {
		logger(4, "Unable to write to log "+fmt.Sprintf("%s", err), true)
		return false
	}
	if xmlRespon.MethodResult != constOK {
		logger(4, "Unable to write to log "+xmlRespon.State.ErrorRet, true)
		return false
	}

	return true
}

//-- Function Builds XMLMC End Point
func getInstanceURL() string {
	xmlmcInstanceConfig.url = "https://"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.zone
	xmlmcInstanceConfig.url += "api.hornbill.com/"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.instance
	xmlmcInstanceConfig.url += "/xmlmc/"

	return xmlmcInstanceConfig.url
}

//-- Function Builds XMLMC End Point
func getInstanceDAVURL() string {
	xmlmcInstanceConfig.url = "https://"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.zone
	xmlmcInstanceConfig.url += "api.hornbill.com/"
	xmlmcInstanceConfig.url += xmlmcInstanceConfig.instance
	xmlmcInstanceConfig.url += "/dav/"

	return xmlmcInstanceConfig.url
}

//buildConnectionString -- Build the connection string for the SQL driver
func buildConnectionString() string {
	//	if SQLImportConf.SQLConf.Server == "" || SQLImportConf.SQLConf.Database == "" || SQLImportConf.SQLConf.UserName == "" || SQLImportConf.SQLConf.Password == "" {
	if SQLImportConf.SQLConf.Server == "" || SQLImportConf.SQLConf.Database == "" || SQLImportConf.SQLConf.UserName == "" {
		//Conf not set - log error and return empty string
		logger(4, "Database configuration not set.", true)
		return ""
	}
	logger(1, "Connecting to Database Server: "+SQLImportConf.SQLConf.Server, true)
	connectString := ""
	switch SQLImportConf.SQLConf.Driver {

	case "mssql":
		connectString = "server=" + SQLImportConf.SQLConf.Server
		connectString = connectString + ";database=" + SQLImportConf.SQLConf.Database
		connectString = connectString + ";user id=" + SQLImportConf.SQLConf.UserName
		connectString = connectString + ";password=" + SQLImportConf.SQLConf.Password
		if SQLImportConf.SQLConf.Encrypt == false {
			connectString = connectString + ";encrypt=disable"
		}
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting := strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + ";port=" + dbPortSetting
		}

	case "mysql":
		connectString = SQLImportConf.SQLConf.UserName + ":" + SQLImportConf.SQLConf.Password
		connectString = connectString + "@tcp(" + SQLImportConf.SQLConf.Server + ":"
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting := strconv.Itoa(SQLImportConf.SQLConf.Port)
			connectString = connectString + dbPortSetting
		} else {
			connectString = connectString + "3306"
		}
		connectString = connectString + ")/" + SQLImportConf.SQLConf.Database

	case "mysql320":
		var dbPortSetting string
		if SQLImportConf.SQLConf.Port != 0 {
			dbPortSetting = strconv.Itoa(SQLImportConf.SQLConf.Port)
		} else {
			dbPortSetting = "3306"
		}
		connectString = "tcp:" + SQLImportConf.SQLConf.Server + ":" + dbPortSetting
		connectString = connectString + "*" + SQLImportConf.SQLConf.Database + "/" + SQLImportConf.SQLConf.UserName + "/" + SQLImportConf.SQLConf.Password
	case "csv":
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";Extended Properties='text;HDR=Yes;FMT=Delimited'"
		SQLImportConf.SQLConf.Driver = "odbc"
	case "excel":
		connectString = "DSN=" + SQLImportConf.SQLConf.Database + ";"
		SQLImportConf.SQLConf.Driver = "odbc"

	}

	return connectString
}

//queryDatabase -- Query Asset Database for assets of current type
//-- Builds map of assets, returns true if successful
func queryDatabase() (bool, []map[string]interface{}) {
	//Clear existing Asset Map down
	ArrUserMaps := make([]map[string]interface{}, 0)
	connString := buildConnectionString()
	if connString == "" {
		return false, ArrUserMaps
	}
	//Connect to the JSON specified DB
	db, err := sqlx.Open(SQLImportConf.SQLConf.Driver, connString)
	if err != nil {
		logger(4, " [DATABASE] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrUserMaps
	}
	defer db.Close()
	//Check connection is open
	err = db.Ping()
	if err != nil {
		logger(4, " [DATABASE] [PING] Database Connection Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrUserMaps
	}
	logger(3, "[DATABASE] Connection Successful", true)
	logger(3, "[DATABASE] Running database query for Customers. Please wait...", true)
	//build query
	sqlQuery := SQLImportConf.SQLConf.Query //BaseSQLQuery
	logger(3, "[DATABASE] Query:"+sqlQuery, false)
	//Run Query
	rows, err := db.Queryx(sqlQuery)
	if err != nil {
		logger(4, " [DATABASE] Database Query Error: "+fmt.Sprintf("%v", err), true)
		return false, ArrUserMaps
	}

	//Build map full of assets
	intUserCount := 0
	for rows.Next() {
		intUserCount++
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			//We are going to skip this record as it did not scan properly
			logger(4, " [DATABASE] Database Scan Error: "+fmt.Sprintf("%v", err), true)
			continue
		}
		//Stick marshalled data map in to parent slice
		ArrUserMaps = append(ArrUserMaps, results)
	}
	defer rows.Close()
	logger(3, fmt.Sprintf("[DATABASE] Found %d results", intUserCount), false)
	return true, ArrUserMaps
}

//processAssets -- Processes Assets from Asset Map
//--If asset already exists on the instance, update
//--If asset doesn't exist, create
func processUsers(arrUsers []map[string]interface{}) {
	bar := pb.StartNew(len(arrUsers))
	logger(1, "Processing Users", false)

	total := len(arrUsers)
	jobs := make(chan map[string]interface{}, total)
	results := make(chan int, total)
	workers := maxGoroutines

	if total < workers {
		workers = total
	}
	//This starts up 3 workers, initially blocked because there are no jobs yet.
	for w := 1; w <= workers; w++ {
		go ProcessUserWorkers(jobs, results, bar)
	}

	//-- Here we send a job for each user we have to process
	for _, usersMaps := range arrUsers {
		jobs <- usersMaps
	}
	close(jobs)
	//-- Finally we collect all the results of the work.
	for a := 1; a <= workers; a++ {
		<-results
	}

	bar.FinishPrint("Processing Complete!")
}

func ProcessUserWorkers(jobs <-chan map[string]interface{}, results chan<- int, bar *pb.ProgressBar) {

	//We should create the APi connections here and pass the reference around
	espXmlmc := apiLib.NewXmlmcInstance(SQLImportConf.URL)
	espXmlmc.SetAPIKey(SQLImportConf.APIKey)

	espXmlmcLookup := apiLib.NewXmlmcInstance(SQLImportConf.URL)
	espXmlmcLookup.SetAPIKey(SQLImportConf.APIKey)

	for customerRecord := range jobs {
		userMap := customerRecord
		userIDField := fmt.Sprintf("%v", SQLImportConf.SQLConf.UserID)
		//Get the asset ID for the current record
		userID := fmt.Sprintf("%s", userMap[userIDField])
		logger(1, "User ID: "+userID, false)
		if userID == "" {
			//No userId so skip to the next
			continue
		}
		//Increment the bar
		mutexBar.Lock()
		bar.Increment()
		mutexBar.Unlock()

		var boolUpdate = false
		var isErr = false
		boolUpdate, err := checkUserOnInstance(userID, espXmlmc)
		if err != nil {
			logger(4, "Unable to Search For User: "+fmt.Sprintf("%v", err), true)
			isErr = true
			continue
		}
		//-- Update or Create Asset
		if !isErr {
			if boolUpdate {
				logger(1, "Update Customer: "+userID, false)
				_, errUpdate := updateUser(userMap, espXmlmc, espXmlmcLookup)
				if errUpdate != nil {
					logger(4, "Unable to Update User: "+fmt.Sprintf("%v", errUpdate), false)
				}
			} else {
				logger(1, "Create Customer: "+userID, false)
				_, errorCreate := createUser(userMap, espXmlmc, espXmlmcLookup)
				if errorCreate != nil {
					logger(4, "Unable to Create User: "+fmt.Sprintf("%v", errorCreate), false)
				}
			}
		}
	}
	//Send an int down the channel to say we are exited
	results <- 0

}

func updateUser(u map[string]interface{}, espXmlmc *apiLib.XmlmcInstStruct, espXmlmcLookup *apiLib.XmlmcInstStruct) (bool, error) {
	buf2 := bytes.NewBufferString("")
	//-- Do we Lookup Site
	p := make(map[string]string)
	for key, value := range u {
		p[key] = fmt.Sprintf("%s", value)
	}
	userID := p[SQLImportConf.SQLConf.UserID]
	for key := range userUpdateArray {
		field := userUpdateArray[key]
		value := SQLImportConf.UserMapping[field] //userMappingMap[name]

		t := template.New(field)
		t, err := t.Parse(value)
		if err != nil {
			logger(4, "Unable to parse TEmplate: "+fmt.Sprintf("%v", err), false)
			continue
		}
		buf := bytes.NewBufferString("")
		t.Execute(buf, p)
		value = buf.String()
		if value == "%!s(<nil>)" {
			value = ""
		}

		//-- Process Site
		if field == "Site" {
			//-- Only use Site lookup if enabled and not set to Update only
			if SQLImportConf.SiteLookup.Enabled && SQLImportConf.OrgLookup.Action != updateString {
				value = getSiteFromLookup(value, buf2, espXmlmcLookup)
			}
		}

		//-- Skip UserType Field
		if field == "UserType" && !SQLImportConf.UpdateUserType {
			value = ""
		}

		//-- Skip Password Field
		if field == "Password" {
			value = ""
		}
		//-- if we have Value then set it
		if value != "" {
			err := espXmlmc.SetParam(field, value)
			if err != nil {
				logger(4, "Cant set Paramter: "+fmt.Sprintf("%v", err), false)
			}
		}
	}

	//-- Check for Dry Run
	if configDryRun != true {
		XMLUpdate, xmlmcErr := espXmlmc.Invoke("admin", "userUpdate")
		var xmlRespon xmlmcResponse
		if xmlmcErr != nil {
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLUpdate), &xmlRespon)
		if err != nil {
			return false, err
		}

		if xmlRespon.MethodResult != constOK && xmlRespon.State.ErrorRet != noValuesToUpdate {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCountInc()
			return false, err

		}
		//-- Only use Org lookup if enabled and not set to create only
		if SQLImportConf.OrgLookup.Enabled && SQLImportConf.OrgLookup.Action != createString && len(SQLImportConf.OrgLookup.OrgUnits) > 0 {
			userAddGroups(p, buf2, espXmlmc)
		}
		//-- Process User Status
		if SQLImportConf.UserAccountStatus.Enabled && SQLImportConf.UserAccountStatus.Action != createString {
			userSetStatus(userID, SQLImportConf.UserAccountStatus.Status, buf2, espXmlmc)
		}

		//-- Add Roles
		if SQLImportConf.UserRoleAction != createString && len(SQLImportConf.Roles) > 0 {
			userAddRoles(userID, buf2, espXmlmc)
		}

		//-- Add Image
		if SQLImportConf.ImageLink.Enabled && SQLImportConf.ImageLink.Action != createString && SQLImportConf.ImageLink.URI != "" {
			userAddImage(p, buf2, espXmlmc)
		}

		//-- Process Profile Details
		boolUpdateProfile := userUpdateProfile(p, buf2, espXmlmc, espXmlmcLookup)
		if boolUpdateProfile != true {
			err = errors.New("User Profile Issue (u): " + buf2.String())
			errorCountInc()
			return false, err
		}
		if xmlRespon.State.ErrorRet != noValuesToUpdate {
			buf2.WriteString(loggerGen(1, "User Update Success"))
			updateCountInc()
		} else {
			updateSkippedCountInc()
		}
		logger(1, buf2.String(), false)
		return true, nil
	}
	//-- Inc Counter
	updateSkippedCountInc()
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	logger(1, "User Update XML "+XMLSTRING, false)
	espXmlmc.ClearParam()

	return true, nil
}

func userAddGroups(p map[string]string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) bool {
	for _, orgUnit := range SQLImportConf.OrgLookup.OrgUnits {
		userAddGroup(p, buffer, orgUnit, espXmlmc)
	}
	return true
}
func userAddImage(p map[string]string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) {
	UserID := p[SQLImportConf.SQLConf.UserID]

	t := template.New("i" + UserID)
	t, _ = t.Parse(SQLImportConf.ImageLink.URI)
	buf := bytes.NewBufferString("")
	t.Execute(buf, p)
	value := buf.String()
	if value == "%!s(<nil>)" {
		value = ""
	}
	buffer.WriteString(loggerGen(2, "Image for user: "+value))
	if value == "" {
		return
	}

	if strings.ToUpper(SQLImportConf.ImageLink.UploadType) != "URI" {
		// get binary to upload via WEBDAV and then set value to relative "session" URI

		rel_link := "session/" + UserID
		url := SQLImportConf.DAVURL + rel_link

		var imageB []byte
		var Berr error
		switch strings.ToUpper(SQLImportConf.ImageLink.UploadType) {
		case "URL":
			resp, err := http.Get(value)
			if err != nil {
				buffer.WriteString(loggerGen(4, "Unable to find "+value+" ["+fmt.Sprintf("%v", http.StatusInternalServerError)+"]"))
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == 201 || resp.StatusCode == 200 {
				imageB, _ = ioutil.ReadAll(resp.Body)

			} else {
				buffer.WriteString(loggerGen(4, "Unsuccesful download: "+fmt.Sprintf("%v", resp.StatusCode)))
				_, _ = io.Copy(ioutil.Discard, resp.Body)
				return
			}
		default:
			imageB, Berr = hex.DecodeString(value[2:]) //stripping leading 0x
			if Berr != nil {
				buffer.WriteString(loggerGen(4, "Unsuccesful Decoding "+fmt.Sprintf("%v", Berr)))
				return
			}

		}
		//WebDAV upload
		if len(imageB) > 0 {
			putbody := bytes.NewReader(imageB)
			req, Perr := http.NewRequest("PUT", url, putbody)
			if Perr != nil {
				buffer.WriteString(loggerGen(4, "PUT Request issue: "+fmt.Sprintf("%v", http.StatusInternalServerError)))
				return
			}
			req.Header.Set("Content-Type", "image/jpeg")
			req.Header.Add("Authorization", "ESP-APIKEY "+SQLImportConf.APIKey)
			req.Header.Set("User-Agent", "Go-http-client/1.1")
			response, Perr := client.Do(req)
			if Perr != nil {
				buffer.WriteString(loggerGen(4, "PUT connection issue: "+fmt.Sprintf("%v", http.StatusInternalServerError)))
				return
			}
			defer response.Body.Close()
			_, _ = io.Copy(ioutil.Discard, response.Body)
			if response.StatusCode == 201 || response.StatusCode == 200 {
				buffer.WriteString(loggerGen(1, "Uploaded"))
				value = "/" + rel_link
			} else {
				buffer.WriteString(loggerGen(4, "Unsuccesful Upload: "+fmt.Sprintf("%v", response.StatusCode)))
				return
			}
		} else {
			buffer.WriteString(loggerGen(4, "No Image to upload"))
			return
		}
	}

	espXmlmc.SetParam("objectRef", "urn:sys:user:"+UserID)
	espXmlmc.SetParam("sourceImage", value)

	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("activity", "profileImageSet")
	var xmlRespon xmlmcprofileSetImageResponse
	if xmlmcErr != nil {
		log.Fatal(xmlmcErr)
		buffer.WriteString(loggerGen(4, "Unable to associate Image to User Profile: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Associate Image to User Profile: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Associate Image to User Profile: "+xmlRespon.State.ErrorRet))
		} else {
			buffer.WriteString(loggerGen(1, "Image added to User: "+UserID))
		}
	}
}
func userAddGroup(p map[string]string, buffer *bytes.Buffer, orgUnit OrgUnitStruct, espXmlmc *apiLib.XmlmcInstStruct) bool {

	//-- Check if Site Attribute is set
	if orgUnit.Attribute == "" {
		buffer.WriteString(loggerGen(2, "Org Lookup is Enabled but Attribute is not Defined"))
		return false
	}
	//-- Get Value of Attribute
	t := template.New("orgunit" + strconv.Itoa(orgUnit.Type))
	t, _ = t.Parse(orgUnit.Attribute)
	buf := bytes.NewBufferString("")
	t.Execute(buf, p)
	value := buf.String()
	if value == "%!s(<nil>)" {
		value = ""
	}
	buffer.WriteString(loggerGen(2, "SQL Attribute for Org Lookup: "+value))
	if value == "" {
		return true
	}

	orgAttributeName := processComplexField(value)
	orgIsInCache, orgID := groupInCache(strconv.Itoa(orgUnit.Type) + orgAttributeName)
	//-- Check if we have Chached the site already
	if orgIsInCache {
		buffer.WriteString(loggerGen(1, "Found Org in Cache "+orgID))
		userAddGroupAsoc(p, orgUnit, orgID, buffer, espXmlmc)
		return true
	}

	//-- We Get here if not in cache
	orgIsOnInstance, orgID := searchGroup(orgAttributeName, orgUnit, buffer, espXmlmc)
	if orgIsOnInstance {
		buffer.WriteString(loggerGen(1, "Org Lookup found Id "+orgID))
		userAddGroupAsoc(p, orgUnit, orgID, buffer, espXmlmc)
		return true
	}
	buffer.WriteString(loggerGen(1, "Unable to Find Organisation "+orgAttributeName))
	return false

}

func userAddGroupAsoc(p map[string]string, orgUnit OrgUnitStruct, orgID string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) {
	UserID := p[SQLImportConf.SQLConf.UserID]
	espXmlmc.SetParam("userId", UserID)
	espXmlmc.SetParam("groupId", orgID)
	espXmlmc.SetParam("memberRole", orgUnit.Membership)
	espXmlmc.OpenElement("options")
	espXmlmc.SetParam("tasksView", strconv.FormatBool(orgUnit.TasksView))
	espXmlmc.SetParam("tasksAction", strconv.FormatBool(orgUnit.TasksAction))
	espXmlmc.CloseElement("options")

	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("admin", "userAddGroup")
	var xmlRespon xmlmcuserSetGroupOptionsResponse
	if xmlmcErr != nil {
		log.Fatal(xmlmcErr)
		buffer.WriteString(loggerGen(4, "Unable to Associate User To Group: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Associate User To Group: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			if xmlRespon.State.ErrorRet != "The specified user ["+UserID+"] already belongs to ["+orgID+"] group" {
				buffer.WriteString(loggerGen(4, "Unable to Associate User To Organisation: "+xmlRespon.State.ErrorRet))
			} else {
				buffer.WriteString(loggerGen(1, "User: "+UserID+" Already Added to Organisation: "+orgID))
			}

		} else {
			buffer.WriteString(loggerGen(1, "User: "+UserID+" Added to Organisation: "+orgID))
		}
	}

}

//-- Function to Check if in Cache
func groupInCache(groupName string) (bool, string) {
	boolReturn := false
	stringReturn := ""
	//-- Check if in Cache
	mutexGroups.Lock()
	for _, group := range groups {
		if group.GroupName == groupName {
			boolReturn = true
			stringReturn = group.GroupID
			break
		}
	}
	mutexGroups.Unlock()
	return boolReturn, stringReturn
}

//-- Function to Check if site is on the instance
func searchGroup(orgName string, orgUnit OrgUnitStruct, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) (bool, string) {
	boolReturn := false
	strReturn := ""
	if orgName == "" {
		return boolReturn, strReturn
	}
	espXmlmc.SetParam("application", "com.hornbill.core")
	espXmlmc.SetParam("queryName", "GetGroupByName")
	espXmlmc.OpenElement("queryParams")
	espXmlmc.SetParam("h_name", orgName)
	espXmlmc.SetParam("h_type", strconv.Itoa(orgUnit.Type))
	espXmlmc.CloseElement("queryParams")

	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("data", "queryExec")
	var xmlRespon xmlmcGroupListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Group: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Group: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Group: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.GroupID != "" {
				strReturn = xmlRespon.Params.RowData.Row.GroupID
				boolReturn = true
				//-- Add Group to Cache
				mutexGroups.Lock()
				var newgroupForCache groupListStruct
				newgroupForCache.GroupID = strReturn
				newgroupForCache.GroupName = strconv.Itoa(orgUnit.Type) + orgName
				name := []groupListStruct{newgroupForCache}
				groups = append(groups, name...)
				mutexGroups.Unlock()
			}
		}
	}

	return boolReturn, strReturn
}

func createUser(u map[string]interface{}, espXmlmc *apiLib.XmlmcInstStruct, espXmlmcLookup *apiLib.XmlmcInstStruct) (bool, error) {
	buf2 := bytes.NewBufferString("")
	//-- Do we Lookup Site
	p := make(map[string]string)

	for key, value := range u {
		p[key] = fmt.Sprintf("%s", value)
	}

	userID := p[SQLImportConf.SQLConf.UserID]

	//-- Loop Through UserProfileMapping
	for key := range userCreateArray {
		field := userCreateArray[key]
		value := SQLImportConf.UserMapping[field] //userMappingMap[name]
		t := template.New(field)
		t, _ = t.Parse(value)
		buf := bytes.NewBufferString("")
		t.Execute(buf, p)
		value = buf.String()
		if value == "%!s(<nil>)" {
			value = ""
		}

		//-- Process Site
		if field == "Site" {
			//-- Only use Site lookup if enabled and not set to Update only
			if SQLImportConf.SiteLookup.Enabled && SQLImportConf.OrgLookup.Action != updateString {
				value = getSiteFromLookup(value, buf2, espXmlmcLookup)
			}
		}
		//-- Process Password Field
		if field == "Password" {
			if value == "" {
				value = generatePasswordString(10)
				logger(1, "Auto Generated Password for: "+userID+" - "+value, false)
			}
			value = base64.StdEncoding.EncodeToString([]byte(value))
		}

		//-- if we have Value then set it
		if value != "" {
			espXmlmc.SetParam(field, value)

		}
	}

	//-- Check for Dry Run
	if configDryRun != true {
		XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userCreate")
		var xmlRespon xmlmcResponse
		if xmlmcErr != nil {
			errorCountInc()
			return false, xmlmcErr
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			errorCountInc()
			return false, err
		}
		if xmlRespon.MethodResult != constOK {
			err = errors.New(xmlRespon.State.ErrorRet)
			errorCountInc()
			return false, err

		}
		logger(1, "User Create Success", false)

		//-- Only use Org lookup if enabled and not set to Update only
		if SQLImportConf.OrgLookup.Enabled && SQLImportConf.OrgLookup.Action != updateString && len(SQLImportConf.OrgLookup.OrgUnits) > 0 {
			userAddGroups(p, buf2, espXmlmc)
		}
		//-- Process Account Status
		if SQLImportConf.UserAccountStatus.Enabled && SQLImportConf.UserAccountStatus.Action != updateString {
			userSetStatus(userID, SQLImportConf.UserAccountStatus.Status, buf2, espXmlmc)
		}

		if SQLImportConf.UserRoleAction != updateString && len(SQLImportConf.Roles) > 0 {
			userAddRoles(userID, buf2, espXmlmc)
		}

		//-- Add Image
		if SQLImportConf.ImageLink.Enabled && SQLImportConf.ImageLink.Action != updateString && SQLImportConf.ImageLink.URI != "" {
			userAddImage(p, buf2, espXmlmc)
		}

		//-- Process Profile Details
		boolUpdateProfile := userUpdateProfile(p, buf2, espXmlmc, espXmlmcLookup)
		if boolUpdateProfile != true {
			err = errors.New("User Profile issue (c): " + buf2.String())
			errorCountInc()
			return false, err
		}

		logger(1, buf2.String(), false)
		createCountInc()
		return true, nil
	}
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	logger(1, "User Create XML "+XMLSTRING, false)
	createSkippedCountInc()
	espXmlmc.ClearParam()

	return true, nil
}

func userUpdateProfile(p map[string]string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct, espXmlmcLookup *apiLib.XmlmcInstStruct) bool {
	UserID := p[SQLImportConf.SQLConf.UserID]
	buffer.WriteString(loggerGen(1, "Processing User Profile Data "+UserID))
	espXmlmc.OpenElement("profileData")
	espXmlmc.SetParam("userID", UserID)
	//-- Loop Through UserProfileMapping
	for key := range userProfileArray {
		field := userProfileArray[key]
		value := SQLImportConf.UserProfileMapping[field]

		t := template.New(field)
		t, _ = t.Parse(value)
		buf := bytes.NewBufferString("")
		t.Execute(buf, p)
		value = buf.String()
		if value == "%!s(<nil>)" {
			value = ""
		}

		if field == "Manager" {
			//-- Process User manager
			if SQLImportConf.UserManagerMapping.Enabled && SQLImportConf.UserManagerMapping.Action != updateString {
				value = getManagerFromLookup(value, buffer, espXmlmcLookup)
			}
		}

		//-- if we have Value then set it
		if value != "" {
			espXmlmc.SetParam(field, value)
		}
	}

	espXmlmc.CloseElement("profileData")
	//-- Check for Dry Run
	if configDryRun != true {
		XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userProfileSet")
		var xmlRespon xmlmcResponse
		if xmlmcErr != nil {
			buffer.WriteString(loggerGen(4, "Unable to Update User Profile: "+fmt.Sprintf("%v", xmlmcErr)))
			return false
		}
		err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
		if err != nil {
			buffer.WriteString(loggerGen(4, "Unable to Update User Profile: "+fmt.Sprintf("%v", err)))

			return false
		}
		if xmlRespon.MethodResult != constOK {
			profileSkippedCountInc()
			if xmlRespon.State.ErrorRet == noValuesToUpdate {
				return true
			}
			err := errors.New(xmlRespon.State.ErrorRet)
			buffer.WriteString(loggerGen(4, "Unable to Update User Profile: "+fmt.Sprintf("%v", err)))
			return false
		}
		profileCountInc()
		buffer.WriteString(loggerGen(1, "User Profile Update Success"))
		return true

	}
	//-- DEBUG XML TO LOG FILE
	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Profile Update XML "+XMLSTRING))
	profileSkippedCountInc()
	espXmlmc.ClearParam()
	return true

}

func userSetStatus(userID string, status string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) bool {
	buffer.WriteString(loggerGen(1, "Set Status for User: "+userID+" Status:"+status))

	espXmlmc.SetParam("userId", userID)
	espXmlmc.SetParam("accountStatus", status)

	XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userSetAccountStatus")

	var XMLSTRING = espXmlmc.GetParam()
	buffer.WriteString(loggerGen(1, "User Create XML "+XMLSTRING))

	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to Set User Status: "+fmt.Sprintf("%s", xmlmcErr), true)

	}
	err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Set User Status "+fmt.Sprintf("%s", err)))
		return false
	}
	if xmlRespon.MethodResult != constOK {
		if xmlRespon.State.ErrorRet != "Failed to update account status (target and the current status is the same)." {
			buffer.WriteString(loggerGen(4, "Unable to Set User Status 111: "+xmlRespon.State.ErrorRet))
			return false
		}
		buffer.WriteString(loggerGen(1, "User Status Already Set to: "+status))
		return true
	}
	buffer.WriteString(loggerGen(1, "User Status Set Successfully"))
	return true
}

func userAddRoles(userID string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) bool {

	espXmlmc.SetParam("userId", userID)
	for _, role := range SQLImportConf.Roles {
		espXmlmc.SetParam("role", role)
		buffer.WriteString(loggerGen(1, "Add Role to User: "+role))
	}
	XMLCreate, xmlmcErr := espXmlmc.Invoke("admin", "userAddRole")
	var xmlRespon xmlmcResponse
	if xmlmcErr != nil {
		logger(4, "Unable to Assign Role to User: "+fmt.Sprintf("%s", xmlmcErr), true)

	}
	err := xml.Unmarshal([]byte(XMLCreate), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Assign Role to User: "+fmt.Sprintf("%s", err)))
		return false
	}
	if xmlRespon.MethodResult != constOK {
		buffer.WriteString(loggerGen(4, "Unable to Assign Role to User: "+xmlRespon.State.ErrorRet))
		return false
	}
	buffer.WriteString(loggerGen(1, "Roles Added Successfully"))
	return true
}

func checkUserOnInstance(userID string, espXmlmc *apiLib.XmlmcInstStruct) (bool, error) {

	espXmlmc.SetParam("entity", "UserAccount")
	espXmlmc.SetParam("keyValue", userID)
	XMLCheckUser, xmlmcErr := espXmlmc.Invoke("data", "entityDoesRecordExist")
	var xmlRespon xmlmcCheckUserResponse
	if xmlmcErr != nil {
		return false, xmlmcErr
	}
	err := xml.Unmarshal([]byte(XMLCheckUser), &xmlRespon)
	if err != nil {
		stringError := err.Error()
		stringBody := string(XMLCheckUser)
		errWithBody := errors.New(stringError + " RESPONSE BODY: " + stringBody)
		return false, errWithBody
	}
	if xmlRespon.MethodResult != constOK {
		err := errors.New(xmlRespon.State.ErrorRet)
		return false, err
	}
	return xmlRespon.Params.RecordExist, nil
}

//-- Function to search for site
func getSiteFromLookup(site string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) string {
	siteReturn := ""

	//-- Get Value of Attribute
	siteAttributeName := processComplexField(site)
	buffer.WriteString(loggerGen(1, "Looking Up Site: "+siteAttributeName))
	if siteAttributeName == "" {
		return ""
	}
	siteIsInCache, SiteIDCache := siteInCache(siteAttributeName)
	//-- Check if we have Cached the site already
	if siteIsInCache {
		siteReturn = strconv.Itoa(SiteIDCache)
		buffer.WriteString(loggerGen(1, "Found Site in Cache: "+siteReturn))
	} else {
		siteIsOnInstance, SiteIDInstance := searchSite(siteAttributeName, buffer, espXmlmc)
		//-- If Returned set output
		if siteIsOnInstance {
			siteReturn = strconv.Itoa(SiteIDInstance)
		}
	}
	buffer.WriteString(loggerGen(1, "Site Lookup found ID: "+siteReturn))
	return siteReturn
}

func processComplexField(s string) string {
	return html.UnescapeString(s)
}

//-- Function to Check if in Cache
func siteInCache(siteName string) (bool, int) {
	boolReturn := false
	intReturn := 0
	mutexSites.Lock()
	//-- Check if in Cache
	for _, site := range sites {
		if site.SiteName == siteName {
			boolReturn = true
			intReturn = site.SiteID
			break
		}
	}
	mutexSites.Unlock()
	return boolReturn, intReturn
}

//-- Function to Check if site is on the instance
func searchSite(siteName string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) (bool, int) {
	boolReturn := false
	intReturn := 0
	if siteName == "" {
		return boolReturn, intReturn
	}
	espXmlmc.SetParam("entity", "Site")
	espXmlmc.SetParam("matchScope", "all")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("h_site_name", siteName)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")
	XMLSiteSearch, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords")

	var xmlRespon xmlmcSiteListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Site: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLSiteSearch), &xmlRespon)
	if err != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Site: "+fmt.Sprintf("%v", err)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Site: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.SiteName != "" {
				if strings.ToLower(xmlRespon.Params.RowData.Row.SiteName) == strings.ToLower(siteName) {
					intReturn = xmlRespon.Params.RowData.Row.SiteID
					boolReturn = true
					//-- Add Site to Cache
					mutexSites.Lock()
					var newSiteForCache siteListStruct
					newSiteForCache.SiteID = intReturn
					newSiteForCache.SiteName = siteName
					name := []siteListStruct{newSiteForCache}
					sites = append(sites, name...)
					mutexSites.Unlock()
				}
			}
		}
	}

	return boolReturn, intReturn
}

func getManagerFromLookup(manager string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) string {

	if manager == "" {
		buffer.WriteString(loggerGen(1, "No Manager to search"))
		return ""
	}
	//-- Get Value of Attribute
	ManagerAttributeName := processComplexField(manager)
	buffer.WriteString(loggerGen(1, "Manager Lookup: "+ManagerAttributeName))

	//-- Dont Continue if we didn't get anything
	if ManagerAttributeName == "" {
		return ""
	}

	buffer.WriteString(loggerGen(1, "Looking Up Manager "+ManagerAttributeName))
	managerIsInCache, ManagerIDCache := managerInCache(ManagerAttributeName)

	//-- Check if we have Chached the site already
	if managerIsInCache {
		buffer.WriteString(loggerGen(1, "Found Manager in Cache "+ManagerIDCache))
		return ManagerIDCache
	}
	buffer.WriteString(loggerGen(1, "Manager Not In Cache Searching"))
	ManagerIsOnInstance, ManagerIDInstance := searchManager(ManagerAttributeName, buffer, espXmlmc)
	//-- If Returned set output
	if ManagerIsOnInstance {
		buffer.WriteString(loggerGen(1, "Manager Lookup found Id "+ManagerIDInstance))

		return ManagerIDInstance
	}

	return ""
}

//-- Search Manager on Instance
func searchManager(managerName string, buffer *bytes.Buffer, espXmlmc *apiLib.XmlmcInstStruct) (bool, string) {
	boolReturn := false
	strReturn := ""
	espXmlmc.SetTrace("SQLUserImport")
	if managerName == "" {
		return boolReturn, strReturn
	}

	espXmlmc.SetParam("entity", "UserAccount")
	espXmlmc.SetParam("matchScope", "all")
	espXmlmc.OpenElement("searchFilter")
	espXmlmc.SetParam("h_name", managerName)
	espXmlmc.CloseElement("searchFilter")
	espXmlmc.SetParam("maxResults", "1")
	XMLUserSearch, xmlmcErr := espXmlmc.Invoke("data", "entityBrowseRecords")
	var xmlRespon xmlmcUserListResponse
	if xmlmcErr != nil {
		buffer.WriteString(loggerGen(4, "Unable to Search for Manager: "+fmt.Sprintf("%v", xmlmcErr)))
	}
	err := xml.Unmarshal([]byte(XMLUserSearch), &xmlRespon)
	if err != nil {
		stringError := err.Error()
		stringBody := string(XMLUserSearch)
		buffer.WriteString(loggerGen(4, "Unable to Search for Manager: "+fmt.Sprintf("%v", stringError+" RESPONSE BODY: "+stringBody)))
	} else {
		if xmlRespon.MethodResult != constOK {
			buffer.WriteString(loggerGen(4, "Unable to Search for Manager: "+xmlRespon.State.ErrorRet))
		} else {
			//-- Check Response
			if xmlRespon.Params.RowData.Row.UserName != "" {
				if strings.ToLower(xmlRespon.Params.RowData.Row.UserName) == strings.ToLower(managerName) {

					strReturn = xmlRespon.Params.RowData.Row.UserID
					boolReturn = true
					//-- Add Site to Cache
					mutexManagers.Lock()
					var newManagerForCache managerListStruct
					newManagerForCache.UserID = strReturn
					newManagerForCache.UserName = managerName
					name := []managerListStruct{newManagerForCache}
					managers = append(managers, name...)
					mutexManagers.Unlock()
				}
			}
		}
	}
	return boolReturn, strReturn
}

//-- Check if Manager in Cache
func managerInCache(managerName string) (bool, string) {
	boolReturn := false
	stringReturn := ""
	//-- Check if in Cache
	mutexManagers.Lock()
	for _, manager := range managers {
		if strings.ToLower(manager.UserName) == strings.ToLower(managerName) {
			boolReturn = true
			stringReturn = manager.UserID
		}
	}
	mutexManagers.Unlock()
	return boolReturn, stringReturn
}

//-- Generate Password String
func generatePasswordString(n int) string {
	var arbytes = make([]byte, n)
	rand.Read(arbytes)
	for i, b := range arbytes {
		arbytes[i] = letterBytes[b%byte(len(letterBytes))]
	}
	return string(arbytes)
}

// =================== COUNTERS =================== //
func errorCountInc() {
	mutexCounters.Lock()
	errorCount++
	mutexCounters.Unlock()
}
func updateCountInc() {
	mutexCounters.Lock()
	counters.updated++
	mutexCounters.Unlock()
}
func updateSkippedCountInc() {
	mutexCounters.Lock()
	counters.updatedSkipped++
	mutexCounters.Unlock()
}
func createSkippedCountInc() {
	mutexCounters.Lock()
	counters.createskipped++
	mutexCounters.Unlock()
}
func createCountInc() {
	mutexCounters.Lock()
	counters.created++
	mutexCounters.Unlock()
}
func profileCountInc() {
	mutexCounters.Lock()
	counters.profileUpdated++
	mutexCounters.Unlock()
}
func profileSkippedCountInc() {
	mutexCounters.Lock()
	counters.profileSkipped++
	mutexCounters.Unlock()
}
