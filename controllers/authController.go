package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/mecirmartin/fiber_api/database"
	"github.com/mecirmartin/fiber_api/models"
	"golang.org/x/crypto/bcrypt"
)

const SECRET_KEY = "Secret"

type RegisterData struct {
	Username string `json:"username" xml:"username" form:"username"`
	LoginData
}

type LoginData struct {
	Email    string `json:"email" xml:"email" form:"email"`
	Password string `json:"password" xml:"password" form:"password"`
}

func Register(c *fiber.Ctx) error {
	data := new(RegisterData)

	err := c.BodyParser(data)

	if err != nil {
		fmt.Printf("Error while parsing data, error: %x", err)
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), 14)

	if err != nil {
		fmt.Printf("Error while encrypting data, error: %x", err)
		return err
	}

	user := models.User{
		Email:    data.Email,
		Username: data.Username,
		Password: string(hashedPassword),
	}

	database.DB.Create(&user)

	return c.JSON(user)
}

func Login(c *fiber.Ctx) error {
	loginData := new(LoginData)

	err := c.BodyParser(&loginData)

	if err != nil {
		fmt.Printf("Error while parsing data, error: %x", err)
		return err
	}

	var user models.User

	database.DB.Where("email = ?", loginData.Email).First(&user)

	if user.Id == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{"message": "User not found"})
	}

	// Compare PW from req.body to hashed value from DB
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{"message": "Password is incorrect"})
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
	})

	token, err := claims.SignedString([]byte(SECRET_KEY))

	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		fmt.Printf("Err %v", err)
		return c.JSON(fiber.Map{"message": "Unable to log in, try again later"})
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Minute * 15),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "Success",
	})
}

func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "User is not logged in",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	database.DB.Where("id = ?", claims.Issuer).First(&user)

	return c.JSON(user)
}

func Logout(c *fiber.Ctx) error {

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}

	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}
