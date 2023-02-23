# CHANGELOG

## 3.2.0 (Feburary, 23rd, 2022)

Feature:

- Enabled option to upload User images from the local filesystem or network shares.

## 3.1.0 (Feburary, 21st, 2022)

Change:

- Complied with latest Go binaries because of security advisory.

## 3.0.0 (December 21st, 2022)

Fix:

- Issue with tool attempting to perform unnecessary updates when there are more than 65k records returned from the DB

Features:

- Added ability for tool to self-update when a minor or patch update is available
- Added support for setting the following new security settings against imported users:
  - 2 Factor Authentication
  - Disable direct login
  - Disable direct login password reset
  - Disable device pairing on user profile

Change:

- Removed references to deprecated ioutil, replaced with io
- Removed API Key output to CLI when running tool

## 2.3.3 (November 18th, 2022)

Fix:

- The way empty strings get processed

## 2.3.2 (July 8th, 2022)

Fix:

- Fixed issue where Hornbill users were not being cached, so could not be updated

## 2.3.1 (December 23rd, 2021)

Change:

- User.Status.Value can now be based on a DB result

## 2.3.0 (December 10th, 2021)

Change:

- Improved performance when caching user account groups

## 2.2.9 (December 7th, 2021)

Fix:

- Issue with attributes not comparing correctly

## 2.2.8 (September 22nd, 2021)

Change:

- Modification such that updates do not require a Name and userType set

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
