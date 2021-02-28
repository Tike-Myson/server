package server

import (
	"github.com/Tike-Myson/database"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type Error struct {
	ErrorStatus int
}

var ErrorResponse Error

func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {

	w.WriteHeader(status)

	t, err := template.ParseFiles("./html/error.html")
	if err != nil {
		log.Println(err.Error())
	}

	ErrorResponse.ErrorStatus = status

	t.Execute(w, ErrorResponse)

}

func Filter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/filter" {
		ErrorHandler(w, r, 404)
		return
	}
	t, err := template.ParseFiles("./html/filter.html")
	if err != nil {
		ErrorHandler(w, r, 500)
		return
	}

	switch r.Method {
	case "GET":
		t.Execute(w, database.PersonalPageInformation)
		return
	case "POST":
		startCreationDate := r.FormValue("startCD")
		endCreationDate := r.FormValue("endCD")
		startFirstAlbum := r.FormValue("startFA")
		if startFirstAlbum == "" {
			startFirstAlbum = "1900-01-01"
		}
		endFirstAlbum := r.FormValue("endFA")
		if endFirstAlbum == "" {
			endFirstAlbum = "2021-01-01"
		}
		location := r.FormValue("location-filter")
		var membersCount []int
		for i := 0; i < 8; i++ {
			key := "mem" + strconv.Itoa(i+1)
			mem := r.FormValue(key)
			if mem == "" {
				continue
			}
			memInt, _ := strconv.Atoi(mem)
			membersCount = append(membersCount, memInt)
		}

		database.GetFilterInformation(startCreationDate, endCreationDate, startFirstAlbum, endFirstAlbum, location, membersCount)
		t, err := template.ParseFiles("./html/index.html")
		if err != nil {
			ErrorHandler(w, r, 500)
			return
		}
		t.Execute(w, database.FilterArr)
		return
	default:
		ErrorHandler(w, r, 500)
		return
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		ErrorHandler(w, r, 404)
		return
	}

	err := database.GetPersonalPageData()
	if err != nil {
		ErrorHandler(w, r, 500)
		return
	}

	if len(database.PersonalPageInformation) == 0 {
		ErrorHandler(w, r, 500)
		return
	}

	t, err := template.ParseFiles("./html/index.html")
	if err != nil {
		ErrorHandler(w, r, 500)
		return
	}

	switch r.Method {
	case "GET":
		t.Execute(w, database.PersonalPageInformation)
		return
	case "POST":
		searchInput := r.FormValue("searchInput")
		fmt.Println(searchInput)
		database.Search(searchInput)
		t.Execute(w, database.SearchArr)
		return
	default:
		ErrorHandler(w, r, 500)
		return
	}


}

func PersonalPage(w http.ResponseWriter, r *http.Request) {

	artistID, err := strconv.Atoi(r.FormValue("id"))

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	err = database.GetPersonalPageData()

	if len(database.PersonalPageInformation) == 0 {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}


	t, err := template.ParseFiles("./html/profile.html")

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, database.PersonalPageInformation[artistID-1])

	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError)
		return
	}

}

