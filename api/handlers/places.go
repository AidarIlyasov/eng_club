package handlers

import (
	"database/sql"
	"encoding/json"
	"eng_club/models"

	"github.com/valyala/fasthttp"
)

type placeBody struct {
	Name      string `json:"name"`
	MetroArea string `json:"metro_area"`
	MapURL    string `json:"map_url"`
	ImageURL  string `json:"image_url"`
}

// ListPlaces handles GET /api/places.
func (d *Deps) ListPlaces(ctx *fasthttp.RequestCtx) {
	places, err := d.DB.ListPlaces()
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if places == nil {
		places = []models.Place{}
	}
	WriteJSON(ctx, fasthttp.StatusOK, places)
}

// GetPlace handles GET /api/places/{id}.
func (d *Deps) GetPlace(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	p, err := d.DB.GetPlace(id)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "place not found")
		return
	}
	WriteJSON(ctx, fasthttp.StatusOK, p)
}

// CreatePlace handles POST /api/places.
func (d *Deps) CreatePlace(ctx *fasthttp.RequestCtx) {
	var body placeBody
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		WriteErr(ctx, fasthttp.StatusBadRequest, "name is required")
		return
	}
	p, err := d.DB.CreatePlace(body.Name, body.MetroArea, body.MapURL, body.ImageURL)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(ctx, fasthttp.StatusCreated, p)
}

// UpdatePlace handles PUT /api/places/{id}.
func (d *Deps) UpdatePlace(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var body placeBody
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		WriteErr(ctx, fasthttp.StatusBadRequest, "name is required")
		return
	}
	p, err := d.DB.UpdatePlace(id, body.Name, body.MetroArea, body.MapURL, body.ImageURL)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		WriteErr(ctx, fasthttp.StatusNotFound, "place not found")
		return
	}
	WriteJSON(ctx, fasthttp.StatusOK, p)
}

// DeletePlace handles DELETE /api/places/{id}.
func (d *Deps) DeletePlace(ctx *fasthttp.RequestCtx) {
	id, err := PathID(ctx, "id")
	if err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	err = d.DB.DeletePlace(id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteErr(ctx, fasthttp.StatusNotFound, "place not found")
			return
		}
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
