## goepay on App Engine

### Prerequisites
* Go appengine SDK
  https://developers.google.com/appengine/downloads#Google_App_Engine_SDK_for_Go

* Godoc sources at go-epay inside $GOPATH
  (go get -d github.com/clouway/go-epay)

### Deployment
```
gcloud --project yourprojectname app deploy --no-promote app.yaml
```
