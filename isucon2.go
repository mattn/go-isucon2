package main

import (
	"bufio"
	_ "code.google.com/p/go-mysql-driver/mysql"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hoisie/web"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var db *sql.DB
var tp *template.Template

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	} `json:"database"`
}

type Artist struct {
	Id   string
	Name string
}

type Ticket struct {
	Id         string
	Name       string
	Count      int
	ArtistName string
}

type Variation struct {
	TicketId string
	Id       string
	Name     string
	SeatIds  [][]string
	Vacancy  int
}

type RecentSold struct {
	Id    string
	AName string
	TName string
	VName string
}

func recentSolds() []RecentSold {
	rows, err := db.Query(`
		SELECT stock.seat_id, variation.name AS v_name, ticket.name AS t_name, artist.name AS a_name FROM stock
		  JOIN variation ON stock.variation_id = variation.id
		  JOIN ticket ON variation.ticket_id = ticket.id
		  JOIN artist ON ticket.artist_id = artist.id
		WHERE order_id IS NOT NULL
		ORDER BY order_id DESC LIMIT 10
	`)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	var rs []RecentSold
	for rows.Next() {
		var id, aname, tname, vname string
		rows.Scan(&id, &aname, &tname, &vname)
		rs = append(rs, RecentSold{id, aname, tname, vname})
	}
	return rs
}

func artists() []Artist {
	rows, err := db.Query(`
        SELECT * FROM artist ORDER BY id
	`)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	var a []Artist
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		a = append(a, Artist{id, name})
	}
	return a
}

func artist(artistId string) *Artist {
	rows, err := db.Query(`
        SELECT id, name FROM artist WHERE id = ? LIMIT 1
	`, artistId)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	if rows.Next() == false {
		panic(err.Error())
	}
	var id, name string
	rows.Scan(&id, &name)
	return &Artist{id, name}
}

func tickets(ticketId string) []Ticket {
	rows, err := db.Query(`
        SELECT id, name FROM ticket WHERE artist_id = ? ORDER BY id
	`, ticketId)
	if err != nil {
		panic(err.Error())
	}

	var t []Ticket
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		t = append(t, Ticket{id, name, 0, ""})
	}
	rows.Close()

	for i := range t {
		rows, err := db.Query(`
            SELECT COUNT(*) FROM variation
            INNER JOIN stock ON stock.variation_id = variation.id
            WHERE variation.ticket_id = ? AND stock.order_id IS NULL
		`, t[i].Id)
		if err != nil {
			panic(err.Error())
		}
		if rows.Next() {
			rows.Scan(&t[i].Count)
		}
		rows.Close()
	}

	return t
}

func ticket(ticketId string) (*Ticket, []Variation) {
	rows, err := db.Query(`
        SELECT t.id, t.name, a.name AS artist_name FROM ticket t INNER JOIN artist a ON t.artist_id = a.id WHERE t.id = ? LIMIT 1
	`, ticketId)
	if err != nil {
		panic(err.Error())
	}

	if rows.Next() == false {
		panic(err.Error())
	}
	var id, name, aname string
	rows.Scan(&id, &name, &aname)
	t := &Ticket{id, name, 0, aname}

	rows.Close()

	rows, err = db.Query(`
        SELECT id, name FROM variation WHERE ticket_id = ? ORDER BY id
	`, ticketId)
	if err != nil {
		panic(err.Error())
	}

	var v []Variation
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		v = append(v, Variation{t.Id, id, name, nil, 0})
	}
	rows.Close()

	for i := range v {
		rows, err := db.Query(`
            SELECT seat_id, order_id FROM stock WHERE variation_id = ?
		`, v[i].Id)
		if err != nil {
			panic(err.Error())
		}
		var seatIds [][]string
		for y := 0; y < 64; y++ {
			seatIds = append(seatIds, []string{})
			for x := 0; x < 64; x++ {
				seatIds[y] = append(seatIds[y], "")
			}
		}

		for rows.Next() {
			var seatId, orderId string
			var x, y int
			rows.Scan(&seatId, &orderId)
			fmt.Sscanf(seatId, "%02d-%02d", &y, &x)
			seatIds[y][x] = orderId
		}
		v[i].SeatIds = seatIds
		rows.Close()

		rows, err = db.Query(`
            SELECT COUNT(*) FROM stock WHERE variation_id = ? AND order_id IS NULL
		`, v[i].Id)
		if err != nil {
			panic(err.Error())
		}
		if rows.Next() {
			rows.Scan(&v[i].Vacancy)
		}
		rows.Close()
	}

	return t, v
}

func main() {
	rootDir := filepath.Dir(os.Args[0])
	env := os.Getenv("ISUCON_ENV")
	if env == "" {
		env = "local"
	}

	configFile := filepath.Clean(rootDir +
		fmt.Sprintf("/../config/common.%s.json", env))
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	var config Config
	json.Unmarshal(f, &config)

	templateFile := filepath.Join(rootDir, "templates", "*.t")
	tp, err := template.ParseGlob(templateFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	db, err = sql.Open("mysql", fmt.Sprintf(
		"%v:%v@tcp(%v:%v)/%s?charset=utf8",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName))
	if err != nil {
		log.Fatal(err.Error())
	}

	web.Get("/", func(ctx *web.Context) {
		if err := tp.ExecuteTemplate(ctx, "index", &struct {
			RecentSolds []RecentSold
			Artists     []Artist
		}{recentSolds(), artists()}); err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
	})

	web.Get("/artist/(.*)", func(ctx *web.Context, id string) {
		a := artist(id)
		if err := tp.ExecuteTemplate(ctx, "artist", &struct {
			RecentSolds []RecentSold
			Name        string
			Tickets     []Ticket
		}{recentSolds(), a.Name, tickets(id)}); err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
	})

	web.Get("/ticket/(.*)", func(ctx *web.Context, id string) {
		t, v := ticket(id)
		if err := tp.ExecuteTemplate(ctx, "ticket", &struct {
			RecentSolds []RecentSold
			Ticket      Ticket
			Variations  []Variation
		}{recentSolds(), *t, v}); err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
	})

	web.Post("/buy", func(ctx *web.Context) {
		variationId, ok := ctx.Params["variation_id"]
		if !ok {
			variationId = ""
		}
		memberId, ok := ctx.Params["memberId"]
		if !ok {
			memberId = ""
		}

		tx, err := db.Begin()
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}

		res, err := tx.Exec(`INSERT INTO order_request (member_id) VALUES (?)`, memberId)
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
		orderId, err := res.LastInsertId()
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
		res, err = tx.Exec(`UPDATE stock SET order_id = ? WHERE variation_id = ? AND order_id IS NULL ORDER BY RAND() LIMIT 1`, orderId, variationId)
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
		affected, err := res.RowsAffected()
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}

		if affected > 0 {
			rows, err := tx.Query(`SELECT seat_id FROM stock WHERE order_id = ? LIMIT 1`, orderId)
			if err != nil {
				tx.Rollback()
				log.Print(err.Error())
				ctx.Abort(500, "Server Error")
			}
			var seatId string
			if rows.Next() {
				rows.Scan(&seatId)
			}
			rows.Close()
			tx.Commit()
			err = tp.ExecuteTemplate(ctx, "complete", &struct {
				SeatId   string
				MemberId string
			}{
				seatId,
				memberId,
			})
		} else {
			tx.Rollback()
			err = tp.ExecuteTemplate(ctx, "soldout", nil)
		}
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
	})

	web.Get("/admin", func(ctx *web.Context) {
		err := tp.ExecuteTemplate(ctx, "admin", nil)
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
	})

	web.Post("/admin", func(ctx *web.Context) {
		sqlFile := filepath.Clean(filepath.Join(rootDir, "../config/database/initial_data.sql"))
		f, err := os.Open(sqlFile)
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
		defer f.Close()
		buf := bufio.NewReader(f)
		for {
			b, err := buf.ReadString('\n')
			if err != nil {
				break
			} else if line := strings.TrimSpace(string(b)); line != "" {
				_, err = db.Exec(line)
				if err != nil {
					log.Print(err.Error())
					ctx.Abort(500, "Server Error")
				}
			}
		}
		ctx.Redirect(302, "/admin")
	})

	web.Get("/admin/order.csv", func(ctx *web.Context) {
		rows, err := db.Query(`
            SELECT order_request.*, stock.seat_id, stock.variation_id, stock.updated_at
            FROM order_request JOIN stock ON order_request.id = stock.order_id
            ORDER BY order_request.id ASC
		`)
		if err != nil {
			log.Print(err.Error())
			ctx.Abort(500, "Server Error")
		}
		defer rows.Close()

		for rows.Next() {
			var id, memberId, seatId, variationId, updatedAt string
			rows.Scan(&id, &memberId, &seatId, &variationId, &updatedAt)
			ctx.SetHeader("Content-Type", "text/csv; charset=utf-8", true)
			line := strings.Join([]string{id, memberId, seatId, variationId, updatedAt}, ",") + "\n"
			ctx.Write([]byte(line))
		}
	})

	web.Config.StaticDir = filepath.Clean(rootDir + "/../staticfiles")

	web.Run(":8081")
}
