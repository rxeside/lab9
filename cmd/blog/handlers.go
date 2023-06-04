package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type indexPage struct {
	Title           string
	SubTitle        string
	FeaturedPosts   []postData
	MostRecentPosts []postData
}

// type featuredPostData struct {
// 	Title       string `db:"title"`
// 	Description string `db:"subtitle"`
// 	PostImg     string `db:"image_url"`
// 	Author      string `db:"author"`
// 	AuthorImg   string `db:"author_url"`
// 	PublishDate string `db:"publish_date"`
// 	PostID      string `db:"post_id"`
// }

// type mostRecentPostData struct {
// 	Title       string `db:"title"`
// 	Description string `db:"subtitle"`
// 	PostImg     string `db:"image_url"`
// 	Author      string `db:"author"`
// 	AuthorImg   string `db:"author_url"`
// 	PublishDate string `db:"publish_date"`
// 	PostID      string `db:"post_id"`
// }

type postData struct {
	PostID      string `db:"post_id"`
	Title       string `db:"title"`
	Description string `db:"subtitle"`
	ImgModifier string `db:"image_url"`
	Author      string `db:"author"`
	AuthorImg   string `db:"author_url"`
	PublishDate string `db:"publish_date"`
}

type createPostRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	AuthorName      string `json:"author"`
	AuthorPhoto     string `json:"avatar"`
	AuthorPhotoName string `json:"avatar_name"`
	Date            string `json:"date"`
	Image           string `json:"hero"`
	ImageName       string `json:"hero_name"`
	Content         string `json:"content"`
}

type postContent struct {
	Title    string `db:"title"`
	Subtitle string `db:"subtitle"`
	Image    string `db:"image_url"`
	Content  string `db:"content"`
}

func index(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := getPosts(db, 1)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		miniPosts, err := getPosts(db, 0)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		ts, err := template.ParseFiles("pages/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		data := indexPage{
			FeaturedPosts:   posts,
			MostRecentPosts: miniPosts,
		}

		err = ts.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		log.Println("Request completed successfully")
	}
}

func post(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		postIDStr := mux.Vars(r)["postID"]

		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "Invalid post id", 403)
			log.Println(err)
			return
		}

		post, err := postByID(db, postID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Post not found", 404)
				log.Println(err)
				return
			}

			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		ts, err := template.ParseFiles("pages/post.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		err = ts.Execute(w, post)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		log.Println("Request completed successfully")
	}
}

func login(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		ts, err := template.ParseFiles("pages/login.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		err = ts.Execute(w, nil)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}
	}
}

func admin(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		ts, err := template.ParseFiles("pages/admin.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		err = ts.Execute(w, post)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}
	}
}

func createPost(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		reqData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "1Error", 500)
			log.Println(err.Error())
			return
		}

		var req createPostRequest

		err = json.Unmarshal(reqData, &req)
		if err != nil {
			http.Error(w, "2Error", 500)
			log.Println(err.Error())
			return
		}

		authorImg, err := base64.StdEncoding.DecodeString(req.AuthorPhoto)
		if err != nil {
			http.Error(w, "img", 500)
			log.Println(err.Error())
			return
		}

		fmt.Println(req.AuthorPhotoName)

		fileAuthor, err := os.Create("static/image/" + req.AuthorPhotoName)
		if err != nil {
			fmt.Println("Unable to create file:", err)
			os.Exit(1)
		}
		defer fileAuthor.Close()
		_, err = fileAuthor.Write(authorImg)
		fmt.Println("Done.")

		image, err := base64.StdEncoding.DecodeString(req.Image)
		if err != nil {
			http.Error(w, "img", 500)
			log.Println(err.Error())
			return
		}

		fileImage, err := os.Create("static/image/" + req.ImageName)
		if err != nil {
			fmt.Println("Unable to create file:", err)
			os.Exit(1)
		}
		defer fileImage.Close()
		_, err = fileImage.Write(image)
		fmt.Println("Done.")
		req.Date = formatDate(req.Date)

		err = saveOrder(db, req)
		if err != nil {
			http.Error(w, "bd", 500)
			log.Println(err.Error())
			return
		}

		return
	}
}

func getPosts(db *sqlx.DB, feature int) ([]postData, error) {
	var query = ""
	if feature == 1 {
		query = `
		SELECT
			title,
			subtitle,
			image_url,
			author,
			author_url,
			publish_date,
			post_id
		FROM
			post
		WHERE featured = 1
	`
	} else if feature == 0 {
		query = `
		SELECT
			title,
			subtitle,
			image_url,
			author,
			author_url,
			publish_date,
			post_id
		FROM
			post
		WHERE featured = 0
	`
	}

	var posts []postData
	err := db.Select(&posts, query)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func postByID(db *sqlx.DB, postID int) (postContent, error) {
	const query = `
		SELECT
			title,
			subtitle,
			image_url,
			content
		FROM
		    post
		WHERE
			post_id = ?
	`

	var post postContent

	err := db.Get(&post, query, postID)
	if err != nil {
		return postContent{}, err
	}

	return post, nil
}

func saveOrder(db *sqlx.DB, req createPostRequest) error {
	const query = `
		INSERT INTO
			post
		(
			title,
			subtitle,
			author,
			author_url,
			publish_date,
			image_url,
			content,
			featured
		)
		VALUES
		(
			?,
			?,
			?,
			CONCAT('static/image/', ?),
			?,
			CONCAT('/static/image/', ?),
			?,
			?
		)
	`

	_, err := db.Exec(query, req.Title, req.Description, req.AuthorName, req.AuthorPhotoName, req.Date, req.ImageName, req.Content, 0)
	return err
}

func formatDate(oldDate string) string {
	dateStr := strings.Split(oldDate, "-")
	newDateStr := dateStr[2] + "/" + dateStr[1] + "/" + dateStr[0]
	return newDateStr
}
