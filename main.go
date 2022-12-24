package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/django"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

type User struct {
	gorm.Model
	Name         string
	Email        *string `gorm:"unique"`
	Age          uint8
	MemberNumber sql.NullString
	ActivatedAt  sql.NullTime
}

var DB *gorm.DB

func ConnectDb() {
	db, err := gorm.Open(sqlite.Open("dev.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	log.Println("connected")
	db.Logger = logger.Default.LogMode(logger.Info)
	log.Println("running migrations")
	db.AutoMigrate(&Product{})

	DB = db
}

func Home(c *fiber.Ctx) error {
	return c.Render("home", fiber.Map{
		"Phrase": "Hello Fiber",
	})
}

func ProductList(c *fiber.Ctx) error {
	products := []Product{}
	DB.Find(&products)
	return c.Render("product-list", fiber.Map{
		"Products": &products,
	})
}

func ProductDetail(c *fiber.Ctx) error {
	product := Product{}
	id := c.Params("id")
	DB.Where("id = ?", id).Find(&product)
	return c.Render("product-detail", fiber.Map{
		"Product": &product,
	})
}

func setUpRoutes(app *fiber.App) {
	app.Get("/", Home)
	app.Get("/products", ProductList)
	app.Get("/products/:id", ProductDetail)
}

//go:embed views/*
var viewsfs embed.FS

func getViewsFileSystem() http.FileSystem {
	fsys, err := fs.Sub(viewsfs, "views")
	if err != nil {
		log.Fatal(err)
	}
	return http.FS(fsys)
}

func main() {
	ConnectDb()

	engine := django.NewFileSystem(getViewsFileSystem(), ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	setUpRoutes(app)

	app.Use(cors.New())

	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404) // Not found
	})

	log.Fatal(app.Listen(":3000"))
}
