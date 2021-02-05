package main

import (
	"fmt"
	"regexp"
	"strings"
)

func getManager(importData *userWorkingDataStruct, currentData userAccountStruct) string {
	//-- Check if Manager Attribute is set
	if SQLImportConf.User.Manager.Value == "" {
		logger(4, "Manager Lookup is Enabled but Attribute is not Defined", false)
		return ""
	}

	//-- Get Value of Attribute
	logger(1, "DB Attribute for Manager Lookup: "+SQLImportConf.User.Manager.Value, false)

	//-- Get Value of Attribute
	ManagerAttributeName := processComplexField(importData.DB, SQLImportConf.User.Manager.Value)
	ManagerAttributeName = processImportAction(importData.Custom, ManagerAttributeName)

	if SQLImportConf.User.Manager.Options.MatchAgainstDistinguishedName {
		logger(1, "Searching Distinguished Name Cache for: "+ManagerAttributeName, false)
		managerID := getUserFromDNCache(ManagerAttributeName)
		if managerID != "" {
			logger(1, "Found Manager in Distinguished Name  Cache: "+managerID, false)
			return managerID
		}
		logger(1, "Unable to find Manager in Distinguished Name  Cache Coninuing search", false)
	}

	//-- Dont Continue if we didn't get anything
	if ManagerAttributeName == "" {
		return ""
	}

	//-- Pull Data from Attriute using regext
	if SQLImportConf.User.Manager.Options.GetStringFromValue.Regex != "" {
		logger(1, "DB Manager String: "+ManagerAttributeName, false)
		ManagerAttributeName = getNameFromDBString(ManagerAttributeName)
	}
	//-- Is Search Enabled
	if SQLImportConf.User.Manager.Options.Search.Enable {
		logger(1, "Search for Manager is Enabled", false)

		logger(1, "Looking Up Manager from Cache: "+ManagerAttributeName, false)
		managerIsInCache, ManagerIDCache := managerInCache(ManagerAttributeName)

		//-- Check if we have Chached the site already
		if managerIsInCache {
			logger(1, "Found Manager in Cache: "+ManagerIDCache, false)
			return ManagerIDCache
		}
		logger(1, "Manager Not In Cache Searching Hornbill Data", false)
		ManagerIsOnInstance, ManagerIDInstance := searchManager(ManagerAttributeName)
		//-- If Returned set output
		if ManagerIsOnInstance {
			logger(1, "Manager Lookup found Id: "+ManagerIDInstance, false)
			return ManagerIDInstance
		}
	} else {
		logger(1, "Search for Manager is Disabled", false)
		//-- Assume data is manager id
		logger(1, "Manager Id: "+ManagerAttributeName, false)
		return ManagerAttributeName
	}

	//else return empty
	return ""
}

//-- Search Manager on Instance
func searchManager(managerName string) (bool, string) {
	//-- ESP Query for site
	if managerName == "" {
		return false, ""
	}

	//-- Add support for Search Feild configuration
	strSearchField := "h_name"
	if SQLImportConf.User.Manager.Options.Search.SearchField != "" {
		strSearchField = SQLImportConf.User.Manager.Options.Search.SearchField
	}

	logger(1, "Manager Search: "+strSearchField+" - "+managerName, false)

	//-- Check User Cache for Manager
	strFieldToMatch := "";
	for _, v := range HornbillCache.Users {
		switch c := strSearchField; c {
			case "h_user_id": strFieldToMatch = v.HUserID
			case "h_login_id": strFieldToMatch = v.HLoginID
			case "h_employee_id": strFieldToMatch = v.HEmployeeID
			case "h_name": strFieldToMatch = v.HName
			case "h_email": strFieldToMatch = v.HEmail
			case "h_attrib_1": strFieldToMatch = v.HAttrib1
			case "h_attrib_2": strFieldToMatch = v.HAttrib2
			case "h_attrib_3": strFieldToMatch = v.HAttrib3
			case "h_sn_a": strFieldToMatch = v.HSnA
			case "h_sn_b": strFieldToMatch = v.HSnB
			case "h_sn_c": strFieldToMatch = v.HSnC
			case "h_site": strFieldToMatch = v.HSite
			case "h_home_organization": strFieldToMatch = v.HHomeOrg
			case "h_attrib_4": strFieldToMatch = v.HAttrib4
			case "h_attrib_5": strFieldToMatch = v.HAttrib5
			case "h_attrib_6": strFieldToMatch = v.HAttrib6
			case "h_attrib_7": strFieldToMatch = v.HAttrib7
			case "h_attrib_8": strFieldToMatch = v.HAttrib8
			case "h_sn_d": strFieldToMatch = v.HSnD
			case "h_sn_e": strFieldToMatch = v.HSnE
			case "h_sn_f": strFieldToMatch = v.HSnF
			case "h_sn_g": strFieldToMatch = v.HSnG
			case "h_sn_h": strFieldToMatch = v.HSnH

/* there should be no reason to match on any of those below
			case "h_mobile": strFieldToMatch = v.HMobile
			case "h_first_name": strFieldToMatch = v.HFirstName
			case "h_middle_name": strFieldToMatch = v.HMiddleName
			case "h_last_name": strFieldToMatch = v.HLastName
			case "h_phone": strFieldToMatch = v.HPhone
			case "h_job_title": strFieldToMatch = v.HJobTitle
			case "h_login_creds": strFieldToMatch = v.HLoginCreds
			case "h_class": strFieldToMatch = v.HClass
			case "h_avail_status": strFieldToMatch = v.HAvailStatus
			case "h_avail_status_msg": strFieldToMatch = v.HAvailStatusMsg
			case "h_timezone": strFieldToMatch = v.HTimezone
			case "h_country": strFieldToMatch = v.HCountry
			case "h_language": strFieldToMatch = v.HLanguage
			case "h_date_time_format": strFieldToMatch = v.HDateTimeFormat
			case "h_date_format": strFieldToMatch = v.HDateFormat
			case "h_time_format": strFieldToMatch = v.HTimeFormat
			case "h_currency_symbol": strFieldToMatch = v.HCurrencySymbol
			case "h_last_logon": strFieldToMatch = v.HLastLogon
			case "h_icon_ref": strFieldToMatch = v.HIconRef
			case "h_icon_checksum": strFieldToMatch = v.HIconChecksum
			case "h_dob": strFieldToMatch = v.HDob
			case "h_account_status": strFieldToMatch = v.HAccountStatus
			case "h_failed_attempts": strFieldToMatch = v.HFailedAttempts
			case "h_idx_ref": strFieldToMatch = v.HIdxRef
			case "h_manager": strFieldToMatch = v.HManager
			case "h_summary": strFieldToMatch = v.HSummary
			case "h_interests": strFieldToMatch = v.HInterests
			case "h_qualifications": strFieldToMatch = v.HQualifications
			case "h_personal_interests": strFieldToMatch = v.HPersonalInterests
			case "h_skills": strFieldToMatch = v.HSkills
			case "h_gender": strFieldToMatch = v.HGender
			case "h_nationality": strFieldToMatch = v.HNationality
			case "h_religion": strFieldToMatch = v.HReligion
			case "h_home_telephone_number": strFieldToMatch = v.HHomeTelephoneNumber
			case "h_home_address": strFieldToMatch = v.HHomeAddress
			case "h_blog": strFieldToMatch = v.HBlog
*/
		}
		if strings.EqualFold(strFieldToMatch, managerName) {
			//-- If not already in cache push to cache
			_, found := HornbillCache.Managers[strings.ToLower(managerName)]
			if !found {
				HornbillCache.Managers[strings.ToLower(managerName)] = v.HUserID
			}
			return true, v.HUserID
		}
	}

	return false, ""
}

//-- Check if Manager in Cache
func managerInCache(managerName string) (bool, string) {
	//-- Check if in Cache
	_, found := HornbillCache.Managers[strings.ToLower(managerName)]
	if found {
		return true, HornbillCache.Managers[strings.ToLower(managerName)]
	}
	return false, ""
}

//-- Takes a string based on a DB DN and returns to the CN String Name
func getNameFromDBString(feild string) string {

	regex := SQLImportConf.User.Manager.Options.GetStringFromValue.Regex
	reverse := SQLImportConf.User.Manager.Options.GetStringFromValue.Reverse
	stringReturn := ""

	//-- Match $variables from String
	re1, err := regexp.Compile(regex)
	if err != nil {
		logger(4, "Error Compiling Regex: "+regex+" Error: "+fmt.Sprintf("%v", err), false)

	}
	//-- Get Array of all Matched max 100
	result := re1.FindAllString(feild, 100)

	//-- Loop Matches
	for _, v := range result {
		//-- String DB String Chars Out from match
		v = strings.Replace(v, "CN=", "", -1)
		v = strings.Replace(v, "OU=", "", -1)
		v = strings.Replace(v, "DC=", "", -1)
		v = strings.Replace(v, "\\", "", -1)
		nameArray := strings.Split(v, ",")

		for _, n := range nameArray {
			n = strings.Trim(n, " ")
			if n != "" {
				if reverse {
					stringReturn = n + " " + stringReturn
				} else {
					stringReturn = stringReturn + " " + n
				}
			}

		}

	}
	stringReturn = strings.Trim(stringReturn, " ")
	return stringReturn
}
