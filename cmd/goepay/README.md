## goepay on App Engine

### Prerequisites
* Go appengine SDK
  https://developers.google.com/appengine/downloads#Google_App_Engine_SDK_for_Go

* Godoc sources at go-epay inside $GOPATH
  (go get -d github.com/clouway/go-epay)

### Setting Up the Environment

The environment configuration is provided through the Environment entity in the datastore where
it's structure should be as follow:

1. Entity kind should be "Environment"
2. Key - "default" if app is appspot domain is used or name of your domain if different one is used: e.g "myepaygw.yourdomain.com" 
3. billingKey (type string) - the json key generated from the IAM console of TelcoNG
4. billingURL (type string) - the url of the billing, e.g https://cloud.telcong.com or of the provided testing environment

### Deployment
```
gcloud --project yourprojectname app deploy --no-promote app.yaml
```
