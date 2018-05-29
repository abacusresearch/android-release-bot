package main

import (
    "fmt"
    "golang.org/x/oauth2"
    "google.golang.org/api/androidpublisher/v2"
    "google.golang.org/api/googleapi"
    "os"
)

func doDeploy(artifactId string, version string) {
    postSlackMessage("Ok, deploying *%v* with version *%v* ...", artifactId, version)

    artifactUrl := locateMavenArtifact(artifactId, version)
    artifactFile := downloadMavenArtifact(artifactUrl)

    if artifactFile == nil {
        return
    }

    defer os.Remove(artifactFile.Name())

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher.New(client)

    if err != nil {
        postSlackMessage("Sorry, I cannot create the publisher: %v", err)
        return
    }

    appId := fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), artifactId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot insert the edit: %v", err)
        return
    }

    apk, err := publisher.Edits.Apks.
            Upload(appId, edit.Id).
            Media(artifactFile, googleapi.ContentType("application/vnd.android.package-archive")).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot upload the APK: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot list the tracks: %v", err)
        return
    }

    track := &androidpublisher.Track {Track: "internal"}

    for _, candidate := range tracks.Tracks {
        if (candidate.Track == "internal") {
            track = candidate
        }
    }

    // Remove the lower versions from the target track.

    if !removeAllVersionCodesFromStoreTrack(publisher, edit, track, appId) {
        return
    }

    // Add the current version to the target track.

    if !addVersionCodeToStoreTrack(publisher, edit, track, appId, apk.VersionCode, 0) {
        return
    }

    _, err = publisher.Edits.
            Commit(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot commit the edit: %v", err)
        return
    }

    postSlackMessage("Done.")
}

func doHalt(appId string, appVersionCode int64) {
    postSlackMessage("Ok, halting *%v* with version code *%v* ...", appId, appVersionCode)

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher.New(client)

    if err != nil {
        postSlackMessage("Sorry, I cannot create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot list the tracks: %v", err)
        return
    }

    // Remove the version from all tracks.

    if !removeVersionCodeFromStoreTracks(publisher, edit, tracks.Tracks, appId, appVersionCode) {
        return
    }

    _, err = publisher.Edits.
            Commit(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot commit the edit: %v", err)
        return
    }

    postSlackMessage("Done.")
}

func doHelp() {
    postSlackMessage("Sorry, I don't understand.")
}

func doPing() {
    postSlackMessage("Pong.")
}

func doPromote(appId string, appVersionCode int64, storeTrack string) {
    postSlackMessage("Ok, promoting *%v* with version code *%v* to track *%v* ...", appId, appVersionCode, storeTrack)

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher.New(client)

    if err != nil {
        postSlackMessage("Sorry, I cannot create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot list the tracks: %v", err)
        return
    }

    track := &androidpublisher.Track {Track: storeTrack}

    for _, candidate := range tracks.Tracks {
        if (candidate.Track == storeTrack) {
            track = candidate;
        }
    }

    for _, candidate := range track.VersionCodes {
        if candidate == appVersionCode {
            postSlackMessage("Version code *%v* already exists in track *%v*.", appVersionCode, storeTrack)
            return
        }
    }

    // Remove all lower versions from the target track.

    if !removeAllVersionCodesFromStoreTrack(publisher, edit, track, appId) {
       return
    }

    // Move the current version to the target tracks.

    if !removeVersionCodeFromStoreTracks(publisher, edit, tracks.Tracks, appId, appVersionCode) {
        return
    }

    if !addVersionCodeToStoreTrack(publisher, edit, track, appId, appVersionCode, 0) {
        return
    }

    _, err = publisher.Edits.
            Commit(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot commit the edit: %v", err)
        return
    }

    postSlackMessage("Done.")
}

func doRollout(appId string, appVersionCode int64, userPercentage int) {
    postSlackMessage("Ok, rolling out *%v* with version code *%v* to *%v%%* ...", appId, appVersionCode, userPercentage)

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher.New(client)

    if err != nil {
        postSlackMessage("Sorry, I cannot create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot list the tracks: %v", err)
        return
    }

    track := &androidpublisher.Track {Track: "rollout"}

    for _, candidate := range tracks.Tracks {
        if (candidate.Track == "rollout") {
            track = candidate;
        }
    }

    exists := false

    for _, candidate := range track.VersionCodes {
        if candidate == appVersionCode {
            exists = true
        }
    }

    userFraction := float64(userPercentage) / 100

    if !exists {

        // Remove all lower versions from the target track.

        if !removeAllVersionCodesFromStoreTrack(publisher, edit, track, appId) {
            return
        }

        // Move the current version to the target tracks.

        if !removeVersionCodeFromStoreTracks(publisher, edit, tracks.Tracks, appId, appVersionCode) {
            return
        }

        if !addVersionCodeToStoreTrack(publisher, edit, track, appId, appVersionCode, userFraction) {
            return
        }
    } else {

        // Change the user fraction.

        if !changeUserFraction(publisher, edit, track, appId, userFraction) {
            return
        }
    }

    _, err = publisher.Edits.
            Commit(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot commit the edit: %v", err)
        return
    }

    postSlackMessage("Done.")
}

func doShowTracks(appId string) {
    postSlackMessage("Ok, showing tracks for *%v* ...", appId)

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher.New(client)

    if err != nil {
        postSlackMessage("Sorry, I cannot create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot list the tracks: %v", err)
        return
    }

    for _, track := range tracks.Tracks {
        if track.UserFraction == 0 {
            postSlackMessage("Track *%v* contains version codes *%v*.", track.Track, track.VersionCodes)
        } else {
            postSlackMessage("Track *%v* contains version codes *%v* at *%v%%*.", track.Track, track.VersionCodes, track.UserFraction * 100)
        }
    }

    postSlackMessage("Done.")
}

func main() {
    handleSlackMessages()
}
