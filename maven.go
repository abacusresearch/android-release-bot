package main

import (
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "strings"
)

func downloadMavenArtifact(url string) *os.File {
    client := &http.Client{}

    request, err := http.NewRequest("GET", url, nil)

    if err != nil {
        postSlackMessage("Sorry, I can't create the HTTP request: %v", err)
        return nil
    }

    request.SetBasicAuth(getConfig("MAVEN_ACCOUNT_NAME"), getConfig("MAVEN_ACCOUNT_PASSWORD"))

    response, err := client.Do(request)

    if err != nil {
        postSlackMessage("Sorry, I can't execute the HTTP request: %v", err)
        return nil
    }

    defer response.Body.Close()

    if response.StatusCode != 200 {
        postSlackMessage("Sorry, I didn't expect that HTTP status code: %v", response.StatusCode)
        return nil
    }

    result, err := ioutil.TempFile("", "")

    if err != nil {
        postSlackMessage("Sorry, I can't create the temporary file: %v", err)
        return nil
    }

    _, err = io.Copy(result, response.Body)

    if err != nil {
        postSlackMessage("Sorry, I can't write the temporary file: %v", err)
        return nil
    }

    _, err = result.Seek(0, 0)

    if err != nil {
        postSlackMessage("Sorry, I can't seek in the temporary file: %v", err)
        return nil
    }

    return result
}

func locateMavenArtifact(artifactId string, version string) string {
    var result strings.Builder

    artifactId = url.PathEscape(artifactId)
    version = url.PathEscape(version)

    result.WriteString(getConfig("MAVEN_REPOSITORY"))
    result.WriteString(strings.Replace(getConfig("MAVEN_GROUP_ID"), ".", "/", -1))
    result.WriteString("/")
    result.WriteString(artifactId)
    result.WriteString("/")
    result.WriteString(version)
    result.WriteString("/")
    result.WriteString(artifactId + "-" + version + ".apk")

    return result.String()
}
