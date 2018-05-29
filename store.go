package main

import (
    "encoding/base64"
    "golang.org/x/oauth2/google"
    "golang.org/x/oauth2/jwt"
    "google.golang.org/api/androidpublisher/v2"
)

func addVersionCodeToStoreTrack(
        publisher *androidpublisher.Service,
        edit *androidpublisher.AppEdit,
        track *androidpublisher.Track,
        appId string,
        appVersionCode int64,
        userFraction float64) bool {
    postSlackMessage("Adding version code *%v* to track *%v*.", appVersionCode, track.Track)

    track.UserFraction = userFraction
    track.VersionCodes = append(track.VersionCodes, appVersionCode)

    _, err := publisher.Edits.Tracks.
            Update(appId, edit.Id, track.Track, track).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot update the track: %v", err)
        return false
    }

    return true
}

func changeUserFraction(
        publisher *androidpublisher.Service,
        edit *androidpublisher.AppEdit,
        track *androidpublisher.Track,
        appId string,
        userFraction float64) bool {
    postSlackMessage("Changing user fraction for track *%v*.", track.Track)

    track.UserFraction = userFraction

    _, err := publisher.Edits.Tracks.
            Update(appId, edit.Id, track.Track, track).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot update the track: %v", err)
        return false
    }

    return true
}

func loadStoreCredentials() *jwt.Config {
    data, err := base64.StdEncoding.DecodeString(getConfig("ANDROID_PUBLISHER_CREDENTIALS"))

    if err != nil {
        postSlackMessage("Sorry, I cannot decode the credentials: %v", err)
        return nil
    }

    result, err := google.JWTConfigFromJSON(
            data,
            "https://www.googleapis.com/auth/androidpublisher")

    if err != nil {
        postSlackMessage("Sorry, I cannot parse the credentials: %v", err)
        return nil
    }

    return result
}

func removeAllVersionCodesFromStoreTrack(
        publisher *androidpublisher.Service,
        edit *androidpublisher.AppEdit,
        track *androidpublisher.Track,
        appId string) bool {
    for _, versionCode := range track.VersionCodes {
        postSlackMessage("Removing version code *%v* from track *%v*.", versionCode, track.Track)
    }

    track.VersionCodes = []int64 {}

    _, err := publisher.Edits.Tracks.
            Update(appId, edit.Id, track.Track, track).
            Do()

    if err != nil {
        postSlackMessage("Sorry, I cannot update the track: %v", err)
        return false
    }

    return true
}

func removeVersionCodeFromStoreTracks(
        publisher *androidpublisher.Service,
        edit *androidpublisher.AppEdit,
        tracks []*androidpublisher.Track,
        appId string,
        appVersionCode int64) bool {
    for _, track := range tracks {
        var appVersionCodes []int64

        for _, candidate := range track.VersionCodes {
            if candidate == appVersionCode {
                postSlackMessage("Removing version code *%v* from track *%v*.", candidate, track.Track)
            } else {
                appVersionCodes = append(appVersionCodes, candidate)
            }
        }

        track.VersionCodes = appVersionCodes

        _, err := publisher.Edits.Tracks.
                Update(appId, edit.Id, track.Track, track).
                Do()

        if err != nil {
            postSlackMessage("Sorry, I cannot update the track: %v", err)
            return false
        }
    }

    return true
}
