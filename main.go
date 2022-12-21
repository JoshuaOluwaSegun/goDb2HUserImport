package main

//----- Packages -----
import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	//-- CLI Colour
	"github.com/blang/semver"
	"github.com/hornbill/color"
	"github.com/rhysd/go-github-selfupdate/selfupdate"

	//-- Hornbill Clone of "github.com/mavricknz/ldap"
	//--Hornbil Clone of "github.com/cheggaaa/pb"

	apiLib "github.com/hornbill/goApiLib"
	//-- For Version checking
)

var (
	onceLog   sync.Once
	loggerAPI *apiLib.XmlmcInstStruct
	mutexLog  = &sync.Mutex{}
	f         *os.File
)

// Main
func main() {
	//-- Start Time for Durration
	Time.startTime = time.Now()
	//-- Start Time for Log File
	Time.timeNow = time.Now().Format(time.RFC3339)
	//-- Remove :
	Time.timeNow = strings.Replace(Time.timeNow, ":", "-", -1)

	//-- Process Flags
	procFlags()

	//-- Used for Building
	if Flags.configVersion {
		fmt.Printf("%v \n", version)
		return
	}

	//-- Check for latest
	doSelfUpdate()

	//-- Load Configuration File Into Struct
	SQLImportConf = loadConfig()

	//-- Validation on Configuration File
	configError := validateConf()

	//-- Check for Error
	if configError != nil {
		logger(4, fmt.Sprintf("%v", configError), true)
		logger(4, "Please Check your Configuration: "+Flags.configID, true)
		return
	}

	if Flags.configInstanceID == "" {
		Flags.configInstanceID = SQLImportConf.InstanceID
	}

	if Flags.configAPIKey == "" {
		Flags.configAPIKey = SQLImportConf.APIKey
	}

	getServerBuild()

	sysSettingPageSize, _ := strconv.Atoi(sysOptionGet("api.xmlmc.queryExec.maxResultsAllowed"))
	if SQLImportConf.Advanced.PageSize < sysSettingPageSize {
		SQLImportConf.Advanced.PageSize = sysSettingPageSize
		logger(0, "[MESSAGE] Overridden Page Size "+fmt.Sprintf("%d", SQLImportConf.Advanced.PageSize), true)
	}
	//-- Check import not already running
	getLastHistory()

	//-- Start Import
	logged := startImportHistory()
	//-- Check for Connections
	if !logged {
		logger(4, "Unable to Connect to Instance", true)
		return
	}

	//-- Clear Old Log Files
	runLogRetentionCheck()

	//-- Get Password Profile
	getPasswordProfile()

	SQLImportConf.User.HornbillUserIDColumn = strings.ToLower(SQLImportConf.User.HornbillUserIDColumn)

	//-- Query DB
	queryDB()

	//-- Process DB User Data First
	//-- So we only store data about users we have
	processDBUsers()

	//-- Fetch Users from Hornbill
	loadUsers()

	//-- Load User Roles
	loadUsersRoles()

	//-- Fetch Sites
	loadSites()

	//-- Fetch Groups
	loadGroups()

	//-- Fetch User Groups
	loadUserGroups()

	//-- Create List of Actions that need to happen
	//-- (Create,Update,profileUpdate,Assign Role, Assign Group, Assign Site)
	processData()

	//-- Run Actions
	finaliseData()

	//-- End Ouput
	outputEnd()
}

func outputFlags() {
	//-- Output
	logger(1, "---- XMLMC SQL Import Utility V"+fmt.Sprintf("%v", version)+" ----", true)

	logger(1, "Flag - Config File "+Flags.configFileName, true)
	logger(1, "Flag - Zone "+Flags.configZone, true)
	logger(1, "Flag - Log Prefix "+Flags.configLogPrefix, true)
	logger(1, "Flag - Dry Run "+fmt.Sprintf("%v", Flags.configDryRun), true)
	logger(1, "Flag - Workers "+fmt.Sprintf("%v", Flags.configWorkers), false)
}

// -- Process Input Flags
func procFlags() {
	//-- Grab Flags
	flag.StringVar(&Flags.configFileName, "file", "conf.json", "Name of Configuration File To Load")
	flag.StringVar(&Flags.configZone, "zone", "eur", "Override the default Zone the instance sits in")
	flag.StringVar(&Flags.configLogPrefix, "logprefix", "", "Add prefix to the logfile")
	flag.BoolVar(&Flags.configDryRun, "dryrun", false, "Allow the Import to run without Creating or Updating users")
	flag.BoolVar(&Flags.configVersion, "version", false, "Output Version")
	flag.IntVar(&Flags.configWorkers, "workers", 1, "Number of Worker threads to use")
	flag.StringVar(&Flags.configMaxRoutines, "concurrent", "1", "Maximum number of requests to import concurrently.")
	flag.IntVar(&Flags.configAPITimeout, "apitimeout", 60, "Number of Seconds to Timeout an API Connection")
	flag.BoolVar(&Flags.configForceRun, "forcerun", false, "Bypass check on existing running import")

	//-- Parse Flags
	flag.Parse()
	Flags.configID = appName
	//-- Output config
	if !Flags.configVersion {
		outputFlags()
	}

	//Check maxGoroutines for valid value
	maxRoutines, err := strconv.Atoi(Flags.configMaxRoutines)
	if err != nil {
		color.Red("Unable to convert maximum concurrency of [" + Flags.configMaxRoutines + "] to type INT for processing")
		return
	}
	maxGoroutines = maxRoutines

	if maxGoroutines < 1 || maxGoroutines > 10 {
		color.Red("The maximum concurrent requests allowed is between 1 and 10 (inclusive).\n\n")
		color.Red("You have selected " + Flags.configMaxRoutines + ". Please try again, with a valid value against ")
		color.Red("the -concurrent switch.")
		return
	}

	//-- Output config
	if !Flags.configVersion {
		logger(2, "---- "+applicationName+" v"+fmt.Sprintf("%v", version)+" ----", true)
		logger(2, "Flag - config "+Flags.configID, true)
		logger(2, "Flag - logprefix "+Flags.configLogPrefix, true)
		logger(2, "Flag - dryrun "+fmt.Sprintf("%v", Flags.configDryRun), true)
		logger(2, "Flag - instanceid "+Flags.configInstanceID, true)
		logger(2, "Flag - apitimeout "+fmt.Sprintf("%v", Flags.configAPITimeout), true)
		logger(2, "Flag - workers "+fmt.Sprintf("%v", Flags.configWorkers)+"\n", true)
		logger(2, "Flag - forcerun "+fmt.Sprintf("%v", Flags.configForceRun), true)
	}
}

// -- Generate Output
func outputEnd() {
	logger(2, "Import Complete", true)
	//-- End output
	if counters.errors > 0 {
		logger(4, "One or more errors encountered, please check the log file", true)
		logger(4, "Error Count: "+fmt.Sprintf("%d", counters.errors), true)
		//logger(4, "Check Log File for Details", true)
	}
	logger(2, "Accounts Processed: "+fmt.Sprintf("%d", len(HornbillCache.UsersWorking)), true)
	logger(2, "Created: "+fmt.Sprintf("%d", counters.created), true)
	logger(2, "Updated: "+fmt.Sprintf("%d", counters.updated), true)

	logger(2, "Status Updates: "+fmt.Sprintf("%d", counters.statusUpdated), true)

	logger(2, "Profiles Updated: "+fmt.Sprintf("%d", counters.profileUpdated), true)

	logger(2, "Images Updated: "+fmt.Sprintf("%d", counters.imageUpdated), true)
	logger(2, "Groups Added: "+fmt.Sprintf("%d", counters.groupUpdated), true)
	logger(2, "Groups Removed: "+fmt.Sprintf("%d", counters.groupsRemoved), true)
	logger(2, "Roles Added: "+fmt.Sprintf("%d", counters.rolesUpdated), true)

	//-- Show Time Takens
	Time.endTime = time.Since(Time.startTime).Round(time.Second)
	logger(2, "Time Taken: "+Time.endTime.String(), true)
	//-- complete
	mutexCounters.Lock()
	counters.traffic += loggerAPI.GetCount()
	counters.traffic += hornbillImport.GetCount()
	mutexCounters.Unlock()

	logger(2, "Total Traffic: "+fmt.Sprintf("%d", counters.traffic), true)

	completeImportHistory()
	logger(2, "---- XMLMC DB Import Complete ---- ", true)
}

func loadConfig() sqlImportConfStruct {
	//-- Check Config File File Exists
	cwd, _ := os.Getwd()
	configurationFilePath := cwd + "/" + Flags.configFileName
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

	eldapConf := sqlImportConfStruct{}

	//-- Decode JSON
	err := decoder.Decode(&eldapConf)
	//-- Error Checking
	if err != nil {
		logger(4, "Error Decoding Configuration File: "+fmt.Sprintf("%v", err), true)
	}
	logger(0, "[MESSAGE] Import Defined Page Size "+fmt.Sprintf("%d", eldapConf.Advanced.PageSize), true)
	//-- Return New Congfig
	return eldapConf
}

// -- Function to Load Configuration File
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
	if SQLImportConf.Action == "" {
		SQLImportConf.Action = "Both"
	}

	return nil
}

// CounterInc Generic Counter Increment
func CounterInc(counter int) {
	mutexCounters.Lock()
	switch counter {
	case 1:
		counters.created++
	case 2:
		counters.updated++
	case 3:
		counters.profileUpdated++
	case 4:
		counters.imageUpdated++
	case 5:
		counters.groupUpdated++
	case 6:
		counters.rolesUpdated++
	case 7:
		counters.errors++
	case 8:
		counters.groupsRemoved++
	case 9:
		counters.statusUpdated++
	}
	mutexCounters.Unlock()
}

func doSelfUpdate() {
	logger(1, "Checking "+repo+" for updates...", true)
	v := semver.MustParse(version)
	latest, found, err := selfupdate.DetectLatest(repo)

	if err != nil {
		logger(5, "Error occurred while detecting version: "+err.Error(), true)
		return
	}
	if !found {
		logger(5, "Could not find Github repo, or existing release in required format.", true)
		return
	}

	latestMajorVersion := strings.Split(fmt.Sprintf("%v", latest.Version), ".")[0]
	latestMinorVersion := strings.Split(fmt.Sprintf("%v", latest.Version), ".")[1]
	latestPatchVersion := strings.Split(fmt.Sprintf("%v", latest.Version), ".")[2]

	currentMajorVersion := strings.Split(version, ".")[0]
	currentMinorVersion := strings.Split(version, ".")[1]
	currentPatchVersion := strings.Split(version, ".")[2]

	//Useful in dev, customers should never see current version > latest release version
	if currentMajorVersion > latestMajorVersion {
		logger(3, "Current version "+version+" (major) is greater than the latest release version on Github "+fmt.Sprintf("%v", latest.Version), true)
		return
	} else {
		if currentMinorVersion > latestMinorVersion {
			logger(3, "Current version "+version+" (minor) is greater than the latest release version on Github "+fmt.Sprintf("%v", latest.Version), true)
			return
		} else if currentPatchVersion > latestPatchVersion {
			logger(3, "Current version "+version+" (patch) is greater than the latest release version on Github "+fmt.Sprintf("%v", latest.Version), true)
			return
		}
	}
	if latestMajorVersion > currentMajorVersion {
		msg := "v" + version + " is not latest, you should upgrade to " + fmt.Sprintf("%v", latest.Version) + " by downloading the latest package from: https://github.com/" + repo + "/releases/latest"
		logger(5, msg, true)
		return
	}

	_, err = selfupdate.UpdateSelf(v, repo)
	if err != nil {
		logger(5, "Binary update failed: "+err.Error(), true)
		return
	}
	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		logger(3, "Current binary is the latest version: "+version, true)
	} else {
		logger(3, "Successfully updated to version: "+fmt.Sprintf("%v", latest.Version), true)
		logger(3, "Release notes:\n"+latest.ReleaseNotes, true)
	}
}
