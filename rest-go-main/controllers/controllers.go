package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"rest/api/database"
	"rest/api/jwtService"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Bank     int    `json:"bank"`
}

func Init() {
	var err error
	db, err = database.Init()
	if err != nil {
		fmt.Print("error initializing database")
		return
	}
	jwtService.SetSecret()
}

func HandleSignup(c *gin.Context) {

	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to parse request body. Wrong json"})
		return
	}

	var username string
	db.QueryRow("SELECT username FROM users WHERE username = $1", user.Username).Scan(&username)
	if len(username) != 0 {
		c.JSON(400, gin.H{"error": "User already exists"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hash)

	_, err = db.Exec("INSERT INTO users (username, password, bank) values ($1, $2, $3)", user.Username, user.Password, 0)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to insert into database"})
		fmt.Println(err)
		return
	}
	c.JSON(200, gin.H{
		"message": "User signed up successfully",
		"user":    user,
	})
}

func HandleLogin(c *gin.Context) {
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to parse request body. Wrong json"})
		return
	}
	var userBd User
	err = db.QueryRow("SELECT id, username, password, bank FROM users WHERE username = $1", user.Username).Scan(&userBd.Id, &userBd.Username, &userBd.Password, &userBd.Bank)

	if err == sql.ErrNoRows {
		c.JSON(400, gin.H{"error": "User does not exist"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(userBd.Password), []byte(user.Password))

	if err == nil {
		claims := jwt.MapClaims{
			"userid":   userBd.Id,
			"username": userBd.Username,
		}
		token, err := jwtService.GenerateJWT(claims)
		if err != nil {
			c.JSON(400, gin.H{"error": "problems with generating token"})
		} else {
			c.JSON(200, gin.H{"token": token, "id": user.Id})
		}
	} else {
		c.JSON(400, gin.H{"error": "incorrect password"})
	}
}

func parseClaims(claims jwt.MapClaims) User {
	user := User{}

	if userID, ok := claims["userid"].(float64); ok {
		user.Id = int(userID)
	}
	if username, ok := claims["username"].(string); ok {
		user.Username = username
	}
	var bank int
	row := db.QueryRow("SELECT bank from users where username=$1", user.Username)
	row.Scan(&bank)
	user.Bank = bank
	return user
}

func GetDocumentation(c *gin.Context) {
	htmlContent, err := os.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading documentation file")
		return
	}
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, string(htmlContent))
}

type Parking struct {
	Id        int    `json:"id"`
	Address   string `json:"address"`
	Places    int    `json:"places"`
	Available int    `json:"available"`
	Latitude  int    `json:"latitude,"`
	Longitude int    `json:"longitude"`
}

func GetParkings(c *gin.Context) {
	var parking Parking
	rows, err := db.Query("SELECT * FROM parkings")
	if err != nil {
		log.Fatal(err)
	}
	var arr []Parking
	for rows.Next() {
		rows.Scan(&parking.Id, &parking.Address, &parking.Places, &parking.Available, &parking.Latitude, &parking.Longitude)
		arr = append(arr, parking)
	}
	c.JSON(200, arr)
}

type Location struct {
	Latitude  int `json:"latitude"`
	Longitude int `json:"longitude"`
}

type Coord struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

func GetClosestParking(c *gin.Context) {
	var location Location
	var coord Coord
	c.BindJSON(&coord)
	location.Latitude, _ = strconv.Atoi(coord.Latitude)
	location.Longitude, _ = strconv.Atoi(coord.Longitude)

	fmt.Println("location is", location)
	parking := findClosestParking(location)
	c.JSON(200, parking)
}

func findClosestParking(location Location) Parking {
	rows, _ := db.Query("SELECT * FROM parkings")
	var parking Parking
	var minimum Parking
	var minVal int64 = math.MaxInt64
	for rows.Next() {
		rows.Scan(&parking.Id, &parking.Address, &parking.Places, &parking.Available, &parking.Latitude, &parking.Longitude)
		distance := math.Sqrt(math.Pow(float64(location.Latitude-parking.Latitude), 2) + math.Pow(float64(location.Longitude-parking.Longitude), 2))
		if int64(distance) < minVal {
			minimum = parking
			minVal = int64(distance)
		}
	}
	return minimum
}

type Spot struct {
	Id         int    `json:"id"`
	Parking_id int    `json:"parking_id"`
	Name       string `json:"name"`
	Available  bool   `json:"available"`
}

func GetParking(c *gin.Context) {
	var spot Spot
	var arr []Spot
	id := c.Param("id")
	rows, _ := db.Query("SELECT * FROM spots where parking_id=$1", id)

	for rows.Next() {
		rows.Scan(&spot.Id, &spot.Parking_id, &spot.Name, &spot.Available)
		arr = append(arr, spot)
	}

	c.JSON(200, arr)
}

func ReserveSpot(c *gin.Context) {
	spotId := c.Param("spotId")
	var parkingId int
	userId := c.Param("userId")
	duration := c.Param("duration")
	db.Exec("UPDATE spots set available='false' where id=$1", spotId)
	db.QueryRow("SELECT parking_id FROM spots where id=$1", spotId).Scan(&parkingId)
	db.Exec("UPDATE parkings set available=available-1 where id=$1", parkingId)
	db.Exec("UPDATE users set bank=bank-$1 where id=$2")
	db.Exec("INSERT INTO reserved (user_id, parking_id, spot_id, duration) values ($1, $2, $3)", userId, parkingId, spotId, duration)
}

type Dto struct {
	Name     string
	Parking  string
	Duration int
}

func CheckReserved(c *gin.Context) {
	userId := c.Param("id")
	rows, _ := db.Query("SELECT name, parking_id, duration FROM reserved where user_id=$1", userId)

	var name string
	var parking_id int
	var duration int
	var dto Dto
	var arr []Dto

	for rows.Next() {
		rows.Scan(&name, &parking_id, &duration)
		dto.Name = name
		dto.Duration = duration
		db.QueryRow("SELECT name FROM parkings where id=$1", parking_id).Scan(dto.Parking)
		arr = append(arr, dto)
	}

	c.JSON(200, arr)
}

func AddMoney(c *gin.Context) {
	id := c.Param("id")
	amount := c.Param("amount")
	db.Exec("UPDATE users set bank=bank+$1 where id=$2", amount, id)
	c.JSON(200, "")
}
