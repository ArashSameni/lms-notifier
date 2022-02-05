package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/0x434d53/openinbrowser"
	"github.com/ArashSameni/lms-notifier/login"
	"github.com/anaskhan96/soup"
)

type Stream struct {
	PageSource string
	Course     *Course
}

type Course struct {
	Name, Link string
	Diffs      []Activity
}

type Activity string

func extractCourses(page string) []Course {
	doc := soup.HTMLParse(page)
	coursesLI := doc.Find("div", "id", "nav-drawer").FindAll("li")

	courses := make([]Course, 0, len(coursesLI))

	for _, c := range coursesLI {
		a := c.Find("a")
		if a.Error != nil {
			continue
		}

		link := a.Attrs()["href"]
		if !strings.Contains(link, "course") {
			continue
		}

		name := c.Find("span", "class", "media-body").Text()
		courses = append(courses, Course{Name: name, Link: link})
	}
	return courses
}

func oldActivities(fileName string) map[Activity]bool {
	old := map[Activity]bool{}

	file, _ := os.Open(fileName)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		old[Activity(scanner.Text())] = true
	}

	return old
}

func getCoursesPage(client *http.Client, courses []Course, downstream chan<- Stream) {
	for i, course := range courses {
		resp, err := client.Get(course.Link)
		if err != nil {
			fmt.Println("Couldn't Get " + course.Name)
			continue
		}
		defer resp.Body.Close()

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Couldn't Get " + course.Name)
			continue
		}

		downstream <- Stream{string(buf), &courses[i]}
	}
	close(downstream)
}

func extractActivities(upstream <-chan Stream) {
	for s := range upstream {
		fileName := "sources/" + s.Course.Name + ".txt"
		doc := soup.HTMLParse(s.PageSource)
		activitiesSpan := doc.FindAll("span", "class", "instancename")
		old := oldActivities(fileName)
		diffs := []Activity{}
		file, _ := os.Create(fileName)
		for _, a := range activitiesSpan {
			if !old[Activity(a.Text())] {
				diffs = append(diffs, Activity(a.Text()))
			}
			fmt.Fprintln(file, a.Text())
		}
		file.Close()
		s.Course.Diffs = diffs
	}
}

func createChangesFile(courses []Course) {
	file, _ := os.Create("changes.html")
	t, _ := template.ParseFiles("template.html")
	t.Execute(file, struct {
		Time    string
		Courses []Course
	}{time.Now().Format("2006-01-02 15:04:05 Mon"), courses})
}

func main() {
	os.Mkdir("sources", os.ModePerm)

	url := flag.String("url", "https://webauth.iut.ac.ir/cas/login?service=https://yekta.iut.ac.ir/login/index.php?authCASattras=CASattras", "LMS login page")
	username := flag.String("u", "", "Username")
	password := flag.String("p", "", "Password")

	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("Please enter username and password(-help for more info)")
		return
	}

	client, source, err := login.Login(*url, *username, *password)
	if err != nil {
		fmt.Println(err)
		return
	}

	stream := make(chan Stream)

	courses := extractCourses(source)
	go getCoursesPage(client, courses, stream)
	extractActivities(stream)
	createChangesFile(courses)
	openinbrowser.Open("changes.html")
}
