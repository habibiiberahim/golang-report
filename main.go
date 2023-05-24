package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"github.com/nfnt/resize"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"image"
	"image/jpeg"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	dsn := "randal:dbrandaltapin2021@tcp(36.95.205.37:3306)/absen?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database")
	}

	app := fiber.New()

	app.Get("/events", func(c *fiber.Ctx) error {
		rows, err := db.Table("absens").Select("acara").Group("acara").Rows()
		if err != nil {
			fmt.Println(err.Error())
		}

		var events []string
		defer rows.Close()
		for rows.Next() {
			var event string
			rows.Scan(&event)

			events = append(events, event)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "get list of events",
			"data":    events,
		})
	})

	app.Get("/events/:event", func(c *fiber.Ctx) error {
		param := c.Params("event")
		event, _ := url.PathUnescape(param)
		rows, _ := db.Table("absens").Select("name, instansi, jabatan, filename").Where("acara", event).Rows()

		var participants [][]string
		no := 1
		defer rows.Close()
		for rows.Next() {
			var name string
			var agency string
			var position string
			var filename string
			rows.Scan(&name, &agency, &position, &filename)

			participant := []string{
				strconv.Itoa(no),
				strings.Title(name),
				strings.Title(agency),
				strings.Title(position),
				filename,
			}
			participants = append(participants, participant)
			no++
		}

		m := pdf.NewMaroto(consts.Portrait, consts.Letter)
		m.SetPageMargins(5, 5, 5)
		m.SetBorder(true)

		m.RegisterHeader(func() {
			m.Row(20, func() {
				m.Col(12, func() {
					m.Text(event, props.Text{
						Top:             3,
						VerticalPadding: 10.0,
						Size:            16,
						Align:           consts.Center,
					})
				})
			})
			m.Row(20, func() {
				m.Col(1, func() {
					m.Text("No", props.Text{
						Size:            10.0,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(3, func() {
					m.Text("Nama", props.Text{
						Size:            10.0,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(3, func() {
					m.Text("Instansi", props.Text{
						Size:            10.0,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(3, func() {
					m.Text("Jabatan", props.Text{
						Size:            10.0,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(2, func() {
					m.Text("TTD", props.Text{
						Size:            10.0,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
			})
		})

		for _, participant := range participants {
			m.Row(20, func() {
				m.Col(1, func() {
					m.Text(participant[0], props.Text{
						Size:            10,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(3, func() {
					m.Text(participant[1], props.Text{
						Size:            10,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(3, func() {
					m.Text(participant[2], props.Text{
						Size:            10,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(3, func() {
					m.Text(participant[3], props.Text{
						Size:            10,
						Family:          consts.Arial,
						Align:           consts.Center,
						Top:             1.0,
						Extrapolate:     false,
						VerticalPadding: 1.0,
					})
				})
				m.Col(2, func() {

					file, err := os.Open("Kusmadi.png")
					if err != nil {
						log.Fatalf("Failed to open image file: %v", err)
					}
					defer file.Close()

					// Decode the PNG image
					img, _, err := image.Decode(file)
					if err != nil {
						log.Fatalf("Failed to decode image: %v", err)
					}

					// Convert the image to JPEG format
					convertedImg := resize.Resize(200, 0, img, resize.Lanczos3)
					outputFile, err := os.Create("signature/Kusmadi.jpg")
					if err != nil {
						log.Fatalf("Failed to create output file: %v", err)
					}
					defer outputFile.Close()
					jpeg.Encode(outputFile, convertedImg, nil)

					err = m.FileImage("signature/Kusmadi.jpg", props.Rect{
						Left:    5,
						Top:     5,
						Center:  true,
						Percent: 80,
					})

					if err != nil {
						c.SendString(err.Error())
					}
				})
			})
		}
		pdfName := fmt.Sprintf("%s .pdf ", event)
		path := fmt.Sprintf("PDF/%s", pdfName)

		err := m.OutputFileAndClose(path)
		if err != nil {
			fmt.Println("Could not save PDF:", err)
			os.Exit(1)
		}

		c.SendFile(path)
		return nil
	})

	log.Fatal(app.Listen(":3000"))

}
