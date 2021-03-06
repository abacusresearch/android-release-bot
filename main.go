package main

import (
    "fmt"
    "golang.org/x/oauth2"
    "google.golang.org/api/googleapi"
    "log"
    "os"

    androidpublisher2 "google.golang.org/api/androidpublisher/v2"
    androidpublisher3 "google.golang.org/api/androidpublisher/v3"
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

    publisher, err := androidpublisher2.New(client)

    if err != nil {
        postSlackMessage("Sorry, I can't create the publisher: %v", err)
        return
    }

    appId := fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), artifactId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't insert the edit: %v", err)
        return
    }

    apk, err := publisher.Edits.Apks.
            Upload(appId, edit.Id).
            Media(artifactFile, googleapi.ContentType("application/vnd.android.package-archive")).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't upload the APK: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't list the tracks: %v", err)
        return
    }

    track := &androidpublisher2.Track {Track: "internal"}

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
        postSlackMessage("Sorry, I can't commit the edit: %v", err)
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

    publisher, err := androidpublisher2.New(client)

    if err != nil {
        postSlackMessage("Sorry, I can't create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't list the tracks: %v", err)
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
        postSlackMessage("Sorry, I can't commit the edit: %v", err)
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

    publisher, err := androidpublisher2.New(client)

    if err != nil {
        postSlackMessage("Sorry, I can't create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't list the tracks: %v", err)
        return
    }

    track := &androidpublisher2.Track {Track: storeTrack}

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
        postSlackMessage("Sorry, I can't commit the edit: %v", err)
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

    publisher, err := androidpublisher2.New(client)

    if err != nil {
        postSlackMessage("Sorry, I can't create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't list the tracks: %v", err)
        return
    }

    track := &androidpublisher2.Track {Track: "rollout"}

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
        postSlackMessage("Sorry, I can't commit the edit: %v", err)
        return
    }

    postSlackMessage("Done.")
}

func doShowReleaseNotes(appId string, appVersionCode int64) {
    postSlackMessage("Ok, showing release notes for *%v* with version code *%v* ...", appId, appVersionCode)

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher3.New(client)

    if err != nil {
        postSlackMessage("Sorry, I can't create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't list the tracks: %v", err)
        return
    }

    exists := false

    for _, track := range tracks.Tracks {
        for _, release := range track.Releases {
            for _, candidate := range release.VersionCodes {
                if candidate != appVersionCode {
                    continue
                }

                exists = true

                for _, releaseNotes := range release.ReleaseNotes {
                    postSlackMessage("*%v*: %v.", releaseNotes.Language, releaseNotes.Text)
                }
            }
        }
    }

    if exists {
        postSlackMessage("Done.")
    } else {
        postSlackMessage("Sorry, I can't find that version code.")
    }
}

func doShowTracks(appId string) {
    postSlackMessage("Ok, showing tracks for *%v* ...", appId)

    credentials := loadStoreCredentials()

    if credentials == nil {
        return
    }

    client := credentials.Client(oauth2.NoContext)

    publisher, err := androidpublisher2.New(client)

    if err != nil {
        postSlackMessage("Sorry, I can't create the publisher: %v", err)
        return
    }

    appId = fmt.Sprintf("%v.%v", getConfig("ANDROID_APP_ID_PREFIX"), appId)

    edit, err := publisher.Edits.
            Insert(appId, nil).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't insert the edit: %v", err)
        return
    }

    tracks, err := publisher.Edits.Tracks.
            List(appId, edit.Id).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I can't list the tracks: %v", err)
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
    log.Print("Starting up ...")

    handleSlackMessages()

    log.Print("Shutting down ...")
}
