# CHANGELOG

## 2.2.7 (August 20th, 2021)

Change:

- Minor (typo) corrections

## 2.2.6 (July 6th, 2021)

Change:

- Rebuilt using latest version of goApiLib, to fix possible issue with connections via a proxy
- Fixed up some linter warnings 

## 2.2.5 (February 22, 2021)

Fixes:

- fix to Create/Update functionality which prevented certain updates.
- sample configuration file had superfluous comma which would produce an error
- tidy up of code

## 2.2.4 (February 5th, 2021)

Changes:

- ability to import configure to only Create or Update

Fixes:

- fix to manager search which had search on manager name hard-coded instead of picking up the configured field (from Manager.Options.Search.SearchField).

## 2.2.3 (December 4th, 2020)

Changes:

- ability to import on employee or logon ID instead

## 2.2.2 (June 12th, 2020)

Changes:

- minor changes to be compatible with new crosscompile script

## 2.2.1 (April 15th, 2020)

Change:

- Updated code to support Core application and platform changes

## 2.2.0 (January 9th, 2020)

Changes:

- Added support for new Employee ID field in user record

## 2.1.0 (November 15th, 2019)

Changes:

- Added support for new Login ID field in user record

## 2.0.1 (October 18th, 2019)

Changes:

- Added feature to allow the setting of a Home Organisation when creating/updating users

## 2.0.0 (October 18th, 2019)

Features:

- Reworking to match LDAP imports - but with local configuration file
- PLEASE NOTE the CONFIGURATION file has changed significantly.

## 1.2.3 (September 26th, 2018)

Fixes:

- Recoding to use entityBrowseRecords2 instead of entityBrowseRecords.
