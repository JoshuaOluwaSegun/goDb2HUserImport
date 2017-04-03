### DB Import Go - [GO](https://golang.org/) Import Script to Hornbill

### Quick links
- [Installation](#installation)
- [Config](#config)
    - [Instance Config](#InstanceConfig)
    - [SQL Config](#SQLConfig)
    - [SQL Mapping](#UserMapping)
- [Execute](#execute)
- [proxy](#proxy)
- [Testing](testing)
- [Scheduling](#scheduling)
- [Logging](#logging)
- [Error Codes](#Error_Codes)

# Installation

#### Windows
* compile into a folder you would like the application to run from e.g. `C:\goDb2HUserImport\`
* Open '''conf.json''' and add in the necessary configration
* Open Command Line Prompt as Administrator
* Change Directory to the folder with goDb2HUserImport.exe `C:\goDb2HUserImport\`
* Run the command goDb2HUserImport.exe -dryrun=true

# config

Example JSON File:

```json
{
    "APIKey": "",
    "InstanceId": "",
    "UpdateUserType": false,
    "UserRoleAction": "Create",
    "SQLConf": {
        "Driver": "mysql",
        "Server": "localhost",
        "Database": "swdata",
        "UserName": "root",
        "Password": "rootpwd",
        "Port": 5002,
        "UserID": "keysearch",
        "Encrypt": false,
        "Query": "SELECT userdb.*, b.fullname AS manager FROM userdb LEFT JOIN userdb b ON b.keysearch = userdb.fk_manager"
    },
    "UserMapping":{
        "userId":"{{.keysearch}}",
        "UserType":"basic",
        "Name":"{{.firstname}} {{.surname}}",
        "Password":"",
        "FirstName":"{{.firstname}}",
        "LastName":"{{.surname}}",
        "JobTitle":"",
        "Site":"{{.site}}",
        "Phone":"{{.telext}}",
        "Email":"{{.email}}",
        "Mobile":"[mobile]",
        "AbsenceMessage":"",
        "TimeZone":"",
        "Language":"",
        "DateTimeFormat":"",
        "DateFormat":"",
        "TimeFormat":"",
        "CurrencySymbol":"",
        "CountryCode":""
    },
    "UserAccountStatus":{
        "Action":"Both",
        "Enabled": false,
        "Status":"active"
    },
    "UserProfileMapping":{
        "MiddleName":"",
        "JobDescription":"",
        "Manager":"{{.manager}}",
        "WorkPhone":"",
        "Qualifications":"",
        "Interests":"",
        "Expertise":"",
        "Gender":"",
        "Dob":"",
        "Nationality":"",
        "Religion":"",
        "HomeTelephone":"{{.telext}}",
        "SocialNetworkA":"",
        "SocialNetworkB":"",
        "SocialNetworkC":"",
        "SocialNetworkD":"",
        "SocialNetworkE":"",
        "SocialNetworkF":"",
        "SocialNetworkG":"",
        "SocialNetworkH":"",
        "PersonalInterests":"",
        "homeAddress":"",
        "PersonalBlog":"",
        "Attrib1":"1",
    	"Attrib2":"2",
    	"Attrib3":"3",
    	"Attrib4":"4",
    	"Attrib5":"5",
    	"Attrib6":"6",
    	"Attrib7":"7",
    	"Attrib8":"8"
    }
    , "UserManagerMapping":{
        "Action":"Both"
        , "Enabled":true
    }
    , "Roles":[
        "Basic User Role"
    ]
    , "SiteLookup":{
        "Action":"Both"
        , "Enabled": true
    }
    , "ImageLink":{
        "Action":"Both"
        , "Enabled": true
        , "UploadType": "URL"
        , "ImageType": "jpg"
        , "URI": "http://sample.myservicedesk.com/sw/clisupp/documents/userdb/images/{{.keysearch}}.jpg"
    }
    , "OrgLookup":{
        "Action":"Both"
        , "Enabled":true
        , "OrgUnits":[
            {
                "Attribute":"{{.department}}",
                "Type":2,
                "Membership":"member",
                "TasksView":false,
                "TasksAction":false
            }, {
                "Attribute":"{{.companyname}}",
                "Type":5,
                "Membership":"member",
                "TasksView":false,
                "TasksAction":false
            }
        ]
    }
}
```
#### InstanceConfig
* "APIKey" - A Valid API Assigned to a user with enough rights to process the import
* "InstanceId" - Instance Id
* "UpdateUserType" - If set to True then the Type of User will be updated when the user account Update is triggered
* "UserRoleAction" - (Both | Update | Create) - When to Set controls what action will assign roles ro a user Create, On Update or Both

#### SQLConf
* "Driver" the driver to use to connect to the database that holds the asset information:
** mssql = Microsoft SQL Server (2005 or above)
** mysql = MySQL Server 4.1+, MariaDB
** mysql320 = MySQL Server v3.2.0 to v4.0
** swsql = Supportworks SQL (Core Services v3.x)
* "Server" The address of the SQL server
* "UserName" The username for the SQL database
* "Password" Password for above User Name
* "Port" SQL port
* "UserID" Specifies the unique identifier field from the query below
* "Encrypt" Boolean value to specify wether the connection between the script and the database should be encrypted. ''NOTE'': There is a bug in SQL Server 2008 and below that causes the connection to fail if the connection is encrypted. Only set this to true if your SQL Server has been patched accordingly.
* "Query" The basic SQL query to retrieve asset information from the data source. See "AssetTypes below for further filtering

#### UserMapping
* Any value formatted with {.field} will be treated as the value contained in the field of the query result record
* Do not try and add any new properties here they will be ignored
* Any Other Value is treated literally as written example:
    * "Name":"{.firstname} {.surname}", - Both Variables are evaluated from the database and set to the Name param
    * "Password":"", - Auto Generated Password (only on insert)
* If Password is left empty then a 10 character random string will be assigned so the user will need to recover the password using forgot my password functionality - The password will also be in the Log File
* "UserType" - This defines if a user is Co-Worker or Basic user and can have the value user or basic.


#### UserAccountStatus
* Action - (Both | Update | Create) - When to Set the User Account Status On Create, On Update or Both
* Enabled - Turns on or off the Status update
* Status - Can be one of the following strings (active | suspended | archived)

#### UserProfileMapping
* Works in the same way as UserMapping
* Do not try and add any new properties here they will be ignored

#### UserManagerMapping
This assumes that the Manager value set up under UserProfileMapping contains the Manager's unique identifier.
* Action - (Both | Update | Create) - When to Set the User Manager On Create, On Update or Both
* Enabled - Turns on or off the Manager Import

#### Roles
This should contain an array of roles to be added to a user when they are created. If importing Basic Users then only the '''Basic User Role''' should be specified any role with a User Privilege level will be rejected

#### SiteLookup
This assumes that the Site value set up under UserMapping contains the Site's unique identifier.

* Action - (Both | Update | Create) - When to Associate Sites On Create, On Update or Both
* Enabled - Turns on or off the Lookup of Sites
* Attribute - The LDAP Attribute to use for the name of the Site ,Any value wrapped with [] will be treaded ad an LDAP field

#### ImageLookup
* Action - (Both | Update | Create) - When to Associate Images On Create, On Update or Both
* Enabled - Turns on or off the Lookup of Images
* UploadType - This can be one of three values
** URI: it will take the URI and pass it on to the Hornbill Server - the Hornbill Server will then pick up the file (and Hornbill is expected to be able to reach the file)
** URL: this will have the executable itself ATTEMPT the upload of the file to which the URI points (eg this file needs to be accessible from the LOCAL machine) _experimental_
** anything else: this will grab the data as a binary string (eg 0x12BE...) and upload that _experimental_
* ImageType - currently NOT in use - will be used in conjunction with UploadType to arrange for data manipulation - currently all is hard-coded towards the example given.
* URI - usage as defined in the UploadType definition



#### OrgLookup
The name of each of the relevant Organisational data group types in Hornbill must match the value of the Attribute in the Database results.
One can link up with various organisational entities here. Globally:
* Action - (Both | Update | Create) - When to Associate Organisation On Create, On Update or Both
* Enabled - Turns on or off the Lookup of Orgnisations
For each of the organisational group types:
* Attribute - The LDAP Attribute to use for the name of the Site ,Any value wrapped with [] will be treaded ad an LDAP field
* Type - The Organisation Type (0=general ,1=team ,2=department ,3=costcenter ,4=division ,5=company)
* Membership - The Organisation Membership the users will be added with (member,teamLeader,manager)
* TasksView - If set true, then the user can view tasks assigned to this group
* TasksAction - If set true, then the user can action tasks assigned to this group.

# execute
Command Line Parameters
* file - Defaults to `conf.json` - Name of the Configuration file to load
* dryrun - Defaults to `false` - Set to True and the XMLMC for Create and Update users will not be called and instead the XML will be dumped to the log file, this is to aid in debugging the initial connection information.
* logprefix - Default to `` - Allows you to define a string to prepend to the name of the log file generated
* zone - Defaults to `eur` - Allows you to change the ZONE used for creating the XMLMC EndPoint URL https://{ZONE}api.hornbill.com/{INSTANCE}/
* workers - Defaults to `3` - Allows you to change the number of worker threads used to process the import, this can improve performance on slow import but using too many workers have a detriment to performance of your Hornbill instance.

# proxy
If one is using a proxy server, please ensure that the __HTTP_PROXY__ environment variable is configured before the goDb2HUserImport.exe is run; eg:

'SET HTTP_PROXY=10.123.123.123:8080'

# Testing
If you run the application with the argument dryrun=true then no users will be created or updated, the XML used to create or update will be saved in the log file so you can ensure the LDAP mappings are correct before running the import.

'goDb2HUserImport.exe -dryrun=true'


# Scheduling

### Windows
You can schedule goDb2HUserImport.exe to run with any optional command line argument from Windows Task Scheduler.
* Ensure the user account running the task has rights to goDb2HUserImport.exe and the containing folder.
* Make sure the Start In parameter contains the folder where goDb2HUserImport.exe resides in otherwise it will not be able to pick up the correct path.

# logging
All Logging output is saved in the log directory in the same directory as the executable the file name contains the date and time the import was run 'SQL_User_Import_2015-11-06T14-26-13Z.log'

# Error Codes
* `100` - Unable to create log File
* `101` - Unable to create log folder
* `102` - Unable to Load Configuration File
