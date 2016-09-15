package controllers

import (
	"io"
	"log"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/david-torres/duelyst-casual/models"
	"golang.org/x/net/websocket"
)

// Socket handles the websocket
func Socket(session *r.Session) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		// decay time of items
		var decay int32 = 86400

		// Initial socket connection, get changefeed
		games, _ := r.Table("games").
			// find games that are +/- decay old (unfortunately now() doesn't update automatically its a point in time from when this changefeed is created)
			Filter(r.Row.Field("timestamp").During(r.EpochTime(r.Now().ToEpochTime().Sub(decay)), r.EpochTime(r.Now().ToEpochTime().Add(decay)))).
			// find games that have not been accepted
			Filter(r.Row.Field("accepted").Eq(false)).
			// get changefeed
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(session)

		// write changes to socket
		go func() {
			for {
				var changefeed interface{}

				for games.Next(&changefeed) {
					err := websocket.JSON.Send(ws, changefeed)
					if err != nil {
						// no data to send
						log.Printf("socket write error changefeed %s", err)
						return
					}
				}
			}
		}()

		// loop forever receiving from socket
		for {
			// read data from socket
			game := new(models.Game)
			err := websocket.JSON.Receive(ws, &game)
			if err != nil {
				// no data received
				if err == io.EOF {
					log.Println("socket got EOF, terminating")
					return
				}

				log.Printf("socket receive error %s", err)
				continue
			}

			go func() {
				// got an id, update
				if game.ID != "" {
					game.Timestamp = time.Now()
					game.Accepted = true

					_, err := r.Table("games").Get(game.ID).Update(game).RunWrite(session)
					if err != nil {
						log.Printf("db write error %s", err)
						return
					}

					log.Printf("wrote to db %+v", game)
				}

				// got a new item, insert
				if game.Creator != "" {
					game.Timestamp = time.Now()

					_, err := r.Table("games").Insert(game).RunWrite(session)
					if err != nil {
						log.Printf("db write error %s", err)
						return
					}

					log.Printf("wrote to db %+v", game)
				}
			}()
		}
	})
}
