package chatbuilder

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

type (
	timeRange [2]int
	greeting  struct {
		text       string
		ascii      string
		days       []string
		timeRanges timeRange
	}
)

var greetings = []greeting{
	{
		text: "Good morning",
		ascii: `
█▀▀ █▀█ █▀█ █▀▄   █▀▄▀█ █▀█ █▀█ █▄░█ █ █▄░█ █▀▀
█▄█ █▄█ █▄█ █▄▀   █░▀░█ █▄█ █▀▄ █░▀█ █ █░▀█ █▄█`,
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Welcome",
		ascii: `
█░█░█ █▀▀ █░░ █▀▀ █▀█ █▀▄▀█ █▀▀
▀▄▀▄▀ ██▄ █▄▄ █▄▄ █▄█ █░▀░█ ██▄`,
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Hey there",
		ascii: `
█░█ █▀▀ █▄█   ▀█▀ █░█ █▀▀ █▀█ █▀▀
█▀█ ██▄ ░█░   ░█░ █▀█ ██▄ █▀▄ ██▄`,
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Monday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   █▀▄▀█ █▀█ █▄░█ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   █░▀░█ █▄█ █░▀█ █▄▀ █▀█ ░█░`,
		days:       []string{"Monday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Tuesday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   ▀█▀ █░█ █▀▀ █▀ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   ░█░ █▄█ ██▄ ▄█ █▄▀ █▀█ ░█░`,
		days:       []string{"Tuesday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Wednesday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   █░█░█ █▀▀ █▀▄ █▄░█ █▀▀ █▀ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   ▀▄▀▄▀ ██▄ █▄▀ █░▀█ ██▄ ▄█ █▄▀ █▀█ ░█░`,
		days:       []string{"Wednesday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Thursday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   ▀█▀ █░█ █░█ █▀█ █▀ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   ░█░ █▀█ █▄█ █▀▄ ▄█ █▄▀ █▀█ ░█░`,
		days:       []string{"Thursday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Friday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   █▀▀ █▀█ █ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   █▀░ █▀▄ █ █▄▀ █▀█ ░█░`,
		days:       []string{"Friday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "That Friday feeling",
		ascii: `
▀█▀ █░█ ▄▀█ ▀█▀   █▀▀ █▀█ █ █▀▄ ▄▀█ █▄█   █▀▀ █▀▀ █▀▀ █░░ █ █▄░█ █▀▀
░█░ █▀█ █▀█ ░█░   █▀░ █▀▄ █ █▄▀ █▀█ ░█░   █▀░ ██▄ ██▄ █▄▄ █ █░▀█ █▄█`,
		days:       []string{"Friday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Welcome to the weekend",
		ascii: `
█░█░█ █▀▀ █░░ █▀▀ █▀█ █▀▄▀█ █▀▀   ▀█▀ █▀█   ▀█▀ █░█ █▀▀   █░█░█ █▀▀ █▀▀ █▄▀ █▀▀ █▄░█ █▀▄
▀▄▀▄▀ ██▄ █▄▄ █▄▄ █▄█ █░▀░█ ██▄   ░█░ █▄█   ░█░ █▀█ ██▄   ▀▄▀▄▀ ██▄ ██▄ █░█ ██▄ █░▀█ █▄▀`,
		days:       []string{"Saturday", "Sunday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Saturday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   █▀ ▄▀█ ▀█▀ █░█ █▀█ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   ▄█ █▀█ ░█░ █▄█ █▀▄ █▄▀ █▀█ ░█░`,
		days:       []string{"Saturday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "What’s on your mind?",
		ascii: `
█░█░█ █░█ ▄▀█ ▀█▀ ▀ █▀   █▀█ █▄░█   █▄█ █▀█ █░█ █▀█   █▀▄▀█ █ █▄░█ █▀▄ ▀█
▀▄▀▄▀ █▀█ █▀█ ░█░ ░ ▄█   █▄█ █░▀█   ░█░ █▄█ █▄█ █▀▄   █░▀░█ █ █░▀█ █▄▀ ░▄`,
		days:       []string{"Saturday", "Sunday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Sunday session?",
		ascii: `
█▀ █░█ █▄░█ █▀▄ ▄▀█ █▄█   █▀ █▀▀ █▀ █▀ █ █▀█ █▄░█ ▀█
▄█ █▄█ █░▀█ █▄▀ █▀█ ░█░   ▄█ ██▄ ▄█ ▄█ █ █▄█ █░▀█ ░▄`,
		days:       []string{"Sunday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Happy Sunday",
		ascii: `
█░█ ▄▀█ █▀█ █▀█ █▄█   █▀ █░█ █▄░█ █▀▄ ▄▀█ █▄█
█▀█ █▀█ █▀▀ █▀▀ ░█░   ▄█ █▄█ █░▀█ █▄▀ █▀█ ░█░`,
		days:       []string{"Sunday"},
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Coffee and Claude time?",
		ascii: `
█▀▀ █▀█ █▀▀ █▀▀ █▀▀ █▀▀   ▄▀█ █▄░█ █▀▄   █▀▀ █░░ ▄▀█ █░█ █▀▄ █▀▀   ▀█▀ █ █▀▄▀█ █▀▀ ▀█
█▄▄ █▄█ █▀░ █▀░ ██▄ ██▄   █▀█ █░▀█ █▄▀   █▄▄ █▄▄ █▀█ █▄█ █▄▀ ██▄   ░█░ █ █░▀░█ ██▄ ░▄`,
		timeRanges: timeRange{6, 12},
	},
	{
		text: "Good afternoon",
		ascii: `
█▀▀ █▀█ █▀█ █▀▄   ▄▀█ █▀▀ ▀█▀ █▀▀ █▀█ █▄░█ █▀█ █▀█ █▄░█
█▄█ █▄█ █▄█ █▄▀   █▀█ █▀░ ░█░ ██▄ █▀▄ █░▀█ █▄█ █▄█ █░▀█`,
		timeRanges: timeRange{12, 17},
	},
	{
		text: "Hi, how are you?",
		ascii: `
█░█ █ ░   █░█ █▀█ █░█░█   ▄▀█ █▀█ █▀▀   █▄█ █▀█ █░█ ▀█
█▀█ █ █   █▀█ █▄█ ▀▄▀▄▀   █▀█ █▀▄ ██▄   ░█░ █▄█ █▄█ ░▄`,
		timeRanges: timeRange{12, 17},
	},
	{
		text: "What’s new?",
		ascii: `
█░█░█ █░█ ▄▀█ ▀█▀ ▀ █▀   █▄░█ █▀▀ █░█░█ ▀█
▀▄▀▄▀ █▀█ █▀█ ░█░ ░ ▄█   █░▀█ ██▄ ▀▄▀▄▀ ░▄`,
		timeRanges: timeRange{12, 17},
	},
	{
		text: "Back at it",
		ascii: `
█▄▄ ▄▀█ █▀▀ █▄▀   ▄▀█ ▀█▀   █ ▀█▀
█▄█ █▀█ █▄▄ █░█   █▀█ ░█░   █ ░█░`,
		timeRanges: timeRange{12, 17},
	},
	{
		text: "Good evening",
		ascii: `
█▀▀ █▀█ █▀█ █▀▄   █▀▀ █░█ █▀▀ █▄░█ █ █▄░█ █▀▀
█▄█ █▄█ █▄█ █▄▀   ██▄ ▀▄▀ ██▄ █░▀█ █ █░▀█ █▄█`,
		timeRanges: timeRange{17, 21},
	},
	{
		text: "Evening",
		ascii: `
█▀▀ █░█ █▀▀ █▄░█ █ █▄░█ █▀▀
██▄ ▀▄▀ ██▄ █░▀█ █ █░▀█ █▄█`,
		timeRanges: timeRange{17, 21},
	},
	{
		text: "How was your day?",
		ascii: `
█░█ █▀█ █░█░█   █░█░█ ▄▀█ █▀   █▄█ █▀█ █░█ █▀█   █▀▄ ▄▀█ █▄█ ▀█
█▀█ █▄█ ▀▄▀▄▀   ▀▄▀▄▀ █▀█ ▄█   ░█░ █▄█ █▄█ █▀▄   █▄▀ █▀█ ░█░ ░▄`,
		timeRanges: timeRange{17, 21},
	},
	{
		text: "How’s it going?",
		ascii: `
█░█ █▀█ █░█░█ ▀ █▀   █ ▀█▀   █▀▀ █▀█ █ █▄░█ █▀▀ ▀█
█▀█ █▄█ ▀▄▀▄▀ ░ ▄█   █ ░█░   █▄█ █▄█ █ █░▀█ █▄█ ░▄`,
		timeRanges: timeRange{21, 6},
	},
	{
		text: "What’s on your mind tonight?",
		ascii: `
█░█░█ █░█ ▄▀█ ▀█▀ ▀ █▀   █▀█ █▄░█   █▄█ █▀█ █░█ █▀█   █▀▄▀█ █ █▄░█ █▀▄   ▀█▀ █▀█ █▄░█ █ █▀▀ █░█ ▀█▀ ▀█
▀▄▀▄▀ █▀█ █▀█ ░█░ ░ ▄█   █▄█ █░▀█   ░█░ █▄█ █▄█ █▀▄   █░▀░█ █ █░▀█ █▄▀   ░█░ █▄█ █░▀█ █ █▄█ █▀█ ░█░ ░▄`,
		timeRanges: timeRange{21, 6},
	},
	{
		text: "Hello, night owl",
		ascii: `
█░█ █▀▀ █░░ █░░ █▀█ ░   █▄░█ █ █▀▀ █░█ ▀█▀   █▀█ █░█░█ █░░
█▀█ ██▄ █▄▄ █▄▄ █▄█ █   █░▀█ █ █▄█ █▀█ ░█░   █▄█ ▀▄▀▄▀ █▄▄`,
		timeRanges: timeRange{21, 6},
	},
}

func getGretting(width int) string {
	g := greetingFor(time.Now())
	if g == nil {
		return ""
	}
	return g.render(width)
}

func greetingFor(now time.Time) *greeting {
	today := now.Weekday().String()

	var generic *greeting
	for i := range greetings {
		g := &greetings[i]
		if !inRange(now.Hour(), g.timeRanges) {
			continue
		}
		if len(g.days) > 0 {
			if slices.Contains(g.days, today) {
				return g
			}
			continue
		}
		if generic == nil {
			generic = g
		}
	}
	return generic
}

func (g greeting) render(width int) string {
	if g.ascii != "" && fits(g.ascii, width) {
		return g.ascii
	}
	return g.text
}

func inRange(hour int, r timeRange) bool {
	if r[0] <= r[1] {
		return hour >= r[0] && hour < r[1]
	}
	return hour >= r[0] || hour < r[1]
}

func fits(s string, width int) bool {
	for line := range strings.Lines(s) {
		if utf8.RuneCountInString(strings.TrimRight(line, "\n")) > width {
			return false
		}
	}
	return true
}
