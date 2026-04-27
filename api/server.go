package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"eng_club/api/handlers"
	"eng_club/config"
	"eng_club/database"
	"eng_club/telegram"
)

// Server exposes REST routes backed by MySQL and the club MySQL manager.
type Server struct {
	cfg    *config.Config
	h      *handlers.Deps
	router *router.Router
}

// NewServer wires routes. cfg must be validated; db and mgr must share the same pool (use club.NewMySQLManagerFromDB).
func NewServer(cfg *config.Config, db *database.MySQLDatabaseManager, telegramBot *telegram.Notifier) *Server {
	s := &Server{
		cfg:    cfg,
		h:      &handlers.Deps{DB: db, TelegramBot: telegramBot},
		router: router.New(),
	}
	s.routes()
	return s
}

// Handler returns the fasthttp handler (with CORS wrapper).
func (s *Server) Handler() fasthttp.RequestHandler {
	return cors(s.router.Handler)
}

func (s *Server) routes() {
	h := s.h

	// Members
	s.router.GET("/api/members", h.ListMembers)
	s.router.GET("/api/pairs/without", h.ListWithoutPairs)
	s.router.POST("/api/pairs/without/add", h.AddWithoutPairs)

	// Metro stations list
	s.router.GET("/api/metro-list", h.GetMetroList)

	// Places
	s.router.GET("/api/places", h.ListPlaces)
	s.router.POST("/api/places", h.CreatePlace)
	s.router.GET("/api/places/{id}", h.GetPlace)
	s.router.PUT("/api/places/{id}", h.UpdatePlace)
	s.router.DELETE("/api/places/{id}", h.DeletePlace)

	// Events
	s.router.GET("/api/events", h.ListEvents)
	s.router.POST("/api/events", h.CreateEvent)
	s.router.GET("/api/events/{id}", h.GetEvent)
	s.router.PUT("/api/events/{id}", h.UpdateEvent)
	s.router.DELETE("/api/events/{id}", h.DeleteEvent)
	s.router.POST("/api/events/{id}/start", h.StartEvent)
	s.router.POST("/api/events/{id}/add-participant", h.AddParticipantToEvent)
	s.router.GET("/api/events/{id}/participants", h.GetEventParticipants)
	s.router.DELETE("/api/events/{id}/participants/{member_id}", h.RemoveParticipantFromEvent)
	s.router.POST("/api/events/{id}/attendance", h.MarkAttendance)
	s.router.POST("/api/events/{id}/move-to-table", h.MoveToTable)
	s.router.GET("/api/events/{id}/activities", h.GetSessionActivities)

	// Image upload
	s.router.POST("/api/images/upload", h.UploadImage)
	s.router.GET("/api/uploads/{filename}", h.ServeImage)
	s.router.GET("/api/uploads/collages/{filename}", h.ServeCollage)
}

// ListenAndServe serves on host:port from cfg.Server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Printf("api listening on %s", addr)
	return fasthttp.ListenAndServe(addr, s.Handler())
}

func cors(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		h := "Access-Control-Allow-Headers"
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header.Set(h, "Content-Type, Authorization")
		if string(ctx.Method()) == fasthttp.MethodOptions {
			ctx.SetStatusCode(http.StatusNoContent)
			return
		}
		next(ctx)
	}
}
