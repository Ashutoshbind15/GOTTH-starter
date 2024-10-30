package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Ashutoshbind15/gotempltryout/data"
	"github.com/Ashutoshbind15/gotempltryout/views/layout"
	"github.com/Ashutoshbind15/gotempltryout/views/user"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
)

type UserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var tokenAuth *jwtauth.JWTAuth

// db -> user funcs

func createUser(uname, pwd string) (string, error) {
	result, err := data.DBClient.Exec(`INSERT INTO USERS(username, password, role) VALUES ($1, $2, $3)`, uname, pwd, "user")
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return "", fmt.Errorf("no rows were inserted")
	}

	return "User created successfully", nil
}

func UserExists(uname string) (bool, error) {
	var exists bool

	err := data.DBClient.QueryRow(`SELECT EXISTS(SELECT 1 FROM USERS WHERE username = $1)`, uname).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error checking if user exists: %w", err)
	}

	return exists, nil
}

// user handlers

func userHandler(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	usertempl := user.Show((claims["uid"]).(string))
	page := layout.Layout(usertempl, "user page")

	page.Render(context.Background(), w)
}

func getSignupHandler(w http.ResponseWriter, r *http.Request) {
	usf := user.SignupForm()
	page := layout.Layout(usf, "signup")

	page.Render(context.Background(), w)
}


func postSignupHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	uexists, err := UserExists(r.FormValue("username"))
	if err != nil {
		panic(err)
	}

	if uexists {
		uerr := user.UserExistsError()
		uerr.Render(context.Background(), w)
		return
	}

	_ , err = createUser(r.FormValue("username"), r.FormValue("password"))

	if err != nil {
		panic(err)
	}
	
	m := make(map[string]interface{})
	m["uid"] = r.FormValue("username")
	m["role"] = "user"

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	_, tokenstr, _ := tokenAuth.Encode(m)

	cook := http.Cookie{
		Name: "jwt",
		Value: tokenstr,
	}

	http.SetCookie(w, &cook)
	w.Header().Set("HX-Redirect", "/users")
	w.WriteHeader(http.StatusAccepted)
}

func main() {
	fmt.Println("Init")
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	data.InitDB()
	// data.InitTables()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Get("/signup", getSignupHandler)
	r.Post("/signup", postSignupHandler)
	
		// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Get("/users", userHandler)
	})

	fs := http.FileServer(http.Dir("public"))
	r.Handle("/public/*", http.StripPrefix("/public/", fs))

	http.ListenAndServe(":3000", r)
}
