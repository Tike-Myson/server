package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Tike-Myson/database"
	uuid "github.com/satori/go.uuid"
)

var Cookies = make(map[string]string)

type Error struct {
	ErrorStatus int
}

var ErrorResponse Error

type NewPostStruct struct {
	Username       string
	Categories     []string
	ImageSizeError int
}

var NewPostResponse NewPostStruct

//NewPost ...
func NewPost(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/newPost" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	t, err := template.ParseFiles("./html/createPost.html")
	if err != nil {
		ErrorHandler(w, r, 500)
		return
	}
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		NewPostResponse.Username = username
		NewPostResponse.Categories = database.GetCategory()
		t.Execute(w, NewPostResponse)
		return
	case "POST":
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			t.Execute(w, nil)
			return
		}
		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			t.Execute(w, nil)
			return
		}
		UserID := database.GetUser(username)
		Title := ""
		Content := ""
		InputTagsSplit := ""
		var InputTags []string
		imgURL := ImageUpload(r)
		if ContentValidation(r.FormValue("TitleInput")) {
			Title = r.FormValue("TitleInput")
		}
		if ContentValidation(r.FormValue("ContentArea")) {
			Content = r.FormValue("ContentArea")
		}
		if ContentValidation(r.FormValue("TagInput")) {
			InputTagsSplit = r.FormValue("TagInput")
			temp := strings.Split(InputTagsSplit, " ")
			InputTags = DeleteDuplicateTags(temp)
		}
		PostID := uuid.NewV4()
		if Title == "" || len(InputTags) == 0 || InputTagsSplit == "" || (Content == "" && imgURL == "") {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if imgURL == "error" {
			NewPostResponse.ImageSizeError = 1
			t.Execute(w, NewPostResponse)
			return
		}
		currentTime := time.Now().Format("2006.01.02 15:04:05")
		ok = database.CreatePost(PostID.String(), UserID, username, Title, Content, currentTime, imgURL)

		for _, v := range InputTags {
			CategoryPostLinkID := uuid.NewV4()
			database.CreateCategoryPostLink(CategoryPostLinkID, PostID, v)
		}

		if ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func DeleteDuplicateTags(arr []string) []string {
	if len(arr) == 1 || len(arr) == 0 {
		return arr
	}
	var correct []string
	for i := range arr {
		count := 0
		for j := range arr {
			if i != j && arr[i] == arr[j] {
				count++
			}
		}
		if count > 0 {
			continue
		}
		correct = append(correct, arr[i])
	}
	if len(correct) == 0 {
		correct = append(correct, arr[0])
	}
	return correct
}

func ContentValidation(str string) bool {
	r := []rune(str)
	count := 0
	for _, v := range r {
		if v == 32 {
			count++
		}
		if v == 9 {
			count++
		}
	}
	if count == len(str) {
		return false
	}
	return true
}

type PostResp struct {
	ResponseCode int
	Username     string
	Posts        []database.Post
	Comments     []database.Comment
}

var Response PostResp

//SignInStruct ...
type SignInStruct struct {
	Username     string
	ResponseCode int
}

type SignUpStruct struct {
	ResponseCode int
}

//SignUp ...
func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	var ResponseSignUp SignUpStruct
	t, err := template.ParseFiles("./html/signup.html")
	if err != nil {
		ErrorHandler(w, r, 500)
		return
	}
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			t.Execute(w, nil)
			return
		}

		_, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			t.Execute(w, nil)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	case "POST":
		pass := r.FormValue("inputPassword")
		UserEmail := r.FormValue("inputEmail")
		UserLogin := r.FormValue("inputLogin")
		if pass == "" || UserEmail == "" || UserLogin == "" {
			ResponseSignUp.ResponseCode = 2
			t.Execute(w, ResponseSignUp)
			return
		}
		u1 := uuid.NewV4()
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		UserID := u1
		if strings.ContainsRune(pass, 9) || strings.ContainsRune(UserEmail, 9) || strings.ContainsRune(UserLogin, 9) || pass == UserLogin || pass == UserEmail || UserLogin == UserEmail {
			ResponseSignUp.ResponseCode = 2
			t.Execute(w, ResponseSignUp)
			return
		}
		if strings.ContainsRune(pass, 32) || strings.ContainsRune(UserEmail, 32) || strings.ContainsRune(UserLogin, 32) || pass == UserLogin || pass == UserEmail || UserLogin == UserEmail {
			ResponseSignUp.ResponseCode = 2
			t.Execute(w, ResponseSignUp)
			return
		}
		UserPassword, err := database.HashPassword(pass)
		if err != nil {
			fmt.Printf("Something went wrong: %s", err)
			return
		}
		ok := database.CreateUser(UserID, UserLogin, UserPassword, UserEmail)
		if ok {
			ResponseSignUp.ResponseCode = 0
			cookie := MakeCookie(UserLogin, Cookies)
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		ResponseSignUp.ResponseCode = 3
		t.Execute(w, ResponseSignUp)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

//SignIn ...
func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signin" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	t, err := template.ParseFiles("./html/signin.html")
	if err != nil {
		ErrorHandler(w, r, 500)
		return
	}
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			t.Execute(w, nil)
			return
		}

		_, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			t.Execute(w, nil)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	case "POST":
		pass := r.FormValue("inputPassword")
		UserLogin := r.FormValue("inputEmail")
		var ResponseSignIn SignInStruct
		ValidationUser, err := database.IsValidateUser(UserLogin, pass)
		if err != nil {
			ResponseSignIn.ResponseCode = 1
			t.Execute(w, ResponseSignIn)
			return
		}
		if ValidationUser.UserID == uuid.Nil {
			ResponseSignIn.ResponseCode = 1
			t.Execute(w, ResponseSignIn)
			return
		}
		ResponseSignIn.Username = UserLogin

		ok, token := IsSessionExist(ValidationUser.Login)
		if ok {
			cookie := DeleteCookie("session", token, "localhost", "")
			http.SetCookie(w, &cookie)
		}

		cookie := MakeCookie(ValidationUser.Login, Cookies)
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
	}

}

func Post(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post" {
		ErrorHandler(w, r, 500)
		return
	}
	t, err := template.ParseFiles("./html/post.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	Response.Posts = nil
	switch r.Method {

	case "GET":
		PostID := r.FormValue("PostID")
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			Response.Posts = database.GetPostByID("", PostID)
			Response.Comments = database.GetAllComments("", PostID)
			Response.Username = ""
			Response.ResponseCode = 401
			t.Execute(w, Response)
			return
		}

		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			Response.Posts = database.GetPostByID("", PostID)
			Response.Comments = database.GetAllComments("", PostID)
			Response.Username = ""
			Response.ResponseCode = 401
			t.Execute(w, Response)
			return
		}
		UserID := database.GetUser(username)
		Response.Posts = database.GetPostByID(UserID.String(), PostID)
		Response.Comments = database.GetAllComments(UserID.String(), PostID)
		Response.Username = username
		Response.ResponseCode = 200
		t.Execute(w, Response)
		return
	case "POST":
		t.Execute(w, nil)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func ImageUpload(r *http.Request) string {
	imgURL := ""
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Println(err)
		return imgURL
	}
	file, handler, err := r.FormFile("ImageInput")
	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err.Error())
		return imgURL
	}
	defer file.Close()
	if handler.Size > 20878139 {
		imgURL = "error"
		return imgURL
	}

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("content", "upload-*.png")

	if err != nil {
		log.Println(err.Error())
		return imgURL
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err.Error())
		return imgURL
	}
	// write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		log.Println(err.Error())
		return imgURL
	}
	imgURL = "/" + tempFile.Name()
	return imgURL
}

func FilterByCategories(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/category" {
		ErrorHandler(w, r, 500)
		return
	}

	t, err := template.ParseFiles("./html/index.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	switch r.Method {
	case "GET":
		categoryName := r.FormValue("CategoryName")
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			Resp.Posts = database.GetPostsByCategory(uuid.Nil.String(), categoryName)
			Resp.Username = ""
			Resp.ResponseCode = 401
			t.Execute(w, Resp)
			return
		}

		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			Resp.Posts = database.GetPostsByCategory(uuid.Nil.String(), categoryName)
			Resp.Username = ""
			Resp.ResponseCode = 401
			t.Execute(w, Resp)
			return
		}
		userUUID := database.GetUser(username)
		Resp.Posts = database.GetPostsByCategory(userUUID.String(), categoryName)
		Resp.Username = username
		Resp.ResponseCode = 200
		t.Execute(w, Resp)
		return
	case "POST":
		t.Execute(w, nil)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func MyPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myPosts" {
		ErrorHandler(w, r, 500)
		return
	}

	t, err := template.ParseFiles("./html/index.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			Resp.Posts = database.GetPostsByAuthorID(uuid.Nil.String())
			Resp.Username = ""
			Resp.ResponseCode = 401
			t.Execute(w, Resp)
			return
		}

		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			Resp.Posts = database.GetPostsByAuthorID(uuid.Nil.String())
			Resp.Username = ""
			Resp.ResponseCode = 401
			t.Execute(w, Resp)
			return
		}
		userUUID := database.GetUser(username)
		Resp.Posts = database.GetPostsByAuthorID(userUUID.String())
		Resp.Username = username
		Resp.ResponseCode = 200
		t.Execute(w, Resp)
		return
	case "POST":
		t.Execute(w, nil)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

type IndexResponse struct {
	Username     string
	ResponseCode int
	Posts        []database.Post
}

var Resp IndexResponse

func LikeComment(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/likecomment" {
		ErrorHandler(w, r, 404)
		return
	}
	cookieUser, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	username, _ := IsTokenExist(cookieUser.Value)
	UserID := database.GetUser(username)
	PostID := r.FormValue("PostID")
	PostFlag := r.FormValue("PostFlag")
	CommentID := r.FormValue("CommentID")
	if CommentID == "" {
		CommentID = uuid.Nil.String()
	}
	u1 := uuid.NewV4()
	if UserID == uuid.Nil {
		http.Redirect(w, r, "/signin", 303)
		return
	}
	if PostFlag == "post" {
		database.CreateCommentLike(u1.String(), UserID.String(), PostID, CommentID)
		http.Redirect(w, r, "/post?PostID="+PostID, 303)
		return
	}
	database.CreateCommentLike(u1.String(), UserID.String(), PostID, CommentID)
	http.Redirect(w, r, "/", 303)
	return
}

//LogOut ...
func LogOut(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	for _, cookie := range r.Cookies() {
		DeletingCookie := DeleteCookie(cookie.Name, cookie.Value, cookie.Domain, cookie.Path)
		http.SetCookie(w, &DeletingCookie)
		http.Redirect(w, r, "/", 303)
		return
	}
}

func Like(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/like" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	cookieUser, err := r.Cookie("session")
	if err != nil {
		ErrorHandler(w, r, 401)
		return
	}
	username, _ := IsTokenExist(cookieUser.Value)
	UserID := database.GetUser(username)
	PostID := r.FormValue("PostID")
	PostFlag := r.FormValue("PostFlag")
	CommentID := r.FormValue("CommentID")
	if CommentID == "" {
		CommentID = uuid.Nil.String()
	}
	u1 := uuid.NewV4()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	if UserID == uuid.Nil {
		http.Redirect(w, r, "/signin", 303)
		return
	}
	if PostFlag == "post" {
		database.CreateLike(u1.String(), UserID.String(), PostID)
		http.Redirect(w, r, "/post?PostID="+PostID, 303)
		return
	}
	database.CreateLike(u1.String(), UserID.String(), PostID)
	http.Redirect(w, r, "/", 303)
	return
}

//Index ...
func Index(w http.ResponseWriter, r *http.Request) {
	database.CreateAllTables()
	if r.URL.Path != "/" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles("./html/index.html")
	if err != nil {
		log.Println(err.Error())
		ErrorHandler(w, r, 500)
		return
	}
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			Resp.Posts = database.GetAllPosts(uuid.Nil, "")
			Resp.Username = ""
			Resp.ResponseCode = 401
			t.Execute(w, Resp)
			return
		}

		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			Resp.Posts = database.GetAllPosts(uuid.Nil, "")
			Resp.Username = ""
			Resp.ResponseCode = 401
			t.Execute(w, Resp)
			return
		}
		userUUID := database.GetUser(username)
		Resp.Posts = database.GetAllPosts(userUUID, username)
		Resp.Username = username
		Resp.ResponseCode = 200
		t.Execute(w, Resp)
		return
	case "POST":
		t.Execute(w, nil)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func LikedPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/likedPosts" {
		ErrorHandler(w, r, 500)
		return
	}

	t, err := template.ParseFiles("./html/index.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/", 303)
			return
		}

		username, ok := IsTokenExist(cookie.Value)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/", 303)
			return
		}
		userUUID := database.GetUser(username)
		PostIDArr := database.GetPostID(userUUID.String())
		Resp.Posts = database.GetLikedPostsByUserID(userUUID.String(), PostIDArr)
		Resp.Username = username
		Resp.ResponseCode = 200
		t.Execute(w, Resp)
		return
	case "POST":
		t.Execute(w, nil)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
	}

}

//ErrorHandler ...
func ErrorHandler(w http.ResponseWriter, r *http.Request, status int) {

	w.WriteHeader(status)

	t, err := template.ParseFiles("./html/error.html")
	if err != nil {
		log.Println(err.Error())
	}

	ErrorResponse.ErrorStatus = status

	t.Execute(w, ErrorResponse)

}

func DislikeComment(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/dislikecomment" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	cookieUser, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username, _ := IsTokenExist(cookieUser.Value)
	UserID := database.GetUser(username)
	PostID := r.FormValue("PostID")
	PostFlag := r.FormValue("PostFlag")
	CommentID := r.FormValue("CommentID")
	if CommentID == "" {
		CommentID = uuid.Nil.String()
	}
	u1 := uuid.NewV4()
	if UserID == uuid.Nil {
		http.Redirect(w, r, "/signin", 303)
		return
	}
	if PostFlag == "post" {
		database.CreateCommentDislike(u1.String(), UserID.String(), PostID, CommentID)
		http.Redirect(w, r, "/post?PostID="+PostID, 303)
		return
	}
	database.CreateCommentDislike(u1.String(), UserID.String(), PostID, CommentID)
	http.Redirect(w, r, "/", 303)
	return
}

func Dislike(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/dislike" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	cookieUser, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username, _ := IsTokenExist(cookieUser.Value)
	UserID := database.GetUser(username)
	PostID := r.FormValue("PostID")
	PostFlag := r.FormValue("PostFlag")
	CommentID := r.FormValue("CommentID")
	if CommentID == "" {
		CommentID = uuid.Nil.String()
	}
	u1 := uuid.NewV4()
	if UserID == uuid.Nil {
		http.Redirect(w, r, "/signin", 303)
		return
	}
	if PostFlag == "post" {
		database.CreateDislike(u1.String(), UserID.String(), PostID)
		http.Redirect(w, r, "/post?PostID="+PostID, 303)
		return
	}
	database.CreateDislike(u1.String(), UserID.String(), PostID)
	http.Redirect(w, r, "/", 303)
	return
}

//MakeCookie ...
func MakeCookie(login string, Cookies map[string]string) http.Cookie {

	u1 := uuid.NewV4()
	Cookies[u1.String()] = login

	expiration := time.Now().Add(1 * time.Hour)
	cookie := http.Cookie{Name: "session", Value: u1.String(), Expires: expiration}
	return cookie

}

//DeleteCookie ...
func DeleteCookie(name, token, domain, path string) http.Cookie {

	delete(Cookies, token)
	cookie := http.Cookie{
		Name:     name,
		Value:    "",
		Domain:   domain,
		Path:     path,
		MaxAge:   0,
		HttpOnly: true,
	}
	return cookie
}

func IsSessionExist(username string) (bool, string) {
	for i, v := range Cookies {
		if username == v {
			return true, i
		}
	}
	return false, ""
}

//IsTokenExist ...
func IsTokenExist(token string) (string, bool) {

	i, found := Cookies[token]
	if found {
		return i, true
	}
	return "", false

}

func Comment(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/comment" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}
	Content := ""
	PostID := r.FormValue("PostID")
	if ContentValidation(r.FormValue("CommentContent")) {
		Content = r.FormValue("CommentContent")
	}
	if Content == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	cookieUser, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	username, ok := IsTokenExist(cookieUser.Value)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u1 := uuid.NewV4()
	UserID := database.GetUser(username)
	currentTime := time.Now().Format("2006.01.02 15:04:05")
	ok = database.CreateComment(u1.String(), UserID.String(), PostID, username, Content, currentTime)
	if ok {
		http.Redirect(w, r, "/post?PostID="+PostID, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", 303)
	return
}
