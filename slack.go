package main

import (
    "fmt"
    "github.com/nlopes/slack"
    "log"
    "regexp"
    "strconv"
    "strings"
)

var rtm *slack.RTM

func handleSlackMessage(event *slack.MessageEvent) {
    text := event.Msg.Text

    if len(event.User) == 0 {
        log.Printf("#%v %v", event.Channel, text)
    } else {
        log.Printf("#%v %v: %v", event.Channel, event.User, text)
    }

    if event.Channel != getConfig("SLACK_BOT_CHANNEL_ID") {
        return
    }

    textPrefix := fmt.Sprintf("<@%s>", getConfig("SLACK_BOT_USER_ID"))

    if !strings.HasPrefix(text, textPrefix) {
        return
    }

    // Handle the 'deploy' command.

    command := regexp.
            MustCompile("<[^>]+> +deploy +([^ ]+) +([^ ]+)").
            FindStringSubmatch(text)

    if len(command) > 0 {
        doDeploy(command[1], command[2])
        return
    }

    // Handle the 'halt' command.

    command = regexp.
            MustCompile("<[^>]+> +halt +([^ ]+) +([^ ]+)").
            FindStringSubmatch(text)

    if len(command) > 0 {
        appVersionCode, err := strconv.ParseInt(command[2], 10, 64)

        if err != nil {
            postSlackMessage("Sorry, I don't understand that version code.")
            return
        }

        doHalt(command[1], appVersionCode)
        return
    }

    // Handle the 'ping' command.

    command = regexp.
            MustCompile("<[^>]+> +ping").
            FindStringSubmatch(text)

    if len(command) > 0 {
        doPing()
        return
    }

    // Handle the 'promote' command.

    command = regexp.
            MustCompile("<[^>]+> +promote +([^ ]+) +([^ ]+) +to +(.*)").
            FindStringSubmatch(text)

    if len(command) > 0 {
        appVersionCode, err := strconv.ParseInt(command[2], 10, 64)

        if err != nil {
            postSlackMessage("Sorry, I don't understand that version code.")
            return
        }

        storeTrack := command[3]

        if storeTrack != "alpha" && storeTrack != "beta" && storeTrack != "internal" {
            if !getConfigExpression("SLACK_GOD_USER_ID").MatchString(event.User) {
                postSlackMessage("Sorry, only gods can do that.")
                return
            }
        }

        doPromote(command[1], appVersionCode, command[3])
        return
    }

    // Handle the 'rollout' command.

    command = regexp.
            MustCompile("<[^>]+> +rollout +([^ ]+) +([^ ]+) +to +(.*)%").
            FindStringSubmatch(text)

    if len(command) > 0 {
        appVersionCode, err := strconv.ParseInt(command[2], 10, 64)

        if err != nil {
            postSlackMessage("Sorry, I don't understand that version code.")
            return
        }

        userPercentage, err := strconv.Atoi(command[3])

        if err != nil {
            postSlackMessage("Sorry, I don't understand that user percentage.")
            return
        }

        if !getConfigExpression("SLACK_GOD_USER_ID").MatchString(event.User) {
            postSlackMessage("Sorry, only gods can do that.")
            return
        }

        doRollout(command[1], appVersionCode, userPercentage)
        return
    }

    // Handle the 'show release notes' command.

    command = regexp.
            MustCompile("<[^>]+> +show +release +notes +for +([^ ]+) +(.+)").
            FindStringSubmatch(text)

    if len(command) > 0 {
        appVersionCode, err := strconv.ParseInt(command[2], 10, 64)

        if err != nil {
            postSlackMessage("Sorry, I don't understand that version code.")
            return
        }

        doShowReleaseNotes(command[1], appVersionCode)
        return
    }

    // Handle the 'show tracks' command.

    command = regexp.
            MustCompile("<[^>]+> +show +tracks +for +(.+)").
            FindStringSubmatch(text)

    if len(command) > 0 {
        doShowTracks(command[1])
        return
    }

    doHelp()
}

func handleSlackMessages() {
    client := slack.New(getConfig("SLACK_BOT_TOKEN"))

    rtm = client.NewRTM()

    go rtm.ManageConnection()

    for event := range rtm.IncomingEvents {
        switch typedEvent := event.Data.(type) {
        case *slack.MessageEvent:
            handleSlackMessage(typedEvent)
        }
    }
}

func postSlackMessage(message string, arguments ...interface{}) {
    messageText := fmt.Sprintf(message, arguments...)

    _, _, err := rtm.PostMessage(
            getConfig("SLACK_BOT_CHANNEL_ID"),
            slack.MsgOptionText(messageText, false))

    if err != nil {
        panic(err)
    }
}
